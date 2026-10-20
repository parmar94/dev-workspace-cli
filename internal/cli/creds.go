package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"runtime"
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
	var fromStdin, create, noBrowser bool
	var repo, access string
	var days int
	var perms []string
	cmd := &cobra.Command{
		Use:   "github [dir]",
		Short: "Log gh in with a fine-grained, expiring personal access token",
		Long: `Log gh in with a fine-grained personal access token and set it up as git's
credential helper. Classic tokens and tokens without an expiry are refused.

With --create, dws opens GitHub's "new token" page pre-filled for this
project: name, expiry, the repository owner and a permission preset. GitHub
can't pre-select repositories, so choose "Only select repositories" and pick
this one, generate the token, and paste it at the prompt.

Access presets: read (contents:read), push (contents:write),
pr (contents:write + pull_requests:write). Add more with --perm.`,
		Example: `  dws creds github --create                 # guided: opens a pre-filled token page
  dws creds github --create --access push --days 7 --perm issues=write
  dws creds github                          # paste an existing token
  pbpaste | dws creds github --token-stdin`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			errOut := cmd.ErrOrStderr()
			if create {
				extra, err := parseKV[string](perms, "--perm")
				if err != nil {
					return err
				}
				if err := guideTokenCreation(ctx, errOut, dirArg(args), repo, access, days, extra, !noBrowser); err != nil {
					return err
				}
				if !fromStdin && !isInteractive() {
					fmt.Fprintf(errOut, "\nThen pass it in from a terminal: dws creds github %s\n", dirArg(args))
					return nil
				}
			}
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
	f := cmd.Flags()
	f.BoolVar(&fromStdin, "token-stdin", false, "read the token from stdin")
	f.BoolVar(&create, "create", false, "open GitHub's token page pre-filled for this project first")
	f.StringVar(&repo, "repo", "", "with --create: owner/name (default: from the git remote \"origin\")")
	f.StringVar(&access, "access", "pr", "with --create: permission preset: read, push or pr")
	f.IntVar(&days, "days", 30, "with --create: days until the token expires (1-366)")
	f.StringArrayVar(&perms, "perm", nil, "with --create: extra permission name=read|write|admin (repeatable)")
	f.BoolVar(&noBrowser, "no-browser", false, "with --create: print the URL instead of opening a browser")
	return cmd
}

// guideTokenCreation prints (and optionally opens) a pre-filled fine-grained
// token URL with instructions for the parts GitHub can't pre-fill.
func guideTokenCreation(ctx context.Context, w io.Writer, dir, repo, access string, days int, extra map[string]string, browser bool) error {
	var owner, name string
	var err error
	if repo != "" {
		owner, name, err = creds.ParseGitHubRepo(repo)
	} else {
		owner, name, err = creds.GitHubRepo(ctx, runner, dir)
	}
	if err != nil {
		return err
	}
	tmpl, err := creds.NewPATTemplate(owner, name, access, days, extra)
	if err != nil {
		return err
	}
	link := tmpl.URL()

	fmt.Fprintf(w, "Create a fine-grained token for %s/%s:\n\n  %s\n\n", owner, name, link)
	fmt.Fprintf(w, "On that page:\n")
	fmt.Fprintf(w, "  1. Check Resource owner is %s (for organizations, re-select it; GitHub may not apply it).\n", owner)
	fmt.Fprintf(w, "  2. Repository access: choose \"Only select repositories\" and pick %s.\n", name)
	fmt.Fprintf(w, "  3. Review the pre-filled permissions and the %d-day expiry, then Generate token and copy it.\n", days)
	if browser {
		opener := "xdg-open"
		if runtime.GOOS == "darwin" {
			opener = "open"
		}
		if _, err := runner.Run(ctx, nil, opener, link); err != nil {
			fmt.Fprintln(w, "\n(Couldn't open a browser; open the link above yourself.)")
		}
	}
	return nil
}

func dirArg(args []string) string {
	if len(args) == 1 {
		return args[0]
	}
	return "."
}
