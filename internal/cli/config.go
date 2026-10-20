package cli

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/parmar94/dev-workspace-cli/internal/config"
)

func newConfigCmd(a *app) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage the global config",
	}
	cmd.AddCommand(
		newConfigInitCmd(a),
		newConfigPathCmd(a),
		newConfigShowCmd(a),
		newConfigEditCmd(a),
	)
	return cmd
}

func newConfigInitCmd(a *app) *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Write a starter global config",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			p, err := a.path()
			if err != nil {
				return err
			}
			if err := config.WriteStarter(p, force); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Wrote %s\n", p)
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "overwrite an existing config")
	return cmd
}

func newConfigPathCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Print the global config path",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			p, err := a.path()
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), p)
			return nil
		},
	}
}

func newConfigShowCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Print the global config",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			p, err := a.path()
			if err != nil {
				return err
			}
			data, err := os.ReadFile(p)
			if errors.Is(err, fs.ErrNotExist) {
				fmt.Fprintf(cmd.ErrOrStderr(), "No config at %s; using built-in defaults. Run `dws config init` to create one.\n", p)
				return nil
			}
			if err != nil {
				return err
			}
			_, err = cmd.OutOrStdout().Write(data)
			return err
		},
	}
}

func newConfigEditCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "edit",
		Short: "Open the global config in $EDITOR (creating it if needed)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			p, err := a.path()
			if err != nil {
				return err
			}
			if _, err := os.Stat(p); errors.Is(err, fs.ErrNotExist) {
				if err := config.WriteStarter(p, false); err != nil {
					return err
				}
			}
			editor := os.Getenv("VISUAL")
			if editor == "" {
				editor = os.Getenv("EDITOR")
			}
			if editor == "" {
				editor = "vi"
			}
			// Run through the shell so editors with arguments ("code -w") work.
			c := exec.Command("sh", "-c", editor+` "$1"`, "sh", p)
			c.Stdin, c.Stdout, c.Stderr = os.Stdin, os.Stdout, os.Stderr
			if err := c.Run(); err != nil {
				return err
			}
			// Catch mistakes right away instead of on the next init.
			c2, err := config.Load(p)
			if err != nil {
				return err
			}
			_, err = c2.EffectiveRegistry()
			return err
		},
	}
}
