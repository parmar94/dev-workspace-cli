package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/parmar94/dev-workspace-cli/internal/creds"
	"github.com/parmar94/dev-workspace-cli/internal/devcontainer"
)

// runner is swapped in tests.
var runner creds.Runner = creds.ExecRunner{}

const credsWarning = "Agents in the container can still use these credentials until they expire; scope them narrowly."

func newCredsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "creds",
		Short: "Pass short-lived credentials into a running dev container",
		Long: `Pass short-lived credentials from the host into a running dev container.

Long-term secrets stay on the host. Credentials are streamed over stdin into
the container's filesystem (not the project or a volume), so a rebuild wipes
them.`,
	}
	var user string
	cmd.PersistentFlags().StringVar(&user, "user", "", "container user (default: remote_user from .devcontainer/dws.yaml)")
	cmd.AddCommand(newCredsAWSCmd(&user), newCredsGitHubCmd(&user))
	return cmd
}

// container finds the running dev container for dir and the user to act as.
func container(ctx context.Context, dir, user string) (*creds.Container, error) {
	if user == "" {
		st, err := devcontainer.LoadState(dir)
		switch {
		case errors.Is(err, fs.ErrNotExist):
			return nil, fmt.Errorf("no %s in %s; pass --user", devcontainer.StateFile, dir)
		case err != nil:
			return nil, err
		}
		user = st.User
	}
	return creds.FindContainer(ctx, runner, dir, user)
}

func newCredsAWSCmd(user *string) *cobra.Command {
	var profile, region string
	cmd := &cobra.Command{
		Use:   "aws [dir]",
		Short: "Pass temporary AWS credentials from a host profile (SSO or assumed role)",
		Example: `  aws sso login --profile dev-readonly
  dws creds aws --profile dev-readonly --region eu-west-1`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			ctr, err := container(ctx, dirArg(args), *user)
			if err != nil {
				return err
			}
			c, err := creds.ExportAWS(ctx, runner, profile)
			if err != nil {
				return err
			}
			if err := creds.InstallAWS(ctx, ctr, c, region); err != nil {
				return err
			}
			msg := "Installed AWS credentials for profile " + profile
			if exp, err := c.Expires(); err == nil {
				msg += fmt.Sprintf(", expiring %s (in %s)", exp.Local().Format(time.Kitchen), time.Until(exp).Round(time.Minute))
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "%s.\n%s\n", msg, credsWarning)
			return nil
		},
	}
	cmd.Flags().StringVar(&profile, "profile", "", "host AWS profile to export (required)")
	cmd.Flags().StringVar(&region, "region", "", "default region to set in the container")
	_ = cmd.MarkFlagRequired("profile")
	return cmd
}

func newCredsGitHubCmd(user *string) *cobra.Command {
	var fromStdin bool
	cmd := &cobra.Command{
		Use:   "github [dir]",
		Short: "Log gh in with a fine-grained, expiring personal access token",
		Long: `Log gh in with a fine-grained personal access token and set it up as git's
credential helper. Classic tokens and tokens without an expiry are refused.

Create one at https://github.com/settings/personal-access-tokens/new limited
to the repositories and permissions this project needs.`,
		Example: `  dws creds github
  pbpaste | dws creds github --token-stdin`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			ctr, err := container(ctx, dirArg(args), *user)
			if err != nil {
				return err
			}
			var raw string
			switch {
			case fromStdin:
				b, err := io.ReadAll(cmd.InOrStdin())
				if err != nil {
					return err
				}
				raw = string(b)
			case isInteractive():
				err := runForm(huh.NewGroup(huh.NewInput().
					Title("Fine-grained GitHub token").
					EchoMode(huh.EchoModePassword).
					Value(&raw)))
				if err != nil {
					return err
				}
			default:
				return errors.New("no terminal to prompt for the token; use --token-stdin")
			}

			tok, err := creds.CheckGitHubToken(ctx, &http.Client{Timeout: 15 * time.Second}, raw)
			if err != nil {
				return err
			}
			if err := creds.InstallGitHub(ctx, ctr, tok); err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "Logged gh in as %s; token expires %s.\n%s\n", tok.Login, tok.Expires, credsWarning)
			return nil
		},
	}
	cmd.Flags().BoolVar(&fromStdin, "token-stdin", false, "read the token from stdin")
	return cmd
}

func dirArg(args []string) string {
	if len(args) == 1 {
		return args[0]
	}
	return "."
}
