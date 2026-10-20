package cli

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/parmar94/dev-workspace-cli/internal/config"
	"github.com/parmar94/dev-workspace-cli/internal/registry"
)

const saveNote = "Note: saving rewrites the config file, so comments in it are not kept."

func newToolsCmd(a *app) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "tools",
		Aliases: []string{"tool"},
		Short:   "List and manage the languages, agents and tools in your registry",
	}
	cmd.AddCommand(newToolsListCmd(a), newToolsAddCmd(a), newToolsRemoveCmd(a))
	return cmd
}

func newToolsListCmd(a *app) *cobra.Command {
	var kind string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List registry entries",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, _, err := a.loadConfig()
			if err != nil {
				return err
			}
			reg, err := cfg.EffectiveRegistry()
			if err != nil {
				return err
			}
			kinds := registry.Kinds
			if kind != "" {
				k, err := parseKind(kind)
				if err != nil {
					return err
				}
				kinds = []registry.Kind{k}
			}
			builtin, err := registry.Builtin()
			if err != nil {
				return err
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "KIND\tNAME\tINSTALL\tSOURCE\tDESCRIPTION")
			for _, k := range kinds {
				for _, n := range reg.Names(k) {
					e, _ := reg.Lookup(n)
					name := n
					if cfg.IsFavorite(n) {
						name = "★ " + n
					}
					install := "feature"
					if e.Script != "" {
						install = "script"
					}
					if len(e.Requires) > 0 {
						install += " (needs " + strings.Join(e.Requires, ",") + ")"
					}
					fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", k, name, install, source(cfg, builtin, n), e.Description)
				}
			}
			return w.Flush()
		},
	}
	cmd.Flags().StringVarP(&kind, "kind", "k", "", "only show one kind: language, agent or tool")
	return cmd
}

func source(cfg *config.Config, builtin *registry.Registry, name string) string {
	_, custom := cfg.Registry.Lookup(name)
	_, isBuiltin := builtin.Lookup(name)
	switch {
	case custom && isBuiltin:
		return "override"
	case custom:
		return "custom"
	default:
		return "built-in"
	}
}

func newToolsAddCmd(a *app) *cobra.Command {
	var (
		kind, feature, script, scriptFile, desc string
		options, env                            []string
		requires, mounts                        []string
		guard                                   registry.Guard
		favorite                                bool
	)
	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Add an entry to the registry, or override a built-in one",
		Long: `Add an entry to the registry, or override a built-in one.

Give either --feature (a dev container feature reference) or --script /
--script-file (shell run as the remote user after the container is created).
In --mount, {{home}} expands to the remote user's home directory.

` + saveNote,
		Example: `  dws tools add httpie --script "sudo apt-get update && sudo apt-get install -y httpie"
  dws tools add deno --kind language --feature ghcr.io/devcontainers-extra/features/deno:1
  dws tools add gemini --kind agent --requires node --script "npm i -g @google/gemini-cli" --favorite`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			k, err := parseKind(kind)
			if err != nil {
				return err
			}
			if scriptFile != "" {
				if script != "" {
					return fmt.Errorf("use only one of --script and --script-file")
				}
				data, err := os.ReadFile(scriptFile)
				if err != nil {
					return err
				}
				script = string(data)
			}
			item := registry.Item{
				Description: desc,
				Feature:     feature,
				Script:      script,
				Requires:    requires,
				Mounts:      mounts,
				Guard:       guard,
			}
			if item.Options, err = parseKV[any](options, "--option"); err != nil {
				return err
			}
			if item.Env, err = parseKV[string](env, "--env"); err != nil {
				return err
			}

			cfg, path, err := a.loadConfig()
			if err != nil {
				return err
			}
			cfg.Registry.Remove(name)
			cfg.Registry.Section(k)[name] = item
			cfg.Removed = slices.DeleteFunc(cfg.Removed, func(s string) bool { return s == name })
			if favorite && !cfg.IsFavorite(name) {
				cfg.Favorites = append(cfg.Favorites, name)
			}
			if _, err := cfg.EffectiveRegistry(); err != nil {
				return err
			}
			if err := cfg.Save(path); err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "Added %s %q to %s\n", k, name, path)
			return nil
		},
	}
	f := cmd.Flags()
	f.StringVarP(&kind, "kind", "k", "tool", "language, agent or tool")
	f.StringVar(&feature, "feature", "", "dev container feature reference")
	f.StringArrayVar(&options, "option", nil, "feature option key=value (repeatable)")
	f.StringVar(&script, "script", "", "install script")
	f.StringVar(&scriptFile, "script-file", "", "read the install script from a file")
	f.StringVarP(&desc, "description", "d", "", "short description")
	f.StringSliceVar(&requires, "requires", nil, "entries this one needs, comma-separated")
	f.StringArrayVar(&mounts, "mount", nil, "mount spec (repeatable)")
	f.StringArrayVar(&env, "env", nil, "container env KEY=value (repeatable)")
	f.StringArrayVar(&guard.Paths, "guard-path", nil, "credential path glob agents may not read, e.g. ~/.mytool/** (repeatable)")
	f.StringArrayVar(&guard.Commands, "guard-cmd", nil, "command prefix agents may not run (repeatable)")
	f.StringArrayVar(&guard.Env, "guard-env", nil, "env var name or glob hidden from agent shells (repeatable)")
	f.BoolVar(&favorite, "favorite", false, "also add to favorites")
	return cmd
}

func newToolsRemoveCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:     "remove <name>...",
		Aliases: []string{"rm"},
		Short:   "Remove entries (built-in ones are hidden)",
		Long:    "Remove entries from the registry. Built-in entries are hidden via `removed` in the config.\n\n" + saveNote,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, path, err := a.loadConfig()
			if err != nil {
				return err
			}
			reg, err := cfg.EffectiveRegistry()
			if err != nil {
				return err
			}
			builtin, err := registry.Builtin()
			if err != nil {
				return err
			}
			for _, n := range args {
				if _, ok := reg.Lookup(n); !ok {
					return fmt.Errorf("%q is not in the registry", n)
				}
				cfg.Registry.Remove(n)
				if _, ok := builtin.Lookup(n); ok && !slices.Contains(cfg.Removed, n) {
					cfg.Removed = append(cfg.Removed, n)
				}
				drop := func(s string) bool { return s == n }
				cfg.Favorites = slices.DeleteFunc(cfg.Favorites, drop)
				d := &cfg.Defaults
				d.Languages = slices.DeleteFunc(d.Languages, drop)
				d.Agents = slices.DeleteFunc(d.Agents, drop)
				d.Tools = slices.DeleteFunc(d.Tools, drop)
			}
			if _, err := cfg.EffectiveRegistry(); err != nil {
				return fmt.Errorf("can't remove: %w", err)
			}
			if err := cfg.Save(path); err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "Removed %s\n", strings.Join(args, ", "))
			return nil
		},
	}
}

func parseKind(s string) (registry.Kind, error) {
	k := registry.Kind(strings.TrimSuffix(s, "s"))
	if !slices.Contains(registry.Kinds, k) {
		return "", fmt.Errorf("unknown kind %q (want language, agent or tool)", s)
	}
	return k, nil
}

func parseKV[V any](pairs []string, flag string) (map[string]V, error) {
	if len(pairs) == 0 {
		return nil, nil
	}
	out := make(map[string]V, len(pairs))
	for _, p := range pairs {
		k, v, ok := strings.Cut(p, "=")
		if !ok || k == "" {
			return nil, fmt.Errorf("%s %q: want key=value", flag, p)
		}
		out[k] = any(v).(V)
	}
	return out, nil
}
