---
name: dws-devcontainer-setup
description: Set up, change or rebuild a project's dev container with the dws CLI. Picks the languages, AI agents (Claude Code, Codex) and tools from the project's files and the user's preferences, previews the result, then writes .devcontainer/. dws update changes an existing setup and dws rebuild applies it. Use when the user asks to create, set up, containerize, update, add to, remove from or rebuild a project's dev environment or dev container. Not for editing the global dws config (use dws-config) or passing credentials into a container (use dws-credentials).
---

# Set up a dev container with dws

`dws` runs on the **host**, not inside a container. It generates `.devcontainer/` (devcontainer.json, install scripts, agent guard policy, and `dws.yaml`, which records the choices).

## Rules

- **Never run `dws init` without `--yes`.** Without it, dws opens interactive prompts that need a TTY.
- **Always run `--dry-run` first**, show the user a summary, and write only once they agree.
- **Never pass `--force`** unless the user explicitly agrees to overwrite an existing `.devcontainer/` or their hand edits.
- **For projects that already have `.devcontainer/dws.yaml`, use `dws update`**, not `dws init`.
- **Don't commit** the generated files unless asked.
- **Don't add or remove registry entries without asking.** `dws tools add/remove` change the user's global config.

## Steps

1. **Check dws is available.** Run `command -v dws`. If it's missing, offer `go install github.com/parmar94/dev-workspace-cli/cmd/dws@latest`.

2. **Read the user's preferences and the catalog:**
   ```sh
   dws config show          # defaults, favorites, custom templates (may not exist yet)
   dws templates list --json
   dws tools list --json    # kind, name, install, requires, favorite, guard
   ```
   Prefer the user's defaults and favorites. If no config exists, mention `dws config init`. Don't create the config yourself.

3. **Work out what the project needs.** Check the root and obvious subfolders for these files. The full mapping is in [reference.md](reference.md).
   - `go.mod` → go
   - `package.json` → node
   - `pyproject.toml`, `requirements*.txt` → python
   - `Cargo.toml` → rust
   - `pom.xml`, `build.gradle*` → java

   Also look for tool signals:
   - `.github/` → gh
   - `*.tf` → terraform
   - k8s manifests or a Helm `Chart.yaml` → kubectl
   - `Dockerfile` or `docker-compose*.yml` that the project builds → docker-in-docker
   - AWS SDK or CDK usage → aws-cli

   Check for an existing `.devcontainer/`. If one exists and has `dws.yaml`, read it as the current selection.

4. **Ask only what you can't infer**, in one short message:
   - Which agents: claude, codex, both, or none. Default to the config's `defaults.agents`.
   - Extra tools beyond those detected.
   - SSH agent forwarding, for private git over SSH.
   - If Claude is selected: the opt-in Claude sandbox (`--agent-sandbox`). It's stronger credential protection, but new network hosts prompt and `docker` can't run inside it. Respect `defaults.agent_sandbox`.

   If a built-in template matches closely, suggest it.

5. **Preview:**
   ```sh
   dws init <dir> --yes --dry-run --lang go,node --agent claude --tool gh,jq --ssh-agent
   # or start from a template, then override lists:
   dws init <dir> --yes --dry-run --template go-web --tool gh,jq,ripgrep
   ```
   A list flag *replaces* the template's or default's list rather than adding to it, so pass the full list. Summarize for the user:
   - the languages, agents and tools
   - any automatically added dependencies (stderr says "Also adding required: …")
   - the mounts
   - whether the agent guard is on

6. **Write it.** Run the same command without `--dry-run`. If a needed language or tool isn't in the registry, see [reference.md](reference.md#missing-registry-entries).

7. **Next steps for the user:**
   - Start the container: VS Code "Reopen in Container", `devcontainer up --workspace-folder <dir>`, or `dws rebuild <dir>`.
   - For cloud or GitHub credentials, they should run `dws creds …` themselves (see dws-credentials).
   - In the container, `ssh-add -l` checks SSH forwarding.

## Changing an existing setup

`dws update` starts from `.devcontainer/dws.yaml`:

```sh
dws update <dir> --dry-run --add-tool terraform --remove-agent codex   # preview; summary on stderr
dws update <dir> --add-tool terraform --remove-agent codex             # after the user agrees
```

- Use `--add-lang/--remove-lang`, `--add-agent/--remove-agent` and `--add-tool/--remove-tool`. `--lang/--agent/--tool` replace a whole list.
- Without change flags, `dws update` prompts on a terminal, so pass `--yes` (it regenerates with the current registry).
- **If it reports files "edited by hand", stop and show the user.** Use `--force` only with their OK, because it discards those edits.
- **Applying a change needs a rebuild:** `dws rebuild <dir>` (or `--rebuild` on update), which replaces the container. Ask first if the container may be in use. It keeps agent logins, but credentials passed with `dws creds` must be passed in again.
