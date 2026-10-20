# dws documentation

`dws` generates and manages [dev containers](https://containers.dev) for your projects. You choose the languages, AI agents and tools once in a global config, and `dws` turns them into a `.devcontainer/` for each project.

## Guides

| Guide | Read it when you want to… |
| --- | --- |
| [Getting started](getting-started.md) | install dws and set up your first project |
| [Configuration](configuration.md) | set your defaults, favorites and templates |
| [Adding tools](registry.md) | install something dws doesn't ship with |
| [Credentials and security](credentials.md) | use AWS, GitHub or SSH in a container without handing secrets to agents |
| [AI agents](agents.md) | run Claude Code or Codex in containers, or let them set up projects for you |
| [Troubleshooting](troubleshooting.md) | fix a specific error |

## Command reference

Generated from the CLI. Run `make docs` after changing commands or flags.

- [`dws init`](reference/dws_init.md): generate a `.devcontainer/`
- [`dws config`](reference/dws_config.md): manage the global config
- [`dws tools`](reference/dws_tools.md): list, add and remove languages, agents and tools
- [`dws templates`](reference/dws_templates.md): list and show templates
- [`dws creds`](reference/dws_creds.md): pass short-lived credentials into a running container
- [`dws skills`](reference/dws_skills.md): install agent skills
- [All commands](reference/dws.md)

## Concepts at a glance

- **Registry:** everything dws can install. Each entry is a *language*, *agent* or *tool*, installed either through a [Dev Container Feature](https://containers.dev/features) or a small shell script. Built-in entries ship with dws, and you can add, override or hide entries in your config.
- **Template:** a named preset of languages, agents and tools, for example `go-web`.
- **Global config:** your defaults, favorites, custom registry entries and templates, at `~/.config/dws/config.yaml`.
- **`dws.yaml`:** written next to `devcontainer.json`. It records what a project was generated with, and never contains secrets.
- **Agent guard:** deny rules that keep AI agents away from the credentials of tools in the same container. A guardrail, not a security boundary.
