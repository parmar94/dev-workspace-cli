package cli

import (
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"

	"github.com/parmar94/dev-workspace-cli/internal/assets"
)

// TestBundledSkillsAreValid checks the frontmatter both Claude Code and Codex
// need, and that relative links point at bundled files.
func TestBundledSkillsAreValid(t *testing.T) {
	names, err := skillNames()
	if err != nil || len(names) == 0 {
		t.Fatalf("skills: %v %v", names, err)
	}
	link := regexp.MustCompile(`\]\(([a-z0-9_.-]+\.md)(#[^)]*)?\)`)
	for _, n := range names {
		data, err := fs.ReadFile(assets.Skills, path.Join("skills", n, "SKILL.md"))
		if err != nil {
			t.Fatal(err)
		}
		parts := strings.SplitN(string(data), "---\n", 3)
		if len(parts) != 3 || parts[0] != "" {
			t.Fatalf("%s: SKILL.md must start with --- frontmatter", n)
		}
		var fm map[string]any
		if err := yaml.Unmarshal([]byte(parts[1]), &fm); err != nil {
			t.Fatalf("%s: frontmatter: %v", n, err)
		}
		// Only portable keys, so the skill works in Codex and the Skills API too.
		for k := range fm {
			if k != "name" && k != "description" {
				t.Errorf("%s: non-portable frontmatter key %q", n, k)
			}
		}
		if fm["name"] != n {
			t.Errorf("%s: name = %v, want directory name", n, fm["name"])
		}
		desc, _ := fm["description"].(string)
		if desc == "" || len(desc) > 1024 {
			t.Errorf("%s: description length %d, want 1-1024", n, len(desc))
		}
		if skillDescription(n) != desc {
			t.Errorf("%s: skillDescription mismatch", n)
		}
		if lines := strings.Count(string(data), "\n"); lines > 500 {
			t.Errorf("%s: SKILL.md has %d lines, keep it under 500", n, lines)
		}
		for _, m := range link.FindAllStringSubmatch(string(data), -1) {
			if _, err := fs.Stat(assets.Skills, path.Join("skills", n, m[1])); err != nil {
				t.Errorf("%s: broken link to %s", n, m[1])
			}
		}
	}
}

func TestSkillsInstall(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	if _, _, err := run(t, "", "skills", "install"); err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{
		".claude/skills/dws-devcontainer-setup/SKILL.md",
		".claude/skills/dws-devcontainer-setup/reference.md",
		".agents/skills/dws-credentials/SKILL.md",
	} {
		if _, err := os.Stat(filepath.Join(home, p)); err != nil {
			t.Errorf("missing %s: %v", p, err)
		}
	}

	if _, _, err := run(t, "", "skills", "install", "dws-config"); err == nil || !strings.Contains(err.Error(), "--force") {
		t.Errorf("reinstall without --force: err = %v", err)
	}
	stale := filepath.Join(home, ".claude/skills/dws-config/old.md")
	if err := os.WriteFile(stale, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := run(t, "", "skills", "install", "dws-config", "--force", "-a", "claude"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Error("--force kept a stale file")
	}

	proj := t.TempDir()
	if _, _, err := run(t, "", "skills", "install", "--scope", "project", "--dir", proj, "-a", "codex"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(proj, ".agents/skills/dws-config/SKILL.md")); err != nil {
		t.Error(err)
	}
	if _, err := os.Stat(filepath.Join(proj, ".claude")); !os.IsNotExist(err) {
		t.Error("installed for claude without being asked")
	}

	for _, args := range [][]string{
		{"skills", "install", "nope"},
		{"skills", "install", "-a", "gemini"},
		{"skills", "install", "--scope", "system"},
	} {
		if _, _, err := run(t, "", args...); err == nil {
			t.Errorf("%v: want error", args)
		}
	}
}
