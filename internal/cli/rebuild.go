package cli

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/parmar94/dev-workspace-cli/internal/devcontainer"
)

// lookPath is swapped in tests.
var lookPath = exec.LookPath

func newRebuildCmd() *cobra.Command {
	var noCache bool
	cmd := &cobra.Command{
		Use:   "rebuild [dir]",
		Short: "Rebuild and restart a project's dev container",
		Long: `Rebuild a project's dev container with the devcontainer CLI, replacing the
existing container. Uses "devcontainer" from PATH, or "npx @devcontainers/cli".

Named volumes (agent logins) are kept. Credentials passed with dws creds live
in the old container and are gone afterwards; pass them in again.

If the folder is open in VS Code, you can use "Dev Containers: Rebuild
Container" there instead.`,
		Example: `  dws rebuild
  dws rebuild ./api --no-cache`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRebuild(cmd, dirArg(args), noCache)
		},
	}
	cmd.Flags().BoolVar(&noCache, "no-cache", false, "rebuild the image without Docker's build cache")
	return cmd
}

// devcontainerCLI finds the devcontainer CLI, falling back to npx.
func devcontainerCLI() (string, []string, error) {
	if p, err := lookPath("devcontainer"); err == nil {
		return p, nil, nil
	}
	if p, err := lookPath("npx"); err == nil {
		return p, []string{"-y", "@devcontainers/cli"}, nil
	}
	return "", nil, errors.New("the devcontainer CLI is not installed: run `npm install -g @devcontainers/cli` (needs Node.js), or rebuild from VS Code")
}

func runRebuild(cmd *cobra.Command, dir string, noCache bool) error {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	if _, err := os.Stat(filepath.Join(abs, devcontainer.Dir, "devcontainer.json")); err != nil {
		return fmt.Errorf("no %s/devcontainer.json in %s; run `dws init` first", devcontainer.Dir, dir)
	}
	name, args, err := devcontainerCLI()
	if err != nil {
		return err
	}
	args = append(args, "up", "--workspace-folder", abs, "--remove-existing-container")
	if noCache {
		args = append(args, "--build-no-cache")
	}

	errOut := cmd.ErrOrStderr()
	fmt.Fprintf(errOut, "Rebuilding the dev container for %s...\n", dir)
	c := exec.CommandContext(cmd.Context(), name, args...)
	c.Stdout, c.Stderr = cmd.OutOrStdout(), errOut
	if err := c.Run(); err != nil {
		return fmt.Errorf("rebuild failed: %w", err)
	}
	fmt.Fprintln(errOut, "\nRebuilt. If the folder is open in VS Code, run \"Dev Containers: Reopen in Container\" to attach to the new container.")
	fmt.Fprintln(errOut, "Credentials passed with `dws creds` were in the old container; pass them in again if needed.")
	return nil
}
