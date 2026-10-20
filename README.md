# dws: dev workspaces in dev containers

[![CI](https://github.com/parmar94/dev-workspace-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/parmar94/dev-workspace-cli/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/parmar94/dev-workspace-cli.svg)](https://pkg.go.dev/github.com/parmar94/dev-workspace-cli)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

`dws` generates and manages [dev containers](https://containers.dev) for your projects. You choose the languages, AI coding agents (Claude Code, Codex) and tools once in a global config. Then you get a ready-to-build `.devcontainer/` for any project in one command, and update or rebuild it later without hand-editing JSON.

```console
$ dws init --template go-web --tool gh,jq,ripgrep --ssh-agent --yes
Wrote .devcontainer/devcontainer.json
Wrote .devcontainer/scripts/post-create.sh
Wrote .devcontainer/scripts/jq.sh
Wrote .devcontainer/scripts/ripgrep.sh
Wrote .devcontainer/agent-policy/claude-managed-settings.json
Wrote .devcontainer/dws.yaml

Next: start it with `dws rebuild .`, or open the folder in VS Code and run "Reopen in Container".
$ dws rebuild
```

## Features

- **Any mix of languages and tools:** go, node, python, rust, java, gh, aws-cli, terraform, kubectl, docker-in-docker and more. Add your own through Dev Container Features or install scripts. Dependencies are resolved for you.
- **AI agents in the container:** Claude Code and Codex. Logins persist across rebuilds and projects.
- **Your preferences, once:** a global config with defaults, favorites, custom tools and templates.
- **Interactive or scripted:** prompts by default; flags, `--yes`, `--dry-run` and `--json` for scripts and agents.
- **Update and rebuild:** `dws update --add-tool terraform` then `dws rebuild`. Hand edits are detected, never silently overwritten.
- **Credentials without handing them to agents:**
  - SSH agent forwarding.
  - Short-lived AWS credentials and fine-grained, expiring GitHub tokens via `dws creds`, with a guided token-creation page.
  - Agent guard rules, and an opt-in OS-level Claude Code sandbox.
- **Agent skills:** `dws skills install` teaches Claude Code and Codex to set up dev containers for you.

## Install

macOS and Linux (including WSL2). You also need Docker, plus VS Code with the Dev Containers extension or the [devcontainer CLI](https://github.com/devcontainers/cli).

**Prebuilt binary:** download the archive for your platform from the [latest release](https://github.com/parmar94/dev-workspace-cli/releases/latest), check it against `checksums.txt`, and put `dws` on your `PATH`.

**With Go:**

```sh
go install github.com/parmar94/dev-workspace-cli/cmd/dws@latest
```

Check it works with `dws version`. Shell completion: `dws completion --help`.

## Quick start

```sh
dws config init                     # your defaults, in ~/.config/dws/config.yaml
cd my-project
dws init                            # interactive; or: dws init --template go --yes
dws rebuild                         # build and start it (or VS Code: "Reopen in Container")
```

Log in to `claude` / `codex` inside the container once. When the container needs cloud or GitHub access, pass in short-lived credentials from the host:

```sh
dws creds github --create           # opens a pre-filled, repo-scoped token page; paste the token
dws creds aws --profile dev         # temporary credentials from an SSO / assume-role profile
```

Later:

```sh
dws update --add-lang python --remove-agent codex --rebuild
```

## Documentation

- [Getting started](docs/getting-started.md)
- [Configuration](docs/configuration.md): defaults, favorites, templates, how values are chosen
- [Adding tools](docs/registry.md): features vs scripts, dependencies, guard rules, sandbox domains
- [Credentials and security](docs/credentials.md): threat model, SSH, GitHub, AWS, agent guard, Claude sandbox
- [AI agents](docs/agents.md): agents in containers, and agents that drive dws
- [Troubleshooting](docs/troubleshooting.md)
- [Command reference](docs/reference/dws.md)

## Status

dws is pre-1.0. It's used daily, but commands and config may change between minor versions; see the [changelog](CHANGELOG.md). Feedback and issues are welcome.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Please report security issues privately as described in [SECURITY.md](SECURITY.md).

## License

[MIT](LICENSE) © 2026 Ashish Jain
