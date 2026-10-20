---
name: dws-credentials
description: Help the user give a dws dev container access to AWS, GitHub or SSH safely. Uses short-lived, narrowly scoped credentials passed with `dws creds` (including guided GitHub token creation), dws agent guard rules and the opt-in Claude Code sandbox, without the agent ever seeing a secret. Use when the user asks how to use AWS, GitHub, git push or SSH keys inside a dev container, how to keep secrets away from AI agents, about the Claude sandbox, or about `dws creds`. Not for general secret management outside dev containers.
---

# Credentials in dws dev containers

## Hard rules

- **Never ask the user to paste a secret into the chat.** Never read, print or copy credential files (`~/.aws/*`, `~/.config/gh/*`, `~/.ssh/*`, `.env`). Never run commands that print tokens (`gh auth token`, `aws configure export-credentials`, `printenv`).
- **The user runs `dws creds` in their own terminal.** Give them the exact command to run.
- **Never put secrets in generated or config files.** That includes `devcontainer.json`, `containerEnv`, `dws.yaml` and the global config.

## Threat model (explain briefly when relevant)

An AI agent in the container runs as the same user as the tools, so it can read whatever they read. dws doesn't pretend otherwise. Instead it:
1. keeps long-term secrets on the host and passes in only **short-lived, narrowly scoped** credentials, and
2. installs **agent guard** deny rules. These are a guardrail, not a security boundary.

## What to recommend

| Need | Recommend |
| --- | --- |
| git over SSH | Generate with `--ssh-agent`. Keys stay on the host and the agent can only *use* them while connected. Check with `ssh-add -l` in the container. If it shows "no identities", run `ssh-add` on the host. |
| GitHub API / `gh` / HTTPS push | A fine-grained PAT limited to this repo, with minimal permissions and an expiry. **The user runs** `dws creds github --create [--access read\|push\|pr] [--days N] [--perm issues=write]`, which opens a pre-filled token page. They pick "Only select repositories" → this repo, generate the token, and paste it at dws's hidden prompt. With an existing token: `dws creds github`. dws refuses classic `ghp_` tokens and tokens without an expiry. You may run `dws creds github --create --no-browser` yourself to show the user the pre-filled URL. It contains no secret and stops before asking for the token. |
| AWS | A dedicated least-privilege role through SSO or assume-role. **The user runs** `aws sso login --profile <p>`, then `dws creds aws --profile <p> [--region r]`. dws refuses long-term access keys and prints the expiry. Re-run when it expires. |
| Claude Code / Codex login | Log in inside the container once. The login lives in the `dws-claude` / `dws-codex` volume, shared across projects. An agent always has access to its own login. Prefer API keys with spend limits if using keys. |

`dws creds` needs the container to be **running**. It streams the secret over stdin into the container's home directory, so a rebuild wipes it. It reads the container user from `.devcontainer/dws.yaml`; pass `--user` if there's no `dws.yaml`.

## Claude Code sandbox (opt-in)

`dws init … --agent-sandbox`, or `dws update --agent-sandbox` then `dws rebuild`, turns on Claude Code's OS-level sandbox for Claude's shell commands:
- blocks reads of the guard paths, including from scripts
- removes exact-name credential env vars
- allows network access only to GitHub plus each selected entry's `domains`; other hosts prompt

**Recommend it** when the user worries about agents reading credentials and can accept the trade-offs:
- new hosts prompt for approval
- `docker` can't run inside the sandbox
- Debian/Ubuntu images only

Check it in the container with `/sandbox`.

## Agent guard

- On by default when agents are selected (`dws init … --agent-guard=false` turns it off).
- **Claude Code:** `/etc/claude-code/managed-settings.d/dws-guard.json` denies `Read(...)` on credential paths and `Bash(... *)` for token commands.
- **Codex:** a `[shell_environment_policy]` block in `~/.codex/config.toml` hides credential env vars. Codex has no read-deny rules.
- Rules come from each selected tool's `guard` in the registry, plus a base set (`~/.ssh`, `.env` files, `env`/`printenv`). For a custom tool, add guard rules with `dws tools add … --guard-path/--guard-cmd/--guard-env` (see dws-config).
- **Limits:** scripts that open files themselves bypass Claude's deny rules, and an agent with `sudo` can remove the policy. For OS-level blocking with Claude, use the sandbox above. For stronger isolation, keep credentialed work (deploys, prod access) on the host.
