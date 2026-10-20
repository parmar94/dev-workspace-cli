# AI agents

dws works with AI coding agents in two ways:

1. **Agents inside containers:** install Claude Code or Codex in a project's dev container.
2. **Agents that drive dws:** let Claude Code or Codex on your host set up dev containers for you, using bundled skills.

## Agents inside containers

```sh
dws init --agent claude,codex --yes
```

| Agent | Installed with | Login stored in |
| --- | --- | --- |
| `claude` (Claude Code) | Anthropic's dev container feature | `dws-claude` volume → `~/.claude` |
| `codex` (OpenAI Codex CLI) | `npm install -g @openai/codex` (adds `node` automatically) | `dws-codex` volume → `~/.codex` |

The volumes are shared by every dws container, so you log in once and stay logged in across rebuilds and projects. To start fresh, remove the volume: `docker volume rm dws-claude`.

The [agent guard](credentials.md#agent-guard) is on by default and keeps agents away from the credentials of the other tools you selected. For Claude Code you can also turn on the [OS-level sandbox](credentials.md#claude-code-sandbox-opt-in) with `--agent-sandbox`.

To add another agent, see [Adding tools](registry.md); the `gemini` example there works the same way.

## Let an agent set up projects for you

dws bundles [agent skills](https://code.claude.com/docs/en/skills): instructions that teach an agent how to use dws safely. Install them on your host:

```sh
dws skills install                                     # Claude Code + Codex, all your projects
dws skills install --agent claude                      # only Claude Code
dws skills install --scope project --dir ~/code/api    # only one repo (commit them to share)
dws skills install --force                             # update after upgrading dws
```

| Agent | User scope | Project scope |
| --- | --- | --- |
| Claude Code | `~/.claude/skills/` | `<dir>/.claude/skills/` |
| Codex | `~/.agents/skills/` | `<dir>/.agents/skills/` |

### Bundled skills

| Skill | Ask things like |
| --- | --- |
| `dws-devcontainer-setup` | "set up a dev container for this repo", "add Python and Terraform to the dev container", "rebuild the dev container" |
| `dws-config` | "make Codex one of my default agents", "save my usual Go setup as a template", "add httpie to dws" |
| `dws-credentials` | "how do I use AWS in the container?", "set up GitHub access for the container", "should I turn on the Claude sandbox?" |

### What the agent will do

For `dws-devcontainer-setup`, the agent:
1. Reads your config, templates and registry (`dws config show`, `dws … list --json`).
2. Looks at the project (`go.mod`, `package.json`, `pyproject.toml`, `*.tf`, `.github/`, …) to infer languages and tools.
3. Asks only what it can't infer, usually which agents to install and whether to forward SSH.
4. Runs `dws init … --yes --dry-run`, shows you a summary, and writes files once you agree.
5. Tells you how to open the container.

The skills tell agents to:
- **never** use `--force` or change your global config without asking
- **never** ask you to paste secrets into the chat
- leave `dws creds` for you to run yourself

Run these agents on your host, where dws and Docker are. An agent inside a dev container can't create or rebuild its own container.
