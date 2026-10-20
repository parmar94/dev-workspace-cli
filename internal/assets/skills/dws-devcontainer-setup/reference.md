# dws init reference

## Flags

| Flag | Meaning |
| --- | --- |
| `-t, --template <name>` | start from a template (`dws templates list --json`) |
| `-l, --lang a,b` | languages (replaces the default/template list) |
| `-a, --agent a,b` | AI agents: `claude`, `codex` |
| `--tool a,b` | extra tools |
| `--ssh-agent` / `--ssh-agent=false` | forward the host SSH agent |
| `--ssh-mode mac\|linux` | override host detection |
| `--git-config` | mount host `~/.gitconfig` (VS Code already copies it) |
| `--agent-guard=false` | skip agent deny rules (on by default) |
| `-p, --port 8080,3000` | forwardPorts |
| `--name`, `--base-image` | container name (default: directory name), image |
| `-y, --yes` | no prompts; **always pass this** |
| `--dry-run` | print files, write nothing |
| `--force` | overwrite an existing `.devcontainer/` (needs the user's OK) |

Merge order: global config `defaults` < `--template` < flags.

## Detecting languages

| Signal | Language |
| --- | --- |
| `go.mod`, `go.work` | go |
| `package.json`, `pnpm-lock.yaml`, `yarn.lock`, `bun.lockb` | node |
| `pyproject.toml`, `requirements*.txt`, `Pipfile`, `setup.py`, `uv.lock` | python |
| `Cargo.toml` | rust |
| `pom.xml`, `build.gradle`, `build.gradle.kts` | java |

In a monorepo, take the union across packages. Ignore vendored or generated folders (`node_modules`, `vendor`, `dist`).

## Detecting tools

| Signal | Tool |
| --- | --- |
| `.github/`, user works with PRs and issues | gh |
| `*.tf`, `.terraform.lock.hcl` | terraform |
| `Chart.yaml`, `kustomization.yaml`, k8s manifests | kubectl |
| project builds or runs containers | docker-in-docker |
| `aws-sdk`, `boto3`, `cdk.json`, `serverless.yml`, SAM templates | aws-cli |
| JSON-heavy scripts | jq |

Dependencies are added automatically (for example, codex pulls in node). Don't add them by hand unless the project needs them anyway.

## Missing registry entries

If the project needs something that isn't in `dws tools list --json`:

1. Look for an official Dev Container Feature on https://containers.dev/features (prefer `ghcr.io/devcontainers/features/*`). Otherwise use an install script that runs as the container user (it may use `sudo apt-get …`).
2. Show the user the exact command and ask before running it, because it changes their global config and drops comments from the config file:
   ```sh
   dws tools add deno --kind language --feature ghcr.io/devcontainers-extra/features/deno:1 -d "Deno runtime"
   dws tools add httpie --script "sudo apt-get update && sudo apt-get install -y httpie" -d "HTTP client"
   ```
3. If the tool stores credentials, add guard rules so agents are denied access:
   `--guard-path "~/.mytool/**" --guard-cmd "mytool auth token" --guard-env "MYTOOL_TOKEN"`.

## Generated files

```
.devcontainer/devcontainer.json
.devcontainer/scripts/post-create.sh, <name>.sh   # script-based installs, volume ownership, guard install
.devcontainer/agent-policy/*                       # Claude managed settings, Codex env policy
.devcontainer/dws.yaml                             # selections; never contains secrets
```
