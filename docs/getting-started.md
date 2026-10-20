# Getting started

## Requirements

- **Docker:** Docker Desktop, OrbStack, or Docker Engine on Linux.
- **Something to open dev containers with:**
  - VS Code with the [Dev Containers extension](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers), or
  - the [devcontainer CLI](https://github.com/devcontainers/cli) (`npm install -g @devcontainers/cli`, or run it with `npx @devcontainers/cli`).
- **Go 1.24+** to build dws.

## Install

```sh
go install github.com/parmar94/dev-workspace-cli/cmd/dws@latest
# or from a checkout
git clone git@github.com:parmar94/dev-workspace-cli.git && cd dev-workspace-cli
make build   # -> bin/dws
```

Check it works with `dws version`.

## 1. Create your global config

```sh
dws config init    # writes ~/.config/dws/config.yaml with comments
dws config edit    # opens it in $VISUAL / $EDITOR
```

Set the languages, agents and tools you usually want under `defaults`, and star your favorites. Every `dws init` starts from these. See [Configuration](configuration.md).

## 2. Generate a dev container

Run the following in a project folder:

```sh
dws init
```

You're asked, one screen at a time, for:
- a template
- languages, agents and tools (favorites are starred and pre-selected)
- SSH agent forwarding
- host git config
- the agent guard

Anything you pass as a flag isn't asked again.

To skip prompts entirely, which suits scripts and CI:

```sh
dws init --template go-web --yes
dws init --lang go,node --agent claude --tool gh,jq --ssh-agent --yes
```

Preview without writing anything:

```sh
dws init --lang go --agent claude,codex --yes --dry-run
```

dws writes:

```
.devcontainer/
  devcontainer.json        # image, features, mounts, env, ports, lifecycle commands
  scripts/post-create.sh   # runs once after the container is created
  scripts/<name>.sh        # one per script-installed entry
  agent-policy/            # agent guard rules (when an agent is selected)
  dws.yaml                 # what this was generated with
```

Commit `.devcontainer/` so collaborators and future you get the same environment. None of these files contain secrets.

## 3. Open it

- **VS Code:** open the folder and choose **Dev Containers: Reopen in Container**.
- **CLI:**
  ```sh
  devcontainer up --workspace-folder .
  devcontainer exec --workspace-folder . bash
  ```

The first build takes a few minutes while features install. After that it's cached.

## 4. Log in to agents and pass credentials

- **Claude Code / Codex:** run `claude` or `codex` in the container and log in once. Logins are stored in the shared `dws-claude` and `dws-codex` Docker volumes, so they survive rebuilds and are shared between projects.
- **GitHub, AWS:** use `dws creds` from the host. See [Credentials and security](credentials.md).
- **SSH:** if you used `--ssh-agent`, `ssh-add -l` in the container lists your host keys.

## Changing a project's setup later

`dws update` starts from what the project was generated with (its `dws.yaml`), applies your changes, and regenerates:

```sh
dws update --add-tool ripgrep,terraform --remove-agent codex --dry-run   # preview
dws update --add-tool ripgrep,terraform --remove-agent codex
dws update                     # on a terminal: prompts, pre-filled with the current setup
dws update --yes               # just regenerate, e.g. after upgrading dws or editing your registry
```

Then rebuild the container to apply it:

```sh
dws rebuild                    # devcontainer CLI: replaces the container, keeps agent-login volumes
dws update --add-lang python --rebuild   # or both in one go
```

In VS Code you can instead run **Dev Containers: Rebuild Container**.

Notes:
- dws records a hash of every file it generates. If you edit `devcontainer.json` or a script by hand, `dws update` stops and lists the files, rather than silently overwriting your edits. Move the edits elsewhere, or pass `--force`.
- `update` replaces `devcontainer.json`, `dws.yaml`, `scripts/` and `agent-policy/`. Other files in `.devcontainer/` are left alone.
- Credentials passed with `dws creds` live in the container, so pass them in again after a rebuild.

## Let an agent do it

Install the bundled skills and ask Claude Code or Codex (running on your machine) to "set up a dev container for this project":

```sh
dws skills install
```

See [AI agents](agents.md#let-an-agent-set-up-projects-for-you).
