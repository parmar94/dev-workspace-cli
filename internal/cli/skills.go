package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/parmar94/dev-workspace-cli/internal/assets"
)

// skillDirs maps agent -> scope -> skills directory. Project paths are
// relative to the project dir, user paths to $HOME.
var skillDirs = map[string]map[string]string{
	"claude": {"user": ".claude/skills", "project": ".claude/skills"},
	"codex":  {"user": ".agents/skills", "project": ".agents/skills"},
}

func newSkillsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "skills",
		Aliases: []string{"skill"},
		Short:   "Install agent skills that teach Claude Code and Codex to use dws",
	}
	cmd.AddCommand(newSkillsListCmd(), newSkillsInstallCmd())
	return cmd
}

func skillNames() ([]string, error) {
	entries, err := fs.ReadDir(assets.Skills, "skills")
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	return names, nil
}

// skillDescription reads the description line from a skill's frontmatter.
func skillDescription(name string) string {
	f, err := assets.Skills.Open(path.Join("skills", name, "SKILL.md"))
	if err != nil {
		return ""
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		if d, ok := strings.CutPrefix(s.Text(), "description: "); ok {
			return d
		}
	}
	return ""
}

func newSkillsListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List the bundled skills",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			names, err := skillNames()
			if err != nil {
				return err
			}
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tDESCRIPTION")
			for _, n := range names {
				d := skillDescription(n)
				if i := strings.Index(d, ". "); i > 0 {
					d = d[:i+1]
				}
				fmt.Fprintf(w, "%s\t%s\n", n, d)
			}
			return w.Flush()
		},
	}
}

func newSkillsInstallCmd() *cobra.Command {
	var agents []string
	var scope, dir string
	var force bool
	cmd := &cobra.Command{
		Use:   "install [skill]...",
		Short: "Install bundled skills for Claude Code and/or Codex",
		Long: `Install the bundled skills (default: all) so Claude Code and Codex can set up
dev containers, manage dws config and handle credentials for you.

  claude  user: ~/.claude/skills    project: <dir>/.claude/skills
  codex   user: ~/.agents/skills    project: <dir>/.agents/skills

Re-run with --force after upgrading dws to update installed skills.`,
		Example: `  dws skills install
  dws skills install --agent claude --scope project --dir .
  dws skills install dws-credentials --force`,
		RunE: func(cmd *cobra.Command, args []string) error {
			all, err := skillNames()
			if err != nil {
				return err
			}
			names := all
			if len(args) > 0 {
				for _, a := range args {
					if !slices.Contains(all, a) {
						return fmt.Errorf("unknown skill %q (available: %s)", a, strings.Join(all, ", "))
					}
				}
				names = args
			}

			var base string
			switch scope {
			case "user":
				if base, err = os.UserHomeDir(); err != nil {
					return err
				}
			case "project":
				if base, err = filepath.Abs(dir); err != nil {
					return err
				}
			default:
				return fmt.Errorf("--scope must be user or project, got %q", scope)
			}

			for _, agent := range uniq(agents) {
				scopes, ok := skillDirs[agent]
				if !ok {
					return fmt.Errorf("unknown agent %q (want claude or codex)", agent)
				}
				root := filepath.Join(base, filepath.FromSlash(scopes[scope]))
				for _, n := range names {
					dst := filepath.Join(root, n)
					if err := installSkill(n, dst, force); err != nil {
						return err
					}
					fmt.Fprintf(cmd.ErrOrStderr(), "Installed %s for %s at %s\n", n, agent, dst)
				}
			}
			return nil
		},
	}
	f := cmd.Flags()
	f.StringSliceVarP(&agents, "agent", "a", []string{"claude", "codex"}, "agents to install for: claude, codex")
	f.StringVar(&scope, "scope", "user", "user (all projects) or project")
	f.StringVar(&dir, "dir", ".", "project directory for --scope project")
	f.BoolVar(&force, "force", false, "replace skills that are already installed")
	return cmd
}

// installSkill copies the embedded skill name into dst.
func installSkill(name, dst string, force bool) error {
	if _, err := os.Stat(dst); err == nil {
		if !force {
			return fmt.Errorf("%s already exists (use --force to replace it)", dst)
		}
		if err := os.RemoveAll(dst); err != nil {
			return err
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	src := path.Join("skills", name)
	return fs.WalkDir(assets.Skills, src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		target := filepath.Join(dst, filepath.FromSlash(strings.TrimPrefix(p, src)))
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := fs.ReadFile(assets.Skills, p)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
}
