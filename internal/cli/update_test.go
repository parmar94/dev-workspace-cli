package cli

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/parmar94/dev-workspace-cli/internal/devcontainer"
)

func TestUpdate(t *testing.T) {
	cfgPath := filepath.Join(t.TempDir(), "config.yaml")
	dir := t.TempDir()
	must := func(args ...string) (string, string) {
		t.Helper()
		out, errOut, err := runWith(t, cfgPath, args...)
		if err != nil {
			t.Fatalf("dws %s: %v\n%s", strings.Join(args, " "), err, errOut)
		}
		return out, errOut
	}

	must("init", dir, "-y", "--lang", "go", "--agent", "claude", "--tool", "gh")
	st := readState(t, dir)
	if len(st.Generated) == 0 || st.Generated[".devcontainer/devcontainer.json"] == "" {
		t.Fatalf("hashes not recorded: %v", st.Generated)
	}

	// Nothing changed.
	_, errOut := must("update", dir, "-y")
	if !strings.Contains(errOut, "Already up to date") {
		t.Errorf("no-op update: %q", errOut)
	}

	// Dry run reports changes without writing.
	before, _ := os.ReadFile(filepath.Join(dir, ".devcontainer/devcontainer.json"))
	out, errOut := must("update", dir, "--add-tool", "jq", "--remove-agent", "claude", "--add-lang", "python", "--dry-run")
	for _, want := range []string{"+ tool jq", "- agent claude", "+ language python", "new      .devcontainer/scripts/jq.sh", "removed  .devcontainer/agent-policy/claude-managed-settings.json"} {
		if !strings.Contains(errOut, want) {
			t.Errorf("dry-run summary missing %q:\n%s", want, errOut)
		}
	}
	if !strings.Contains(out, "# --- .devcontainer/scripts/jq.sh") {
		t.Errorf("dry-run should print changed files:\n%s", out)
	}
	after, _ := os.ReadFile(filepath.Join(dir, ".devcontainer/devcontainer.json"))
	if string(before) != string(after) {
		t.Error("dry run wrote files")
	}

	// Real update.
	must("update", dir, "--add-tool", "jq", "--remove-agent", "claude", "--ssh-agent=false")
	st = readState(t, dir)
	if !slices.Equal(st.Tools, []string{"gh", "jq"}) || len(st.Agents) != 0 || slices.Equal(st.Languages, nil) {
		t.Errorf("state after update: %+v", st)
	}
	if _, err := os.Stat(filepath.Join(dir, ".devcontainer/agent-policy")); !os.IsNotExist(err) {
		t.Error("agent policy left behind after removing the only agent")
	}

	// Hand edits block an update unless --force.
	dc := filepath.Join(dir, ".devcontainer/devcontainer.json")
	if err := os.WriteFile(dc, append(after, ' '), 0o644); err != nil {
		t.Fatal(err)
	}
	_, _, err := runWith(t, cfgPath, "update", dir, "--add-tool", "ripgrep")
	if err == nil || !strings.Contains(err.Error(), "edited by hand") || !strings.Contains(err.Error(), "devcontainer.json") {
		t.Errorf("hand edit: err = %v", err)
	}
	must("update", dir, "--add-tool", "ripgrep", "--force")
	if got := readState(t, dir).Tools; !slices.Contains(got, "ripgrep") {
		t.Errorf("tools = %v", got)
	}
}

func TestUpdateErrors(t *testing.T) {
	_, _, err := run(t, "", "update", t.TempDir(), "-y")
	if err == nil || !strings.Contains(err.Error(), "run `dws init` first") {
		t.Errorf("no state: err = %v", err)
	}

	dir := t.TempDir()
	if _, _, err := run(t, "", "init", dir, "-y", "--lang", "go"); err != nil {
		t.Fatal(err)
	}
	_, _, err = run(t, "", "update", dir, "--add-tool", "nope")
	if err == nil || !strings.Contains(err.Error(), `unknown tool "nope"`) {
		t.Errorf("unknown tool: err = %v", err)
	}
}

func TestSelectionChanges(t *testing.T) {
	old := devcontainer.State{Name: "a", Languages: []string{"go"}, SSH: devcontainer.SSHMac, Ports: []int{80}}
	sel := old
	sel.Languages = []string{"go", "node"}
	sel.SSH = devcontainer.SSHNone
	sel.Ports = nil
	got := strings.Join(selectionChanges(old, sel), "\n")
	for _, want := range []string{"+ language node", "ssh agent: mac -> off", "ports: [80] -> []"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
}

func TestRebuild(t *testing.T) {
	dir := t.TempDir()
	if _, _, err := run(t, "", "rebuild", dir); err == nil || !strings.Contains(err.Error(), "dws init") {
		t.Errorf("no devcontainer: err = %v", err)
	}
	if _, _, err := run(t, "", "init", dir, "-y"); err != nil {
		t.Fatal(err)
	}

	// A fake devcontainer CLI that records its arguments.
	bin := t.TempDir()
	argsFile := filepath.Join(bin, "args")
	script := "#!/bin/sh\necho \"$@\" > " + argsFile + "\n"
	if err := os.WriteFile(filepath.Join(bin, "devcontainer"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	old := lookPath
	defer func() { lookPath = old }()
	lookPath = func(name string) (string, error) {
		if name == "devcontainer" {
			return filepath.Join(bin, "devcontainer"), nil
		}
		return "", errors.New("not found")
	}

	if _, _, err := run(t, "", "rebuild", dir, "--no-cache"); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(argsFile)
	abs, _ := filepath.Abs(dir)
	want := "up --workspace-folder " + abs + " --remove-existing-container --build-no-cache\n"
	if string(got) != want {
		t.Errorf("args = %q, want %q", got, want)
	}

	// Falls back to npx, and explains when neither exists.
	lookPath = func(string) (string, error) { return "", errors.New("not found") }
	if _, _, err := run(t, "", "rebuild", dir); err == nil || !strings.Contains(err.Error(), "npm install -g @devcontainers/cli") {
		t.Errorf("missing CLI: err = %v", err)
	}
}
