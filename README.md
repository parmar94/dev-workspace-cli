# dev-workspace-cli

A Go CLI for setting up and managing project workspaces in dev containers.

> **Status:** early stage, personal use. Nothing is implemented yet. It may be open-sourced later.

## Goals

- **Per-project workspaces:** generate a dev container for a project, with the language(s) and tools it needs.
- **AI agents:** configure the AI coding agents you want available inside the container.
- **Global config:** keep your defaults, favourite setups and commonly used tools in one place, and reuse them across projects.
- **Update and rebuild:** change a project's dev container config and rebuild the container from the CLI.

## Planned workflow

1. Set your defaults in the global config.
2. Run the CLI in a project to create its dev container, starting from those defaults.
3. Adjust languages, tools or agents later, then rebuild.

## Development

Work happens on feature branches off `main`. Design and plans will be added as features are scoped.

## License

[MIT](LICENSE) © 2026 Ashish Jain
