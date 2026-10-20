package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

// newGenDocsCmd writes the Markdown command reference (docs/reference).
// It is hidden because it's only used by `make docs`.
func newGenDocsCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "gen-docs <dir>",
		Short:  "Generate the Markdown command reference",
		Args:   cobra.ExactArgs(1),
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := genDocs(args[0]); err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "Wrote command reference to %s\n", args[0])
			return nil
		},
	}
}

// genDocs renders the reference from a fresh command tree so output doesn't
// depend on how the current process was invoked.
func genDocs(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	root := newRootCmd()
	root.DisableAutoGenTag = true // no dates, so regenerating is diff-free
	return doc.GenMarkdownTree(root, dir)
}
