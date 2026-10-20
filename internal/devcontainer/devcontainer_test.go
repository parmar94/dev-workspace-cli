package devcontainer

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/parmar94/dev-workspace-cli/internal/registry"
)

var update = flag.Bool("update", false, "rewrite golden files")

func resolve(t *testing.T, names ...string) []registry.Entry {
	t.Helper()
	r, err := registry.Builtin()
	if err != nil {
		t.Fatal(err)
	}
	es, _, err := r.Resolve(names)
	if err != nil {
		t.Fatal(err)
	}
	return es
}

func base(name string) Spec {
	return Spec{Name: name, Image: "mcr.microsoft.com/devcontainers/base:ubuntu", RemoteUser: "vscode"}
}

func TestRenderGolden(t *testing.T) {
	tests := map[string]func(*testing.T) Spec{
		"bare": func(t *testing.T) Spec {
			return base("bare")
		},
		"go-single": func(t *testing.T) Spec {
			s := base("go-single")
			s.Entries = resolve(t, "go")
			return s
		},
		"multi-lang-claude": func(t *testing.T) Spec {
			s := base("multi")
			s.Entries = resolve(t, "go", "python", "claude", "gh")
			s.Ports = []int{8080}
			return s
		},
		"codex-adds-node-scripts": func(t *testing.T) Spec {
			s := base("codex")
			s.Entries = resolve(t, "go", "codex", "jq")
			return s
		},
		"ssh-mac-gitconfig": func(t *testing.T) Spec {
			s := base("ssh-mac")
			s.Entries = resolve(t, "go")
			s.SSH = SSHMac
			s.GitConfig = true
			return s
		},
		"guard-claude-codex-aws": func(t *testing.T) Spec {
			s := base("guard")
			s.Entries = resolve(t, "go", "claude", "codex", "aws-cli", "gh")
			s.AgentGuard = true
			return s
		},
		"sandbox-claude-python-aws": func(t *testing.T) Spec {
			s := base("sandbox")
			s.Entries = resolve(t, "python", "claude", "aws-cli")
			s.AgentGuard = true
			s.AgentSandbox = true
			return s
		},
		"sandbox-only": func(t *testing.T) Spec {
			s := base("sandbox-only")
			s.Entries = resolve(t, "go", "claude")
			s.AgentSandbox = true
			return s
		},
		"guard-without-agents": func(t *testing.T) Spec {
			s := base("guard-none")
			s.Entries = resolve(t, "go", "aws-cli")
			s.AgentGuard = true
			return s
		},
		"ssh-linux-root": func(t *testing.T) Spec {
			s := base("ssh-linux")
			s.RemoteUser = "root"
			s.Entries = resolve(t, "claude")
			s.SSH = SSHLinux
			return s
		},
	}

	for name, mk := range tests {
		t.Run(name, func(t *testing.T) {
			files, err := Render(mk(t))
			if err != nil {
				t.Fatal(err)
			}
			var got strings.Builder
			for _, f := range files {
				mode := ""
				if f.Exec {
					mode = " (exec)"
				}
				got.WriteString("=== " + f.Path + mode + "\n")
				got.Write(f.Data)
			}

			golden := filepath.Join("testdata", name+".golden")
			if *update {
				if err := os.MkdirAll("testdata", 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(golden, []byte(got.String()), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			want, err := os.ReadFile(golden)
			if err != nil {
				t.Fatalf("%v (run go test ./internal/devcontainer -update)", err)
			}
			if got.String() != string(want) {
				t.Errorf("output mismatch for %s\n--- got\n%s\n--- want\n%s", golden, got.String(), want)
			}
		})
	}
}

func TestRenderErrors(t *testing.T) {
	if _, err := Render(Spec{}); err == nil {
		t.Error("empty spec should fail")
	}
	s := base("x")
	s.SSH = "windows"
	if _, err := Render(s); err == nil {
		t.Error("unknown ssh mode should fail")
	}
}

func TestFeatureID(t *testing.T) {
	for in, want := range map[string]string{
		"ghcr.io/devcontainers/features/go:1":                      "ghcr.io/devcontainers/features/go",
		"ghcr.io/anthropics/devcontainer-features/claude-code:1.0": "ghcr.io/anthropics/devcontainer-features/claude-code",
		"localhost:5000/features/x":                                "localhost:5000/features/x",
	} {
		if got := featureID(in); got != want {
			t.Errorf("featureID(%q) = %q, want %q", in, got, want)
		}
	}
}
