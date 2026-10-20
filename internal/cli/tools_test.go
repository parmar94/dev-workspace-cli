package cli

import (
	"encoding/json"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/parmar94/dev-workspace-cli/internal/config"
)

func TestToolsAddRemove(t *testing.T) {
	cfgPath := filepath.Join(t.TempDir(), "config.yaml")
	must := func(args ...string) string {
		t.Helper()
		out, _, err := runWith(t, cfgPath, args...)
		if err != nil {
			t.Fatalf("dws %s: %v", strings.Join(args, " "), err)
		}
		return out
	}

	must("tools", "add", "httpie", "-d", "HTTP client", "--script", "apt install httpie", "--favorite")
	must("tools", "add", "go", "--kind", "language", "--feature", "ghcr.io/devcontainers/features/go:1", "--option", "version=1.24")
	must("tools", "add", "gemini", "--kind", "agent", "--requires", "node", "--script", "npm i -g gemini", "--env", "A=b")

	out := must("tools", "list")
	for _, want := range []string{"★ httpie", "custom", "override", "gemini", "script (needs node)"} {
		if !strings.Contains(out, want) {
			t.Errorf("list missing %q:\n%s", want, out)
		}
	}
	if out := must("tools", "list", "--kind", "agents"); strings.Contains(out, "httpie") {
		t.Errorf("kind filter leaked tools:\n%s", out)
	}

	// The new entries are usable from init.
	dry := must("init", t.TempDir(), "-y", "--dry-run", "--lang", "go", "--agent", "gemini", "--tool", "httpie")
	for _, want := range []string{`"version": "1.24"`, "scripts/gemini.sh", "apt install httpie", `"A": "b"`} {
		if !strings.Contains(dry, want) {
			t.Errorf("init output missing %q", want)
		}
	}

	// Removing a dependency that something still needs is refused.
	if _, _, err := runWith(t, cfgPath, "tools", "remove", "node"); err == nil || !strings.Contains(err.Error(), `unknown entry "node"`) {
		t.Errorf("removing node: err = %v", err)
	}

	must("tools", "remove", "httpie", "fzf")
	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if slices.Contains(cfg.Favorites, "httpie") {
		t.Error("httpie still a favorite")
	}
	if !slices.Equal(cfg.Removed, []string{"fzf"}) {
		t.Errorf("removed = %v, want [fzf]", cfg.Removed)
	}
	if out := must("tools", "list"); strings.Contains(out, "fzf") || strings.Contains(out, "httpie") {
		t.Errorf("removed entries still listed:\n%s", out)
	}

	// Adding a hidden built-in back un-hides it.
	must("tools", "add", "fzf", "--script", "apt install fzf")
	cfg, _ = config.Load(cfgPath)
	if len(cfg.Removed) != 0 {
		t.Errorf("removed = %v, want empty", cfg.Removed)
	}
}

func TestToolsAddErrors(t *testing.T) {
	tests := map[string][]string{
		"no install method": {"tools", "add", "x"},
		"both methods":      {"tools", "add", "x", "--feature", "f", "--script", "s"},
		"bad kind":          {"tools", "add", "x", "--kind", "lib", "--script", "s"},
		"bad option":        {"tools", "add", "x", "--feature", "f", "--option", "novalue"},
		"bad name":          {"tools", "add", "Bad Name", "--script", "s"},
	}
	for name, args := range tests {
		t.Run(name, func(t *testing.T) {
			if _, _, err := run(t, "", args...); err == nil {
				t.Error("want error")
			}
		})
	}
}

func TestListJSON(t *testing.T) {
	out, _, err := run(t, "favorites: [gh]\n", "tools", "list", "--json", "--kind", "tool")
	if err != nil {
		t.Fatal(err)
	}
	var tools []toolJSON
	if err := json.Unmarshal([]byte(out), &tools); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out)
	}
	var gh *toolJSON
	for i := range tools {
		if tools[i].Kind != "tool" {
			t.Errorf("kind filter leaked %s", tools[i].Name)
		}
		if tools[i].Name == "gh" {
			gh = &tools[i]
		}
	}
	if gh == nil || !gh.Favorite || gh.Install != "feature" || !slices.Contains(gh.Guard.Commands, "gh auth token") {
		t.Errorf("gh = %+v", gh)
	}

	out, _, err = run(t, "", "templates", "list", "--json")
	if err != nil {
		t.Fatal(err)
	}
	var tmpls []struct {
		Name      string   `json:"name"`
		Languages []string `json:"languages"`
	}
	if err := json.Unmarshal([]byte(out), &tmpls); err != nil || len(tmpls) == 0 {
		t.Fatalf("templates JSON: %v\n%s", err, out)
	}
}

func TestTemplatesCommands(t *testing.T) {
	cfg := "templates:\n  mine:\n    description: Mine\n    languages: [rust]\n"
	out, _, err := run(t, cfg, "templates", "list")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"go-web", "built-in", "mine", "custom"} {
		if !strings.Contains(out, want) {
			t.Errorf("list missing %q:\n%s", want, out)
		}
	}
	out, _, err = run(t, cfg, "templates", "show", "mine")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "- rust") {
		t.Errorf("show output:\n%s", out)
	}
	if _, _, err := run(t, cfg, "templates", "show", "nope"); err == nil {
		t.Error("unknown template should fail")
	}
}
