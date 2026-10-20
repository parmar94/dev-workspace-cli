package cli

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/parmar94/dev-workspace-cli/internal/devcontainer"
)

type updateOpts struct {
	selectionFlags
	addLangs, rmLangs   []string
	addAgents, rmAgents []string
	addTools, rmTools   []string
	yes                 bool
	dryRun              bool
	force               bool
	rebuild             bool
	noCache             bool
}

// selectionFlagNames are the flags that change the selection; when none is
// given, update prompts (on a terminal) instead.
var selectionFlagNames = []string{
	"name", "base-image", "lang", "agent", "tool", "ssh-agent", "ssh-mode",
	"git-config", "agent-guard", "agent-sandbox", "port",
	"add-lang", "remove-lang", "add-agent", "remove-agent", "add-tool", "remove-tool",
}

func newUpdateCmd(a *app) *cobra.Command {
	o := &updateOpts{}
	cmd := &cobra.Command{
		Use:   "update [dir]",
		Short: "Change a project's dev container and regenerate it",
		Long: `Change the languages, agents, tools or options of a project generated with
dws init, and regenerate .devcontainer/ from its dws.yaml.

Start from the recorded selection and adjust it with --add-*/--remove-*, or
replace a whole list with --lang/--agent/--tool. With no changes on a
terminal, you're prompted with the current selection. With no changes and
--yes, files are regenerated with the current registry (useful after
upgrading dws or editing your registry).

Files you edited by hand since dws generated them are not overwritten unless
you pass --force.`,
		Example: `  dws update --add-tool ripgrep,terraform --remove-agent codex
  dws update --add-lang python --dry-run
  dws update --yes --rebuild          # regenerate with the current registry, then rebuild`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.runUpdate(cmd, dirArg(args), o)
		},
	}
	f := cmd.Flags()
	o.register(f)
	f.StringSliceVar(&o.addLangs, "add-lang", nil, "add languages")
	f.StringSliceVar(&o.rmLangs, "remove-lang", nil, "remove languages")
	f.StringSliceVar(&o.addAgents, "add-agent", nil, "add agents")
	f.StringSliceVar(&o.rmAgents, "remove-agent", nil, "remove agents")
	f.StringSliceVar(&o.addTools, "add-tool", nil, "add tools")
	f.StringSliceVar(&o.rmTools, "remove-tool", nil, "remove tools")
	f.BoolVarP(&o.yes, "yes", "y", false, "don't prompt")
	f.BoolVar(&o.dryRun, "dry-run", false, "show what would change without writing")
	f.BoolVar(&o.force, "force", false, "overwrite generated files you edited by hand")
	f.BoolVar(&o.rebuild, "rebuild", false, "rebuild the container after updating")
	f.BoolVar(&o.noCache, "no-cache", false, "with --rebuild, skip Docker's build cache")
	return cmd
}

func (a *app) runUpdate(cmd *cobra.Command, dir string, o *updateOpts) error {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	old, err := devcontainer.LoadState(abs)
	if errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("no %s in %s; run `dws init` first", devcontainer.StateFile, dir)
	}
	if err != nil {
		return err
	}
	cfg, _, err := a.loadConfig()
	if err != nil {
		return err
	}
	reg, err := cfg.EffectiveRegistry()
	if err != nil {
		return err
	}

	sel := old
	sel.Languages = slices.Clone(old.Languages)
	sel.Agents = slices.Clone(old.Agents)
	sel.Tools = slices.Clone(old.Tools)
	sel.Ports = slices.Clone(old.Ports)
	sel.Generated = nil
	changed := cmd.Flags().Changed
	if err := o.apply(changed, &sel); err != nil {
		return err
	}
	sel.Languages = addRemove(sel.Languages, o.addLangs, o.rmLangs)
	sel.Agents = addRemove(sel.Agents, o.addAgents, o.rmAgents)
	sel.Tools = addRemove(sel.Tools, o.addTools, o.rmTools)

	anyChange := slices.ContainsFunc(selectionFlagNames, changed)
	if !anyChange && !o.yes && isInteractive() {
		if err := promptSelection(&sel, cfg, reg, changed); err != nil {
			return err
		}
	}
	sel.Languages, sel.Agents, sel.Tools = uniq(sel.Languages), uniq(sel.Agents), uniq(sel.Tools)

	files, added, err := render(sel, reg)
	if err != nil {
		return err
	}

	errOut := cmd.ErrOrStderr()
	changes := selectionChanges(old, sel)
	statuses := fileStatuses(abs, old, files)
	if len(changes) == 0 && len(statuses) == 0 {
		fmt.Fprintln(errOut, "Already up to date.")
		if o.rebuild && !o.dryRun {
			return runRebuild(cmd, dir, o.noCache)
		}
		return nil
	}
	if len(added) > 0 {
		fmt.Fprintf(errOut, "Also adding required: %s\n", strings.Join(added, ", "))
	}
	warnSelection(errOut, sel)
	for _, c := range changes {
		fmt.Fprintln(errOut, c)
	}
	for _, s := range statuses {
		fmt.Fprintln(errOut, s)
	}

	if edited := old.Modified(abs); len(edited) > 0 && !o.force {
		return fmt.Errorf("these generated files were edited by hand since dws wrote them: %s\nmove your changes elsewhere, or pass --force to overwrite them", strings.Join(edited, ", "))
	}
	if o.dryRun {
		printFiles(cmd.OutOrStdout(), changedFiles(abs, files))
		return nil
	}
	if err := devcontainer.Write(abs, files, true); err != nil {
		return err
	}
	fmt.Fprintf(errOut, "Updated %s.\n", filepath.Join(dir, devcontainer.Dir))
	if o.rebuild {
		return runRebuild(cmd, dir, o.noCache)
	}
	fmt.Fprintf(errOut, "Rebuild to apply: dws rebuild %s\n", dir)
	return nil
}

// addRemove appends add (skipping existing) and then drops rm.
func addRemove(list, add, rm []string) []string {
	for _, a := range add {
		if !slices.Contains(list, a) {
			list = append(list, a)
		}
	}
	return slices.DeleteFunc(list, func(s string) bool { return slices.Contains(rm, s) })
}

// selectionChanges describes how sel differs from old, one line each.
func selectionChanges(old, sel selection) []string {
	var out []string
	diffList := func(kind string, a, b []string) {
		for _, x := range b {
			if !slices.Contains(a, x) {
				out = append(out, fmt.Sprintf("  + %s %s", kind, x))
			}
		}
		for _, x := range a {
			if !slices.Contains(b, x) {
				out = append(out, fmt.Sprintf("  - %s %s", kind, x))
			}
		}
	}
	diffList("language", old.Languages, sel.Languages)
	diffList("agent", old.Agents, sel.Agents)
	diffList("tool", old.Tools, sel.Tools)
	field := func(name string, a, b any) {
		if fmt.Sprint(a) != fmt.Sprint(b) {
			out = append(out, fmt.Sprintf("  ~ %s: %v -> %v", name, a, b))
		}
	}
	field("name", old.Name, sel.Name)
	field("base image", old.BaseImage, sel.BaseImage)
	field("ssh agent", orNone(string(old.SSH)), orNone(string(sel.SSH)))
	field("git config", old.GitConfig, sel.GitConfig)
	field("agent guard", old.AgentGuard, sel.AgentGuard)
	field("agent sandbox", old.AgentSandbox, sel.AgentSandbox)
	field("ports", old.Ports, sel.Ports)
	return out
}

func orNone(s string) string {
	if s == "" {
		return "off"
	}
	return s
}

// fileStatuses lists generated files that are new, changed or no longer
// generated. dws.yaml is skipped: it changes whenever anything else does.
func fileStatuses(root string, old selection, files []devcontainer.File) []string {
	var out []string
	seen := map[string]bool{}
	for _, f := range files {
		seen[f.Path] = true
		if f.Path == devcontainer.StateFile {
			continue
		}
		cur, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(f.Path)))
		switch {
		case err != nil:
			out = append(out, "  new      "+f.Path)
		case !bytes.Equal(cur, f.Data):
			out = append(out, "  changed  "+f.Path)
		}
	}
	for _, p := range sortedKeys(old.Generated) {
		if !seen[p] {
			out = append(out, "  removed  "+p)
		}
	}
	return out
}

// changedFiles returns the files whose content differs from what's on disk.
func changedFiles(root string, files []devcontainer.File) []devcontainer.File {
	var out []devcontainer.File
	for _, f := range files {
		cur, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(f.Path)))
		if err != nil || !bytes.Equal(cur, f.Data) {
			out = append(out, f)
		}
	}
	return out
}
