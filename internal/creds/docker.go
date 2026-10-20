// Package creds hands short-lived credentials from the host to a running dev
// container. Secrets are only ever passed over stdin, never as command-line
// arguments or environment variables, and never written to the project.
package creds

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
)

// Runner runs an external command. It exists so tests can fake docker/aws.
type Runner interface {
	Run(ctx context.Context, stdin io.Reader, name string, args ...string) ([]byte, error)
}

// ExecRunner runs commands on the host.
type ExecRunner struct{}

// Run executes name with args, returning stdout. On failure the error
// includes stderr.
func (ExecRunner) Run(ctx context.Context, stdin io.Reader, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdin = stdin
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		if msg := strings.TrimSpace(stderr.String()); msg != "" {
			return nil, fmt.Errorf("%s: %w: %s", name, err, msg)
		}
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	return stdout.Bytes(), nil
}

// Container is a running dev container and the user to act as inside it.
type Container struct {
	ID   string
	User string
	r    Runner
}

// FindContainer returns the running dev container for the project at dir,
// using the label the devcontainer CLI and VS Code set on it.
func FindContainer(ctx context.Context, r Runner, dir, user string) (*Container, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	candidates := []string{abs}
	if real, err := filepath.EvalSymlinks(abs); err == nil && real != abs {
		candidates = append(candidates, real)
	}
	for _, p := range candidates {
		out, err := r.Run(ctx, nil, "docker", "ps", "-q", "--filter", "label=devcontainer.local_folder="+p)
		if err != nil {
			return nil, err
		}
		if ids := strings.Fields(string(out)); len(ids) > 0 {
			return &Container{ID: ids[0], User: user, r: r}, nil
		}
	}
	return nil, fmt.Errorf("no running dev container for %s; start it first (VS Code \"Reopen in Container\" or `devcontainer up --workspace-folder %s`)", abs, dir)
}

// Exec runs args in the container as c.User, feeding stdin.
func (c *Container) Exec(ctx context.Context, stdin io.Reader, args ...string) ([]byte, error) {
	flags := []string{"exec", "-u", c.User}
	if stdin != nil {
		flags = append(flags, "-i")
	}
	return c.r.Run(ctx, stdin, "docker", append(append(flags, c.ID), args...)...)
}

// WriteHomeFile writes data to $HOME/rel inside the container with mode 600,
// streaming it over stdin.
func (c *Container) WriteHomeFile(ctx context.Context, rel string, data []byte) error {
	const script = `umask 077; f="$HOME/$1"; mkdir -p "$(dirname "$f")"; cat > "$f"`
	_, err := c.Exec(ctx, bytes.NewReader(data), "sh", "-c", script, "sh", rel)
	return err
}
