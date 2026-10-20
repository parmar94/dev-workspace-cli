package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the dws version",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, _ []string) {
			fmt.Fprintf(cmd.OutOrStdout(), "dws %s (%s)\n", version, commit)
		},
	}
}
