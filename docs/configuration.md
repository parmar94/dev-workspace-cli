# Configuration

## Where it lives

`dws config path` prints the location. dws uses the first of:

1. `--config <file>` on any command
2. `$DWS_CONFIG`
3. `$XDG_CONFIG_HOME/dws/config.yaml`
4. `~/.config/dws/config.yaml`

A missing file is fine; dws then uses only built-in entries and no defaults.

| Command | Does |
| --- | --- |
| `dws config init` | write a commented starter file (`--force` to overwrite) |
| `dws config edit` | open it in `$VISUAL` / `$EDITOR` (creates it if missing), then validate |
| `dws config show` | print it |

## Full example

```yaml
defaults:
  languages: [go]
  agents: [claude]
  tools: [gh, jq]
  ssh_agent: true
  base_image: mcr.microsoft.com/devcontainers/base:ubuntu
  remote_user: vscode

favorites: [go, node, claude, gh, jq, ripgrep]

registry:
  tools:
    httpie:
      description: HTTP client
      script: sudo apt-get update && sudo apt-get install -y httpie

removed: [fzf]

templates:
  my-go:
    description: My usual Go service
    languages: [go]
    agents: [claude]
    tools: [gh, jq, ripgrep]
    ssh_agent: true
```

### `defaults`

The starting selection for every `dws init`.

| Key | Meaning |
| --- | --- |
| `languages`, `agents`, `tools` | entry names from `dws tools list` |
| `ssh_agent` | forward your SSH agent by default |
| `base_image` | image the container is built from (default `mcr.microsoft.com/devcontainers/base:ubuntu`) |
| `remote_user` | user inside the container (default `vscode`). Volume mounts and scripts use this user's home. |

If you change `base_image`, pick an image that has your `remote_user` with passwordless `sudo`. The [devcontainers base images](https://github.com/devcontainers/images) all do.

### `favorites`

Entries starred (★) and listed first in interactive prompts, and marked `"favorite": true` in `dws tools list --json`. Favorites aren't selected automatically; put entries in `defaults` for that.

### `registry` and `removed`

Add your own languages, agents and tools, override built-in ones, or hide built-ins. See [Adding tools](registry.md).

### `templates`

Your own presets, used with `dws init --template <name>`. A template with the same name as a built-in replaces it.

| Key | Meaning |
| --- | --- |
| `description` | shown in `dws templates list` |
| `languages`, `agents`, `tools` | entry names |
| `ssh_agent` | optional; overrides `defaults.ssh_agent` |
| `base_image` | optional; overrides `defaults.base_image` |

Built-in templates:

| Template | Contents |
| --- | --- |
| `go` | go, claude, gh, jq |
| `go-web` | go, node, claude, gh, jq |
| `python` | python, claude, gh, jq |
| `fullstack` | go, node, python, claude, codex, gh, jq, ripgrep, docker-in-docker |

## How values are chosen

For each setting, the highest-priority source that sets it wins:

1. flags on `dws init`
2. the `--template`
3. `defaults` in your config

**Lists are replaced, not merged.** `--template go --tool ripgrep` gives you only `ripgrep` as a tool, not `gh, jq, ripgrep`. Pass the full list: `--tool gh,jq,ripgrep`.

In interactive mode, each prompt is pre-filled with the merged result, and prompts for values you set with flags are skipped.

## Editing safely

- `dws config edit` validates the file when the editor closes. Other commands validate on every load and print every problem at once, for example `invalid registry: tool "x": set exactly one of feature or script`.
- `dws tools add` and `dws tools remove` rewrite the file, so **comments are lost**. Edit by hand if you want to keep them.
- Never put secrets in the config. See [Credentials and security](credentials.md).
