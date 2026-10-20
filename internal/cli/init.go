package cli

import (
	"fmt"
	"io"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/parmar94/dev-workspace-cli/internal/config"
	"github.com/parmar94/dev-workspace-cli/internal/devcontainer"
	"github.com/parmar94/dev-workspace-cli/internal/registry"
)

type initOpts struct {
	selectionFlags
	template string
	yes      bool
	dryRun   bool
	force    bool
}

// selection is the merged result of config defaults, template, flags and
// prompts. It maps directly onto devcontainer.State.
type selection = devcontainer.State

func newInitCmd(a *app) *cobra.Command {
	o := &initOpts{}
	cmd := &cobra.Command{
		Use:   "init [dir]",
		Short: "Generate a .devcontainer for a project",
		Long: `Generate a .devcontainer for a project (default: current directory).

Values come from, lowest to highest priority: global config defaults, the
--template, then flags. A template or a list flag (--lang, --agent, --tool)
replaces the list below it rather than adding to it.`,
		Example: `  dws init
  dws init --template go-web --yes
  dws init ./api --lang go --agent claude,codex --tool gh,jq --ssh-agent --yes --dry-run`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := "."
			if len(args) == 1 {
				dir = args[0]
			}
			return a.runInit(cmd, dir, o)
		},
	}
	f := cmd.Flags()
	f.StringVarP(&o.template, "template", "t", "", "start from a `name`d template (see dws templates list)")
	o.register(f)
	f.BoolVarP(&o.yes, "yes", "y", false, "don't prompt; use defaults, template and flags")
	f.BoolVar(&o.dryRun, "dry-run", false, "print the files instead of writing them")
	f.BoolVar(&o.force, "force", false, "overwrite an existing .devcontainer")
	return cmd
}

func (a *app) runInit(cmd *cobra.Command, dir string, o *initOpts) error {
	cfg, _, err := a.loadConfig()
	if err != nil {
		return err
	}
	reg, err := cfg.EffectiveRegistry()
	if err != nil {
		return err
	}
	tmpls, err := cfg.EffectiveTemplates()
	if err != nil {
		return err
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return err
	}

	interactive := !o.yes && isInteractive()
	if interactive && !cmd.Flags().Changed("template") {
		if o.template, err = promptTemplate(tmpls); err != nil {
			return err
		}
	}

	sel, err := buildSelection(cmd, cfg, tmpls, o, abs)
	if err != nil {
		return err
	}

	if interactive {
		if err := promptSelection(&sel, cfg, reg, cmd.Flags().Changed); err != nil {
			return err
		}
	}

	files, added, err := render(sel, reg)
	if err != nil {
		return err
	}

	out, errOut := cmd.OutOrStdout(), cmd.ErrOrStderr()
	if len(added) > 0 {
		fmt.Fprintf(errOut, "Also adding required: %s\n", strings.Join(added, ", "))
	}
	warnSelection(errOut, sel)
	if o.dryRun {
		printFiles(out, files)
		return nil
	}
	if err := devcontainer.Write(abs, files, o.force); err != nil {
		return err
	}
	for _, f := range files {
		fmt.Fprintf(errOut, "Wrote %s\n", filepath.Join(dir, f.Path))
	}
	fmt.Fprintf(errOut, "\nNext: open the folder in VS Code and run \"Reopen in Container\", or\n  devcontainer up --workspace-folder %s\n", dir)
	return nil
}

// buildSelection merges config defaults < template < flags.
func buildSelection(cmd *cobra.Command, cfg *config.Config, tmpls map[string]registry.Template, o *initOpts, abs string) (selection, error) {
	d := cfg.Defaults
	sel := selection{
		Version:   1,
		Name:      filepath.Base(abs),
		BaseImage: cfg.BaseImage(),
		User:      cfg.RemoteUser(),
		Languages: d.Languages,
		Agents:    d.Agents,
		Tools:     d.Tools,
		// On by default; only matters when an agent is selected.
		AgentGuard:   true,
		AgentSandbox: d.AgentSandbox,
	}
	if d.SSHAgent {
		sel.SSH = hostSSHMode()
	}

	if o.template != "" {
		t, ok := tmpls[o.template]
		if !ok {
			return sel, fmt.Errorf("unknown template %q (see `dws templates list`)", o.template)
		}
		sel.Template = o.template
		sel.Languages, sel.Agents, sel.Tools = t.Languages, t.Agents, t.Tools
		if t.BaseImage != "" {
			sel.BaseImage = t.BaseImage
		}
		if t.SSHAgent != nil {
			sel.SSH = devcontainer.SSHNone
			if *t.SSHAgent {
				sel.SSH = hostSSHMode()
			}
		}
	}

	if err := o.apply(cmd.Flags().Changed, &sel); err != nil {
		return sel, err
	}
	return sel, nil
}

// render checks the selection against the registry and produces the files.
func render(sel selection, reg *registry.Registry) ([]devcontainer.File, []string, error) {
	sel.Languages, sel.Agents, sel.Tools = uniq(sel.Languages), uniq(sel.Agents), uniq(sel.Tools)
	for kind, names := range map[registry.Kind][]string{
		registry.KindLanguage: sel.Languages,
		registry.KindAgent:    sel.Agents,
		registry.KindTool:     sel.Tools,
	} {
		if err := checkKind(reg, kind, names); err != nil {
			return nil, nil, err
		}
	}

	entries, added, err := reg.Resolve(slices.Concat(sel.Languages, sel.Agents, sel.Tools))
	if err != nil {
		return nil, nil, err
	}
	files, err := devcontainer.Render(devcontainer.Spec{
		Name:         sel.Name,
		Image:        sel.BaseImage,
		RemoteUser:   sel.User,
		Entries:      entries,
		SSH:          sel.SSH,
		GitConfig:    sel.GitConfig,
		Ports:        sel.Ports,
		AgentGuard:   sel.AgentGuard,
		AgentSandbox: sel.AgentSandbox,
	})
	if err != nil {
		return nil, nil, err
	}
	sel.Record(files)
	state, err := sel.File()
	if err != nil {
		return nil, nil, err
	}
	return append(files, state), added, nil
}

// warnSelection prints notes about options that have no effect.
func warnSelection(w io.Writer, sel selection) {
	if sel.AgentSandbox && !slices.Contains(sel.Agents, "claude") {
		fmt.Fprintln(w, "Note: --agent-sandbox only applies to Claude Code, which isn't selected.")
	}
}

func checkKind(reg *registry.Registry, kind registry.Kind, names []string) error {
	for _, n := range names {
		e, ok := reg.Lookup(n)
		if !ok {
			return fmt.Errorf("unknown %s %q (available: %s)", kind, n, strings.Join(reg.Names(kind), ", "))
		}
		if e.Kind != kind {
			return fmt.Errorf("%q is a %s, not a %s", n, e.Kind, kind)
		}
	}
	return nil
}

func printFiles(w io.Writer, files []devcontainer.File) {
	for i, f := range files {
		if i > 0 {
			fmt.Fprintln(w)
		}
		fmt.Fprintf(w, "# --- %s\n", f.Path)
		w.Write(f.Data)
	}
}

func hostSSHMode() devcontainer.SSHMode {
	if runtime.GOOS == "darwin" {
		return devcontainer.SSHMac
	}
	return devcontainer.SSHLinux
}

func uniq(s []string) []string {
	var out []string
	for _, v := range s {
		if v = strings.TrimSpace(v); v != "" && !slices.Contains(out, v) {
			out = append(out, v)
		}
	}
	return out
}
