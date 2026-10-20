# dev-workspace-cli

`dws` is a Go CLI that generates and manages [dev containers](https://containers.dev) for your projects. You choose the languages, AI agents (Claude Code, Codex) and tools once in a global config, then get a ready-to-build `.devcontainer/` for each project in one command.

> **Status:** early stage, personal use. It may be open-sourced later.

## Features

- **Multi-language projects:** pick any mix of go, node, python, rust, java or your own languages, plus tools like gh, aws-cli, terraform, kubectl and docker-in-docker.
- **AI agents in the container:** Claude Code and Codex, with logins that persist across rebuilds and projects.
- **Global config:** your defaults, favorites, custom tools (features or install scripts) and templates.
- **Interactive or scripted:** prompts by default; flags, `--yes` and `--dry-run` for scripts and agents.
- **Safer credentials:** SSH agent forwarding, short-lived AWS/GitHub credentials with `dws creds`, and agent guard rules that keep agents away from tool credentials.
- **Agent skills:** `dws skills install` teaches Claude Code and Codex to set up dev containers for you.

## Quick start

```sh
go install github.com/parmar94/dev-workspace-cli/cmd/dws@latest

dws config init                                   # your defaults, in ~/.config/dws/config.yaml
cd my-project
dws init                                          # interactive
dws init --template go-web --ssh-agent --yes      # or scripted
devcontainer up --workspace-folder .              # or VS Code: "Reopen in Container"
```

Inside the container, log in to `claude` / `codex` once. From the host, pass in short-lived credentials when needed:

```sh
dws creds github                                  # fine-grained, expiring token
dws creds aws --profile dev-readonly              # temporary creds from SSO / assumed role
```

## Documentation

- [Getting started](docs/getting-started.md)
- [Configuration](docs/configuration.md): defaults, favorites, templates, how values are chosen
- [Adding tools](docs/registry.md): features vs scripts, dependencies, guard rules
- [Credentials and security](docs/credentials.md): threat model, SSH, GitHub, AWS, agent guard
- [AI agents](docs/agents.md): agents in containers, and agents that drive dws
- [Troubleshooting](docs/troubleshooting.md)
- [Command reference](docs/reference/dws.md)

## Development

```sh
make build test lint
make docs                                   # regenerate docs/reference after changing commands
go test ./internal/devcontainer -update     # regenerate golden files
```

Bundled assets live in `internal/assets/`:
- `registry.yaml`: the built-in tools
- `templates/`
- `config.yaml`: the starter config
- `skills/`: agent skills

Work happens on feature branches off `main`.

**Planned:**
- `dws update` and `dws rebuild`
- a guided flow for creating fine-grained GitHub PATs (GitHub has no API to create them)
- opt-in Claude Code sandbox settings

## License

[MIT](LICENSE) © 2026 Ashish Jain
