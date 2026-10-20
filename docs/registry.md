# Adding tools

The registry is everything `dws init` can install. `dws tools list` shows it:

```
KIND      NAME              INSTALL               SOURCE    DESCRIPTION
language  go                feature               built-in  Go toolchain
agent     codex             script (needs node)   built-in  OpenAI Codex CLI
tool      ★ gh              feature               built-in  GitHub CLI
tool      httpie            script                custom    HTTP client
```

`SOURCE` is one of:
- `built-in`: ships with dws
- `custom`: defined only in your config
- `override`: your config replaces a built-in

Add `--json` for machine-readable output, or `--kind tool` to filter.

## Entry fields

| Field | Required | Meaning |
| --- | --- | --- |
| `feature` | one of feature/script | Dev Container Feature reference, e.g. `ghcr.io/devcontainers/features/go:1` |
| `options` | no | feature options, e.g. `{version: "1.24"}` |
| `script` | one of feature/script | shell run as the container user after creation. `sudo` works. |
| `description` | no | shown in lists and prompts |
| `requires` | no | other entry names, installed first and added automatically |
| `mounts` | no | Docker mount specs. `{{home}}` becomes the container user's home. |
| `env` | no | `containerEnv` values. **Never secrets.** |
| `guard` | no | credential paths, commands and env vars to keep from agents; see below |

Names must be lowercase `a-z 0-9 . _ -` and unique across languages, agents and tools.

## Feature or script?

**Prefer a feature.** Browse [containers.dev/features](https://containers.dev/features), prefer `ghcr.io/devcontainers/features/*`, and pin a major version (`:1`) so rebuilds are reproducible. Features install at image build time and are cached.

**Use a script** when there's no feature. Scripts run in `post-create.sh` on every fresh container, so keep them quick and idempotent.

## Examples

```sh
# Language from a community feature
dws tools add deno --kind language \
  --feature ghcr.io/devcontainers-extra/features/deno:1 -d "Deno runtime"

# Feature with options
dws tools add go --kind language \
  --feature ghcr.io/devcontainers/features/go:1 --option version=1.24
#   ^ same name as the built-in, so this overrides it

# apt package
dws tools add httpie -d "HTTP client" \
  --script "sudo apt-get update && sudo apt-get install -y --no-install-recommends httpie"

# Longer script from a file
dws tools add mytool --script-file ./install-mytool.sh

# Agent that needs node, with its config persisted in a shared volume
dws tools add gemini --kind agent --requires node \
  --script "npm install -g @google/gemini-cli" \
  --mount "source=dws-gemini,target={{home}}/.gemini,type=volume" --favorite

# Tool that stores credentials: keep agents away from them
dws tools add azure-cli --feature ghcr.io/devcontainers/features/azure-cli:1 -d "Azure CLI" \
  --guard-path "~/.azure/**" --guard-cmd "az account get-access-token" --guard-env "AZURE_*"
```

Volumes listed in `mounts` are handed to the container user automatically, since Docker creates new volumes owned by root.

The same entries in YAML, under `registry:` in your config:

```yaml
registry:
  agents:
    gemini:
      description: Gemini CLI
      script: npm install -g @google/gemini-cli
      requires: [node]
      mounts: ["source=dws-gemini,target={{home}}/.gemini,type=volume"]
  tools:
    azure-cli:
      feature: ghcr.io/devcontainers/features/azure-cli:1
      guard:
        paths: ["~/.azure/**"]
        commands: ["az account get-access-token"]
        env: ["AZURE_*"]
```

## Guard rules

`guard` describes a tool's credentials so the [agent guard](credentials.md#agent-guard) can keep AI agents away from them:

| Key | Becomes |
| --- | --- |
| `paths` | Claude Code `Read(<path>)` deny rules (gitignore-style globs; `~/` = home) |
| `commands` | Claude Code `Bash(<command> *)` deny rules (prefix match) |
| `env` | Codex `shell_environment_policy` exclusions (names or `*` globs) |

Add guard rules to any tool that logs in or stores tokens.

## Removing and hiding

```sh
dws tools remove fzf httpie
```

- **Custom entries** are deleted.
- **Built-ins** are hidden by adding them to `removed:`.
- Removed names are also dropped from `defaults` and `favorites`.
- dws refuses to remove an entry another entry `requires` (for example `node` while `codex` needs it).

`dws tools remove` doesn't check templates. A template that still names a removed entry fails at `dws init` with `unknown tool "…"`.

To un-hide a built-in, delete it from `removed:`, or `dws tools add` it again.
