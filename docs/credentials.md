# Credentials and security

## The honest threat model

An AI agent running in your dev container runs **as the same user** as `aws`, `gh` and `git`. Anything those tools can read, the agent can read too: by reading the file, running the tool, writing a small script, or using `sudo`. No configuration inside one container fully hides a secret from an agent that shares it.

dws therefore aims for:

1. **No long-term secrets in containers.** They stay on your host. Containers get short-lived, narrowly scoped credentials, on demand. If one leaks, it's limited and expires soon.
2. **Guardrails against casual access.** Agent deny rules stop agents from reading credential files or printing tokens through the obvious routes. They're a speed bump, not a wall.
3. **Nothing secret in files.** `devcontainer.json`, `containerEnv`, `dws.yaml` and your global config get committed, shared or shown in `docker inspect`, so dws never writes secrets there, and you shouldn't either.

For anything high-stakes, such as production deploys, keep the credentials and the command on the host.

## SSH: agent forwarding

```sh
dws init --ssh-agent ...
```

Your private keys never enter the container. The container gets a socket to your host's SSH agent. Agents can *use* your keys to sign (`git push`) while the container runs, but can't copy them.

| Host | What dws configures |
| --- | --- |
| macOS (Docker Desktop, OrbStack) | mounts `/run/host-services/ssh-auth.sock` and hands it to the container user on every start (it's root-only by default) |
| Linux | bind-mounts your `$SSH_AUTH_SOCK` at `/ssh-agent` |

To check, run `ssh-add -l` in the container. "The agent has no identities" means forwarding works but your host agent has no keys loaded; run `ssh-add` on the host. VS Code forwards the agent by itself. `--ssh-agent` also makes it work with the devcontainer CLI and other editors.

## GitHub: fine-grained, expiring tokens

1. Create a token at <https://github.com/settings/personal-access-tokens/new>:
   - **Repository access:** only the repositories this project needs.
   - **Permissions:** the minimum, for example *Contents: read and write* and *Pull requests: read and write*.
   - **Expiration:** short, for example 7 or 30 days.
2. Start the container, then pass the token in from the host:
   ```sh
   dws creds github                       # hidden prompt
   pbpaste | dws creds github --token-stdin
   ```

dws checks the token with GitHub before passing it in:
- **Classic `ghp_` tokens are refused:** they can't be limited to specific repositories.
- **Tokens without an expiry are refused.**

It then runs `gh auth login --with-token` and `gh auth setup-git` in the container, so both `gh` and HTTPS `git push` work. This needs the `gh` tool in the container.

## AWS: temporary credentials from SSO or a role

Set up a profile for dev work on your host with a **least-privilege role**. Either SSO:

```ini
# ~/.aws/config on the host
[profile dev-readonly]
sso_session = my-sso
sso_account_id = 123456789012
sso_role_name = DeveloperReadOnly
region = eu-west-1
```

or assume-role (`role_arn` + `source_profile`). Then:

```sh
aws sso login --profile dev-readonly          # if SSO
dws creds aws --profile dev-readonly --region eu-west-1
# Installed AWS credentials for profile dev-readonly, expiring 2:30PM (in 59m0s).
```

- dws runs `aws configure export-credentials` on the host and writes the result as the `[default]` profile in the container's `~/.aws/credentials` (mode 600).
- **Profiles with long-term access keys are refused:** they have no expiry.
- Run it again when the credentials expire.

## Claude Code and Codex logins

Agents need their own credentials, so they can always read them. Keep those low-risk:
- Log in inside the container. The login lives in the `dws-claude` / `dws-codex` Docker volume, never in the project.
- If you use API keys, set spend limits on them.
- If you suspect a container was compromised, log out and back in, which rotates the token. Remove the volume if you need to: `docker volume rm dws-claude`.

## How `dws creds` handles secrets

- It finds the running container through the label the devcontainer CLI and VS Code set (`devcontainer.local_folder`).
- Secrets are streamed over **stdin** into `docker exec`. They never appear as command-line arguments (visible in `ps`) or environment variables.
- They're written to the container's home directory, not the project and not a volume, so **a rebuild wipes them**.
- It reads the container user from `.devcontainer/dws.yaml`; pass `--user` for containers dws didn't generate.

## Agent guard

On by default when you select an agent; turn it off with `dws init --agent-guard=false`.

Rules come from the `guard` of each selected tool (see [Adding tools](registry.md#guard-rules)) plus a base set:

| Base rule | Why |
| --- | --- |
| `~/.ssh/**` | private keys, if any were copied in |
| `**/.env`, `**/.env.*` | project secrets |
| `env`, `printenv` | dump every environment variable |

Built-in tools add, for example:
- **aws-cli:** `~/.aws/**`; `aws configure export-credentials`, `aws configure get`, `aws sts get-session-token`; `AWS_*`
- **gh:** `~/.config/gh/**`; `gh auth token`; `GH_TOKEN`, `GITHUB_TOKEN`
- **terraform, kubectl, docker-in-docker:** their credential files

How each agent enforces them:

| Agent | Installed as | Enforces |
| --- | --- | --- |
| Claude Code | `/etc/claude-code/managed-settings.d/dws-guard.json` (root-owned managed settings, which outrank user and project settings) | `Read(...)` denies for paths; `Bash(... *)` denies for commands |
| Codex | a marked `[shell_environment_policy]` block in `~/.codex/config.toml` | hides the listed env vars from commands Codex runs. Codex has no file-read deny rules. If you already have a `[shell_environment_policy]` table, dws leaves it alone and prints the block for you to merge. |

**Limits, plainly:**
- Claude Code's deny rules cover its file tools and file commands it recognizes (`cat`, `head`, …). They don't cover a script that opens files itself.
- An agent with `sudo` (the default container user has it) can remove the policy.
- For OS-level enforcement, look at Claude Code's [sandbox](https://code.claude.com/docs/en/sandboxing) `sandbox.credentials` settings. Inside Docker it requires `enableWeakerNestedSandbox`, and it adds network-approval prompts.

## Checklist

- [ ] `--ssh-agent` instead of copying keys
- [ ] fine-grained, repo-scoped, expiring GitHub tokens via `dws creds github`
- [ ] a dedicated least-privilege AWS role via `dws creds aws`
- [ ] agent guard on (`--agent-guard`, the default)
- [ ] nothing secret in `.devcontainer/`, `containerEnv` or the global config
- [ ] production credentials and deploys stay on the host
