// Package cli wires up the dws command tree.
package cli

import (
	"github.com/spf13/cobra"
)

// Set at build time via -ldflags.
var (
	version = "dev"
	commit  = "none"
)

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:          "dws",
		Short:        "Set up and manage project workspaces in dev containers",
		SilenceUsage: true,
	}

	root.AddCommand(newVersionCmd())

	return root
}

// Execute runs the root command.
func Execute() error {
	return newRootCmd().Execute()
}
