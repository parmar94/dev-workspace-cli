# dev-workspace-cli

`dws` is a Go CLI for setting up and managing project workspaces in dev containers.

> **Status:** early stage, personal use. It may be open-sourced later.

## Goals

- **Per-project workspaces:** generate a dev container for a project, with the language(s) and tools it needs.
- **AI agents:** configure the AI coding agents you want available inside the container.
- **Global config:** keep your defaults, favourite setups and commonly used tools in one place, and reuse them across projects.
- **Update and rebuild:** change a project's dev container config and rebuild the container from the CLI. *(planned)*

## Install

```sh
go install github.com/parmar94/dev-workspace-cli/cmd/dws@latest
# or from a checkout
make build   # -> bin/dws
```

## Quick start

```sh
dws config init            # write ~/.config/dws/config.yaml with starter defaults
dws init                   # interactive: pick a template, languages, agents, tools
dws init --template go-web --yes
dws init ./api --lang go,node --agent claude,codex --tool gh,jq --ssh-agent --yes --dry-run
```

Then open the folder in VS Code and run **Reopen in Container**, or run `devcontainer up --workspace-folder .`.

### What gets generated

```
.devcontainer/
  devcontainer.json      # base image, features, mounts, env, ports
  scripts/post-create.sh # only if a script-based entry or --git-config is used
  scripts/<name>.sh      # one per script-based entry
  agent-policy/          # agent guard rules, installed by post-create.sh
  dws.yaml               # the selections used, for future `dws update`
```

Languages and tools are installed as [Dev Container Features](https://containers.dev/features) where one exists. Otherwise a small install script runs after the container is created. Dependencies are added automatically (for example, `codex` pulls in `node`) and installed first.

AI agent config directories (`~/.claude`, `~/.codex`) are stored in named Docker volumes (`dws-claude`, `dws-codex`). Logins therefore survive rebuilds and are shared across projects.

## Choosing values

Each value is taken from the first source that sets it, from highest to lowest priority:

1. Flags
2. `--template`
3. `defaults` in the global config

A template or a list flag (`--lang`, `--agent`, `--tool`) **replaces** the list from the source below it. It doesn't add to it. In interactive mode, anything not set by a flag is asked, pre-filled with the merged values. `--yes` skips prompts. Set `ACCESSIBLE=1` for plain line-based prompts.

| Flag | Description |
| --- | --- |
| `-t, --template` | start from a template |
| `-l, --lang` / `-a, --agent` / `--tool` | comma-separated entries |
| `--ssh-agent` | forward the host SSH agent |
| `--ssh-mode mac\|linux` | how to forward it (default: detect host) |
| `--git-config` | mount host `~/.gitconfig` and include it |
| `--agent-guard` | deny agents access to tools' credentials (default on; see below) |
| `-p, --port` | ports to forward |
| `--name`, `--base-image` | container name and image |
| `-y, --yes`, `--dry-run`, `--force` | no prompts / print only / overwrite existing `.devcontainer` |

### SSH agent forwarding

VS Code forwards your SSH agent on its own. `--ssh-agent` makes forwarding work with the `devcontainer` CLI and other editors too:

- **mac** (Docker Desktop / OrbStack): mounts `/run/host-services/ssh-auth.sock`.
- **linux**: bind-mounts the host's `$SSH_AUTH_SOCK` at `/ssh-agent`.

Check it inside the container with `ssh-add -l`.

## Credentials and AI agents

**Threat model, plainly:** an AI agent running in the container runs as the same user as your tools. Anything `aws`, `gh` or `git` can read, the agent can read too, for example via a script or `sudo`. dws can't make a secret readable by a tool but not by an agent in the same container. What it does instead:

1. **Keep long-term secrets on the host.** Only short-lived, narrowly scoped credentials go in, on demand, with `dws creds`. If an agent reads one, it's limited in scope and expires soon.
2. **Agent guard (guardrail, not a boundary).** Agents get deny rules for the credential files and credential-printing commands of the tools you selected.

Never put secrets in `devcontainer.json`, `containerEnv` or `dws.yaml`. These get committed, and show up in `docker inspect`.

### `dws creds`

```sh
# AWS: exports temporary credentials from a host SSO / assume-role profile.
aws sso login --profile dev-readonly
dws creds aws --profile dev-readonly --region eu-west-1

# GitHub: fine-grained PAT with an expiry, limited to this repo.
dws creds github                    # hidden prompt
pbpaste | dws creds github --token-stdin
```

- Credentials are streamed over stdin into the running container's `~/.aws` or gh config. They never appear on a command line or in env, and never in the project or a volume, so a rebuild wipes them.
- **AWS:** profiles with long-term access keys are refused. Use a dedicated, least-privilege role for dev work.
- **GitHub:** classic `ghp_` tokens and tokens without an expiry are refused. `gh` also becomes git's credential helper.
- **SSH:** use `--ssh-agent` (see above). Agents can use your keys while you're connected but can't copy them out.
- **Claude Code / Codex logins:** stored in the `dws-claude` and `dws-codex` volumes. An agent always has access to its own login.

### Agent guard (`--agent-guard`, on by default)

Each registry entry can declare `guard` rules, on top of a base set: `~/.ssh`, `.env` files, `env` and `printenv`.

```yaml
guard:
  paths: ["~/.aws/**"]                          # agents may not read
  commands: ["aws configure export-credentials"] # agents may not run
  env: ["AWS_*"]                                # hidden from agent shells
```

| Agent | What dws installs | Covers |
| --- | --- | --- |
| Claude Code | `/etc/claude-code/managed-settings.d/dws-guard.json` with `permissions.deny` rules (managed settings outrank user and project settings) | `paths` → `Read(...)`, `commands` → `Bash(... *)` |
| Codex | a marked `[shell_environment_policy]` block in `~/.codex/config.toml` (skipped if you already have that table) | `env` only. Codex has no read-deny rules |

**Limits:**
- Claude Code's deny rules cover its file tools and the file commands it recognizes, like `cat`. They don't cover a script that opens files itself.
- An agent with `sudo` can remove the policy.
- For OS-level enforcement, see Claude Code's [sandbox](https://code.claude.com/docs/en/sandboxing) `sandbox.credentials` settings. Inside Docker these need `enableWeakerNestedSandbox`.

Turn the guard off with `--agent-guard=false`. Add rules for your own tools with `dws tools add ... --guard-path/--guard-cmd/--guard-env`.

## Global config

`dws config path|show|init|edit`. The file lives at `$DWS_CONFIG`, then `$XDG_CONFIG_HOME/dws/config.yaml`, then `~/.config/dws/config.yaml`.

```yaml
defaults:
  languages: [go]
  agents: [claude]
  tools: [gh]
  ssh_agent: true
  base_image: mcr.microsoft.com/devcontainers/base:ubuntu
  remote_user: vscode
favorites: [go, claude, gh, jq]   # starred and listed first in prompts
registry:                         # your own entries / overrides
  tools:
    httpie:
      description: HTTP client
      script: sudo apt-get update && sudo apt-get install -y httpie
removed: [fzf]                    # hide built-in entries
templates:
  my-go:
    languages: [go]
    agents: [claude]
    tools: [gh, jq, ripgrep]
```

### Registry: languages, agents and tools

```sh
dws tools list [--kind tool]
dws tools add httpie --script "sudo apt-get update && sudo apt-get install -y httpie" --favorite
dws tools add deno --kind language --feature ghcr.io/devcontainers-extra/features/deno:1
dws tools add gemini --kind agent --requires node --script "npm i -g @google/gemini-cli"
dws tools remove fzf
```

Each entry needs either a `feature` (with optional `options`) or a `script`. It can also have `requires`, `mounts` and `env`. In mounts, `{{home}}` expands to the remote user's home directory. `tools add` and `tools remove` rewrite the config file, so comments in it are not kept.

### Templates

```sh
dws templates list
dws templates show go-web
```

The built-in templates are `go`, `go-web`, `python` and `fullstack`. Add your own under `templates:` in the config.

## Development

```sh
make build test lint
go test ./internal/devcontainer -update   # regenerate golden files
```

Work happens on feature branches off `main`.

**Planned:** `dws update`, `dws rebuild`, a guided flow for creating fine-grained GitHub PATs (GitHub has no API to create them), and opt-in Claude Code sandbox settings.

## License

[MIT](LICENSE) © 2026 Ashish Jain
