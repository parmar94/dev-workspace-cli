---
name: dws-config
description: Manage the user's global dws configuration, which covers default languages, agents and tools, favorites, custom registry entries (new languages, tools or agents with install features or scripts and credential guard rules), hidden built-ins, and personal templates. Use when the user wants to change dws defaults, save a favorite setup as a template, add or remove a tool from dws, or asks what dws can install. Not for generating a project's dev container (use dws-devcontainer-setup).
---

# Manage dws global config

The config is YAML at `dws config path`: `$DWS_CONFIG`, else `$XDG_CONFIG_HOME/dws/config.yaml`, else `~/.config/dws/config.yaml`. It belongs to the user. Show proposed changes and get a yes before writing.

## Read first

```sh
dws config path
dws config show            # raw file; says so if it doesn't exist
dws tools list --json      # effective registry: built-ins + custom − removed
dws templates list --json
```

If there's no config yet, offer `dws config init`. It writes a commented starter file.

## Make changes

| Goal | How |
| --- | --- |
| Change `defaults`, `favorites`, add a template | Edit the YAML file directly, which keeps the user's comments. Then run `dws tools list` to confirm it still parses and validates. |
| Add or override a registry entry | `dws tools add <name> --kind language\|agent\|tool (--feature <ref> [--option k=v] \| --script "…") [-d desc] [--requires a,b] [--mount spec] [--env K=V] [--guard-path p] [--guard-cmd c] [--guard-env E] [--favorite]` |
| Hide or remove an entry | `dws tools remove <name>…`. Built-ins are hidden through `removed:`. It refuses if another entry still `requires` it. |

`dws tools add/remove` rewrite the file, so **comments are lost**. Tell the user before running them, or edit the YAML by hand instead. Never use `dws config edit`, because it opens an interactive editor.

## Schema

```yaml
defaults:
  languages: [go]
  agents: [claude]          # claude, codex
  tools: [gh]
  ssh_agent: true
  base_image: mcr.microsoft.com/devcontainers/base:ubuntu
  remote_user: vscode       # mount targets use this user's home
favorites: [go, claude, gh] # starred and listed first in prompts
registry:                   # custom or overriding entries; names: a-z 0-9 . _ -
  tools:                    # or languages: / agents:
    httpie:
      description: HTTP client
      script: sudo apt-get update && sudo apt-get install -y httpie   # OR feature: <ref> + options: {}
      requires: []          # other entry names, installed first
      mounts: ["source=dws-x,target={{home}}/.x,type=volume"]
      env: {KEY: value}     # containerEnv; never secrets
      guard:                # denied to AI agents
        paths: ["~/.httpie/**"]
        commands: ["http --auth"]
        env: ["HTTPIE_*"]
removed: [fzf]              # hide built-ins
templates:
  my-go:
    description: My Go setup
    languages: [go]
    agents: [claude]
    tools: [gh, jq]
    ssh_agent: true         # optional
    base_image: …           # optional
```

## Rules

- Each entry needs **exactly one** of `feature` or `script`. Prefer official features from https://containers.dev/features, pinned to a major version (`:1`).
- **Never put secrets in config.** Not in `env`, not in scripts. Credentials go through `dws creds` (see dws-credentials).
- When adding a tool that stores credentials, propose `guard` rules for its credential files, token-printing commands and env vars.
- A template's lists **replace** the defaults rather than adding to them, so include everything the template needs.
