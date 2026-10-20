package cli

import (
	"context"
	"io"
	"strings"
	"testing"
)

// recordingRunner records commands and returns canned stdout per command.
type recordingRunner struct {
	calls []string
	out   map[string]string
}

func (r *recordingRunner) Run(_ context.Context, _ io.Reader, name string, args ...string) ([]byte, error) {
	line := strings.Join(append([]string{name}, args...), " ")
	r.calls = append(r.calls, line)
	for prefix, out := range r.out {
		if strings.HasPrefix(line, prefix) {
			return []byte(out), nil
		}
	}
	return nil, nil
}

func withRunner(t *testing.T, r *recordingRunner) {
	old := runner
	runner = r
	t.Cleanup(func() { runner = old })
}

func TestCredsGitHubCreate(t *testing.T) {
	r := &recordingRunner{out: map[string]string{"git -C": "git@github.com:acme/api.git\n"}}
	withRunner(t, r)

	// No terminal and no --token-stdin: print the guided URL and stop.
	_, errOut, err := run(t, "", "creds", "github", t.TempDir(), "--create", "--access", "push", "--days", "7", "--perm", "issues=write")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"https://github.com/settings/personal-access-tokens/new?",
		"target_name=acme", "expires_in=7", "contents=write", "issues=write",
		`pick api`, "Then pass it in from a terminal",
	} {
		if !strings.Contains(errOut, want) {
			t.Errorf("output missing %q:\n%s", want, errOut)
		}
	}
	opened := false
	for _, c := range r.calls {
		if (strings.HasPrefix(c, "open ") || strings.HasPrefix(c, "xdg-open ")) && strings.Contains(c, "personal-access-tokens") {
			opened = true
		}
		if strings.HasPrefix(c, "docker") {
			t.Errorf("looked for a container in URL-only mode: %s", c)
		}
	}
	if !opened {
		t.Errorf("browser not opened: %v", r.calls)
	}

	// --repo and --no-browser skip git and the browser.
	r.calls = nil
	if _, _, err := run(t, "", "creds", "github", "--create", "--repo", "o/n", "--no-browser"); err != nil {
		t.Fatal(err)
	}
	if len(r.calls) != 0 {
		t.Errorf("unexpected commands: %v", r.calls)
	}

	// With --token-stdin it carries on to the container.
	_, _, err = run(t, "", "creds", "github", t.TempDir(), "--create", "--repo", "o/n", "--no-browser", "--token-stdin", "--user", "vscode")
	if err == nil || !strings.Contains(err.Error(), "start it first") {
		t.Errorf("err = %v", err)
	}

	if _, _, err := run(t, "", "creds", "github", "--create", "--repo", "o/n", "--access", "all"); err == nil {
		t.Error("bad preset should fail")
	}
}
