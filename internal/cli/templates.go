package cli

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"go.yaml.in/yaml/v3"

	"github.com/parmar94/dev-workspace-cli/internal/registry"
)

func newTemplatesCmd(a *app) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "templates",
		Aliases: []string{"template"},
		Short:   "List and show templates (add your own under `templates:` in the config)",
	}
	cmd.AddCommand(
		&cobra.Command{
			Use:   "list",
			Short: "List templates",
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, _ []string) error {
				cfg, _, err := a.loadConfig()
				if err != nil {
					return err
				}
				tmpls, err := cfg.EffectiveTemplates()
				if err != nil {
					return err
				}
				w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
				fmt.Fprintln(w, "NAME\tSOURCE\tDESCRIPTION")
				for _, n := range sortedKeys(tmpls) {
					src := "built-in"
					if _, ok := cfg.Templates[n]; ok {
						src = "custom"
					}
					fmt.Fprintf(w, "%s\t%s\t%s\n", n, src, tmpls[n].Description)
				}
				return w.Flush()
			},
		},
		&cobra.Command{
			Use:   "show <name>",
			Short: "Print a template",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				cfg, _, err := a.loadConfig()
				if err != nil {
					return err
				}
				tmpls, err := cfg.EffectiveTemplates()
				if err != nil {
					return err
				}
				t, ok := tmpls[args[0]]
				if !ok {
					return fmt.Errorf("unknown template %q", args[0])
				}
				enc := yaml.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent(2)
				return enc.Encode(map[string]registry.Template{args[0]: t})
			},
		},
	)
	return cmd
}
