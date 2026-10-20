package cli

import (
	"errors"
	"os"
	"slices"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/x/term"

	"github.com/parmar94/dev-workspace-cli/internal/config"
	"github.com/parmar94/dev-workspace-cli/internal/devcontainer"
	"github.com/parmar94/dev-workspace-cli/internal/registry"
)

var errAborted = errors.New("aborted")

// isInteractive reports whether prompts can be shown.
func isInteractive() bool {
	return term.IsTerminal(os.Stdin.Fd()) && term.IsTerminal(os.Stdout.Fd())
}

// runForm runs a form. Set ACCESSIBLE=1 for plain line-based prompts
// (screen readers, dumb terminals).
func runForm(groups ...*huh.Group) error {
	err := huh.NewForm(groups...).WithAccessible(os.Getenv("ACCESSIBLE") != "").Run()
	if errors.Is(err, huh.ErrUserAborted) {
		return errAborted
	}
	return err
}

// promptTemplate asks which template to start from. "" means none.
func promptTemplate(tmpls map[string]registry.Template) (string, error) {
	if len(tmpls) == 0 {
		return "", nil
	}
	opts := []huh.Option[string]{huh.NewOption("None (use my defaults)", "")}
	for _, name := range sortedKeys(tmpls) {
		label := name
		if d := tmpls[name].Description; d != "" {
			label += " - " + d
		}
		opts = append(opts, huh.NewOption(label, name))
	}
	var choice string
	err := runForm(huh.NewGroup(
		huh.NewSelect[string]().
			Title("Start from a template?").
			Options(opts...).
			Value(&choice),
	))
	return choice, err
}

// promptSelection asks for every value that wasn't set by a flag,
// pre-filled with the current selection.
func promptSelection(sel *selection, cfg *config.Config, reg *registry.Registry, changed func(string) bool) error {
	var fields []huh.Field
	lists := []struct {
		flag  string
		kind  registry.Kind
		title string
		value *[]string
	}{
		{"lang", registry.KindLanguage, "Languages", &sel.Languages},
		{"agent", registry.KindAgent, "AI agents", &sel.Agents},
		{"tool", registry.KindTool, "Tools", &sel.Tools},
	}
	for _, l := range lists {
		if changed(l.flag) {
			continue
		}
		opts := entryOptions(reg, cfg, l.kind, *l.value)
		fields = append(fields, huh.NewMultiSelect[string]().
			Title(l.title).
			Description("space to toggle, / to filter, ★ = favorite").
			Options(opts...).
			Filterable(true).
			Height(min(len(opts), 10)+2).
			Value(l.value))
	}

	ssh := sel.SSH != devcontainer.SSHNone
	askSSH := !changed("ssh-agent") && !changed("ssh-mode")
	if askSSH {
		fields = append(fields, huh.NewConfirm().
			Title("Forward your SSH agent into the container?").
			Value(&ssh))
	}
	gitCfg := sel.GitConfig
	if !changed("git-config") {
		fields = append(fields, huh.NewConfirm().
			Title("Mount your host ~/.gitconfig?").
			Description("VS Code copies it already; useful with the devcontainer CLI or other editors.").
			Value(&gitCfg))
	}

	if len(fields) == 0 {
		return nil
	}
	groups := make([]*huh.Group, len(fields))
	for i, f := range fields {
		groups[i] = huh.NewGroup(f)
	}
	if err := runForm(groups...); err != nil {
		return err
	}

	if askSSH {
		switch {
		case !ssh:
			sel.SSH = devcontainer.SSHNone
		case sel.SSH == devcontainer.SSHNone:
			sel.SSH = hostSSHMode()
		}
	}
	sel.GitConfig = gitCfg
	return nil
}

// entryOptions lists registry entries of kind with favorites first, and
// pre-selects the ones in current.
func entryOptions(reg *registry.Registry, cfg *config.Config, kind registry.Kind, current []string) []huh.Option[string] {
	names := reg.Names(kind)
	slices.SortStableFunc(names, func(a, b string) int {
		fa, fb := cfg.IsFavorite(a), cfg.IsFavorite(b)
		switch {
		case fa && !fb:
			return -1
		case fb && !fa:
			return 1
		}
		return 0
	})
	opts := make([]huh.Option[string], len(names))
	for i, n := range names {
		e, _ := reg.Lookup(n)
		var label strings.Builder
		if cfg.IsFavorite(n) {
			label.WriteString("★ ")
		}
		label.WriteString(n)
		if e.Description != "" {
			label.WriteString(" - " + e.Description)
		}
		opts[i] = huh.NewOption(label.String(), n).Selected(slices.Contains(current, n))
	}
	return opts
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}
