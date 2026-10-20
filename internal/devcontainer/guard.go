package devcontainer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path"
	"slices"
	"strings"

	"github.com/parmar94/dev-workspace-cli/internal/registry"
)

// Agent guard: deny rules that stop AI agents from casually reading the
// credentials of tools installed alongside them. These are guardrails, not a
// security boundary: an agent running as the same user can still get at
// anything the user can (for example via a script, or sudo).

// policyDir holds generated agent policy files; post-create.sh installs them.
var policyDir = path.Join(Dir, "agent-policy")

const (
	// Claude Code reads drop-in files from managed-settings.d, which outrank
	// user and project settings and are merged with other managed files.
	claudeManagedPath = "/etc/claude-code/managed-settings.d/dws-guard.json"
	codexBeginMarker  = "# >>> dws agent guard"
	codexEndMarker    = "# <<< dws agent guard"
)

// agentPolicy says which agent policies to generate.
type agentPolicy struct {
	// Guard adds deny rules (Claude permissions, Codex env filters).
	Guard bool
	// Sandbox turns on Claude Code's OS-level Bash sandbox.
	Sandbox bool
}

// agentGuard returns the policy files and the post-create steps that install
// them, for the agents among entries. It returns nothing if no supported
// agent is selected.
func agentGuard(entries []registry.Entry, p agentPolicy) ([]File, string, error) {
	g := registry.BaseGuard
	domains := slices.Clone(registry.BaseDomains)
	var agents []string
	for _, e := range entries {
		g = g.Merge(e.Guard)
		for _, d := range e.Domains {
			if !slices.Contains(domains, d) {
				domains = append(domains, d)
			}
		}
		if e.Kind == registry.KindAgent {
			agents = append(agents, e.Name)
		}
	}

	var files []File
	var steps strings.Builder
	for _, a := range agents {
		switch a {
		case "claude":
			f, err := claudePolicy(g, domains, p)
			if err != nil {
				return nil, "", err
			}
			files = append(files, f)
			if p.Sandbox {
				// bubblewrap and socat are the sandbox's Linux dependencies.
				steps.WriteString("\necho \"==> Installing Claude Code sandbox dependencies\"\n")
				steps.WriteString("sudo apt-get update && sudo apt-get install -y --no-install-recommends bubblewrap socat\n")
			}
			steps.WriteString("\necho \"==> Installing Claude Code agent policy\"\n")
			fmt.Fprintf(&steps, "sudo install -D -m 0644 ../agent-policy/%s %s\n", path.Base(f.Path), claudeManagedPath)
		case "codex":
			// Codex has no read-deny rules; it can only hide env vars from the
			// commands it runs.
			if !p.Guard || len(g.Env) == 0 {
				continue
			}
			f := codexPolicy(g)
			files = append(files, f)
			steps.WriteString("\necho \"==> Installing Codex agent guard\"\n")
			steps.WriteString(codexInstall(path.Base(f.Path)))
		}
	}
	return files, steps.String(), nil
}

func claudePolicy(g registry.Guard, domains []string, p agentPolicy) (File, error) {
	policy := map[string]any{}
	if p.Guard {
		var deny []string
		for _, path := range g.Paths {
			deny = append(deny, "Read("+path+")")
		}
		for _, c := range g.Commands {
			deny = append(deny, "Bash("+c+" *)")
		}
		policy["permissions"] = map[string]any{"deny": deny}
	}
	if p.Sandbox {
		policy["sandbox"] = claudeSandbox(g, domains)
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(policy); err != nil {
		return File{}, err
	}
	return File{Path: path.Join(policyDir, "claude-managed-settings.json"), Data: buf.Bytes()}, nil
}

// claudeSandbox enables Claude Code's Bash sandbox. Unlike permission rules,
// it is enforced by the OS for every command Claude runs, including scripts.
func claudeSandbox(g registry.Guard, domains []string) map[string]any {
	var denyRead []string
	for _, p := range g.Paths {
		// Sandbox paths are filesystem paths, not gitignore patterns: deny the
		// directory itself, and anchor "**/x" at the workspaces folder.
		p = strings.TrimSuffix(p, "/**")
		if strings.HasPrefix(p, "**/") {
			p = "/workspaces/" + p
		}
		denyRead = append(denyRead, p)
	}
	var envVars []map[string]string
	for _, e := range g.Env {
		if !strings.Contains(e, "*") { // the sandbox matches exact names only
			envVars = append(envVars, map[string]string{"name": e, "mode": "deny"})
		}
	}
	sandbox := map[string]any{
		"enabled": true,
		// Docker containers are unprivileged, so bubblewrap can't mount a
		// fresh /proc. The container itself is the outer boundary.
		"enableWeakerNestedSandbox": true,
		"filesystem":                map[string]any{"denyRead": denyRead},
		"network":                   map[string]any{"allowedDomains": domains},
	}
	if len(envVars) > 0 {
		sandbox["credentials"] = map[string]any{"envVars": envVars}
	}
	return sandbox
}

func codexPolicy(g registry.Guard) File {
	var b strings.Builder
	b.WriteString(codexBeginMarker + " (managed by dws; edits between these markers are replaced)\n")
	b.WriteString("[shell_environment_policy]\n")
	b.WriteString("filters = { ")
	for i, e := range g.Env {
		if i > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "%q = \"exclude\"", e)
	}
	b.WriteString(" }\n")
	b.WriteString(codexEndMarker + "\n")
	return File{Path: path.Join(policyDir, "codex-config.toml"), Data: []byte(b.String())}
}

// codexInstall merges the dws block into ~/.codex/config.toml, which lives in
// a volume shared across projects. An existing dws block is replaced; a
// user-written [shell_environment_policy] table is left alone.
func codexInstall(name string) string {
	return fmt.Sprintf(`cfg="$HOME/.codex/config.toml"
mkdir -p "$(dirname "$cfg")" && touch "$cfg"
sed -i '/^%s/,/^%s/d' "$cfg"
if grep -q '^\[shell_environment_policy\]' "$cfg"; then
  echo "    skipped: $cfg already has [shell_environment_policy]; merge this in yourself:" >&2
  sed 's/^/    /' ../agent-policy/%s >&2
else
  cat ../agent-policy/%s >> "$cfg"
fi
`, codexBeginMarker, codexEndMarker, name, name)
}
