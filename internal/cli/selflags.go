package cli

import (
	"fmt"

	"github.com/spf13/pflag"

	"github.com/parmar94/dev-workspace-cli/internal/devcontainer"
)

// selectionFlags are the flags shared by init and update that set fields of
// a selection directly.
type selectionFlags struct {
	name      string
	baseImage string
	langs     []string
	agents    []string
	tools     []string
	sshAgent  bool
	sshMode   string
	gitConfig bool
	guard     bool
	sandbox   bool
	ports     []int
}

func (o *selectionFlags) register(f *pflag.FlagSet) {
	f.StringVar(&o.name, "name", "", "container name (default: directory name)")
	f.StringVar(&o.baseImage, "base-image", "", "base image (default from config)")
	f.StringSliceVarP(&o.langs, "lang", "l", nil, "languages, comma-separated")
	f.StringSliceVarP(&o.agents, "agent", "a", nil, "AI agents, comma-separated (claude, codex)")
	f.StringSliceVar(&o.tools, "tool", nil, "extra tools, comma-separated")
	f.BoolVar(&o.sshAgent, "ssh-agent", false, "forward the host SSH agent")
	f.StringVar(&o.sshMode, "ssh-mode", "", "how to forward the SSH agent: mac or linux (default: detect host)")
	f.BoolVar(&o.gitConfig, "git-config", false, "mount the host ~/.gitconfig and include it")
	f.BoolVar(&o.guard, "agent-guard", true, "deny agents access to credential files and commands of selected tools (use --agent-guard=false to skip)")
	f.BoolVar(&o.sandbox, "agent-sandbox", false, "turn on Claude Code's OS-level sandbox for its shell commands (blocks credential reads, limits network)")
	f.IntSliceVarP(&o.ports, "port", "p", nil, "ports to forward, comma-separated")
}

// apply sets the fields whose flags were given. Lists replace, not merge.
func (o *selectionFlags) apply(changed func(string) bool, sel *selection) error {
	if changed("name") {
		sel.Name = o.name
	}
	if changed("base-image") {
		sel.BaseImage = o.baseImage
	}
	if changed("lang") {
		sel.Languages = o.langs
	}
	if changed("agent") {
		sel.Agents = o.agents
	}
	if changed("tool") {
		sel.Tools = o.tools
	}
	if changed("ssh-mode") {
		m := devcontainer.SSHMode(o.sshMode)
		if m != devcontainer.SSHMac && m != devcontainer.SSHLinux {
			return fmt.Errorf("--ssh-mode must be mac or linux, got %q", o.sshMode)
		}
		sel.SSH = m
	}
	if changed("ssh-agent") {
		switch {
		case !o.sshAgent:
			sel.SSH = devcontainer.SSHNone
		case sel.SSH == devcontainer.SSHNone:
			sel.SSH = hostSSHMode()
		}
	}
	if changed("git-config") {
		sel.GitConfig = o.gitConfig
	}
	if changed("agent-guard") {
		sel.AgentGuard = o.guard
	}
	if changed("agent-sandbox") {
		sel.AgentSandbox = o.sandbox
	}
	if changed("port") {
		sel.Ports = o.ports
	}
	return nil
}
