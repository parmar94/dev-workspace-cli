// Package cli wires up the dws command tree.
package cli

import (
	"github.com/spf13/cobra"

	"github.com/parmar94/dev-workspace-cli/internal/config"
)

// Set at build time via -ldflags.
var (
	version = "dev"
	commit  = "none"
)

// app holds state shared by all commands.
type app struct {
	configPath string
}

func (a *app) path() (string, error) {
	if a.configPath != "" {
		return a.configPath, nil
	}
	return config.Path()
}

func (a *app) loadConfig() (*config.Config, string, error) {
	p, err := a.path()
	if err != nil {
		return nil, "", err
	}
	c, err := config.Load(p)
	return c, p, err
}

func newRootCmd() *cobra.Command {
	a := &app{}
	root := &cobra.Command{
		Use:          "dws",
		Short:        "Set up and manage project workspaces in dev containers",
		SilenceUsage: true,
	}
	root.PersistentFlags().StringVar(&a.configPath, "config", "", "global config file (default ~/.config/dws/config.yaml, or $DWS_CONFIG)")

	root.AddCommand(
		newVersionCmd(),
		newConfigCmd(a),
		newInitCmd(a),
	)

	return root
}

// Execute runs the root command.
func Execute() error {
	return newRootCmd().Execute()
}
