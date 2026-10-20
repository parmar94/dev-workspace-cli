package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"

	"github.com/parmar94/dev-workspace-cli/internal/devcontainer"
)

// run executes dws with args against an isolated config and returns stdout,
// stderr and the error.
func run(t *testing.T, cfg string, args ...string) (string, string, error) {
	t.Helper()
	cfgPath := filepath.Join(t.TempDir(), "config.yaml")
	if cfg != "" {
		if err := os.WriteFile(cfgPath, []byte(cfg), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	var out, errOut bytes.Buffer
	root := newRootCmd()
	root.SetOut(&out)
	root.SetErr(&errOut)
	root.SetArgs(append([]string{"--config", cfgPath}, args...))
	err := root.Execute()
	return out.String(), errOut.String(), err
}

func readState(t *testing.T, dir string) devcontainer.State {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, devcontainer.StateFile))
	if err != nil {
		t.Fatal(err)
	}
	var s devcontainer.State
	if err := yaml.Unmarshal(data, &s); err != nil {
		t.Fatal(err)
	}
	return s
}

const testConfig = `
defaults:
  languages: [python]
  agents: [claude]
  tools: [gh]
  ssh_agent: true
`

func TestInitMergeOrder(t *testing.T) {
	tests := []struct {
		name  string
		args  []string
		langs string
		tools string
		ssh   bool
	}{
		{"defaults", nil, "python", "gh", true},
		{"template replaces lists", []string{"-t", "fullstack"}, "go,node,python", "gh,jq,ripgrep,docker-in-docker", true},
		{"flags beat template", []string{"-t", "go", "--lang", "rust", "--ssh-agent=false"}, "rust", "gh,jq", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			_, _, err := run(t, testConfig, append([]string{"init", dir, "-y"}, tc.args...)...)
			if err != nil {
				t.Fatal(err)
			}
			s := readState(t, dir)
			if got := strings.Join(s.Languages, ","); got != tc.langs {
				t.Errorf("languages = %s, want %s", got, tc.langs)
			}
			if got := strings.Join(s.Tools, ","); got != tc.tools {
				t.Errorf("tools = %s, want %s", got, tc.tools)
			}
			if (s.SSH != devcontainer.SSHNone) != tc.ssh {
				t.Errorf("ssh = %q, want enabled=%v", s.SSH, tc.ssh)
			}
			if s.Name != filepath.Base(dir) {
				t.Errorf("name = %s", s.Name)
			}
		})
	}
}

func TestInitDryRunWritesNothing(t *testing.T) {
	dir := t.TempDir()
	out, errOut, err := run(t, "", "init", dir, "-y", "--dry-run",
		"--lang", "go", "--agent", "codex", "--tool", "jq", "--ssh-mode", "linux", "-p", "8080")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"# --- .devcontainer/devcontainer.json", "/ssh-agent", "8080", "scripts/codex.sh", "# --- .devcontainer/dws.yaml"} {
		if !strings.Contains(out, want) {
			t.Errorf("dry-run output missing %q", want)
		}
	}
	if !strings.Contains(errOut, "Also adding required: node") {
		t.Errorf("missing dependency note: %q", errOut)
	}
	if _, err := os.Stat(filepath.Join(dir, ".devcontainer")); !os.IsNotExist(err) {
		t.Error("dry run wrote files")
	}
}

func TestInitRefusesExistingWithoutForce(t *testing.T) {
	dir := t.TempDir()
	if _, _, err := run(t, "", "init", dir, "-y"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := run(t, "", "init", dir, "-y"); err == nil || !strings.Contains(err.Error(), "--force") {
		t.Errorf("err = %v, want --force hint", err)
	}
	if _, _, err := run(t, "", "init", dir, "-y", "--force", "--lang", "go"); err != nil {
		t.Fatal(err)
	}
}

func TestInitErrors(t *testing.T) {
	tests := map[string][]string{
		"unknown template": {"-t", "nope"},
		"unknown language": {"--lang", "cobol"},
		"wrong kind":       {"--lang", "gh"},
		"bad ssh mode":     {"--ssh-mode", "windows"},
	}
	want := map[string]string{
		"unknown template": `unknown template "nope"`,
		"unknown language": `unknown language "cobol"`,
		"wrong kind":       `"gh" is a tool, not a language`,
		"bad ssh mode":     "--ssh-mode must be mac or linux",
	}
	for name, args := range tests {
		t.Run(name, func(t *testing.T) {
			_, _, err := run(t, "", append([]string{"init", t.TempDir(), "-y", "--dry-run"}, args...)...)
			if err == nil || !strings.Contains(err.Error(), want[name]) {
				t.Errorf("err = %v, want %q", err, want[name])
			}
		})
	}
}
