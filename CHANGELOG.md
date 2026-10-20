# Changelog

All notable changes to this project are documented here. The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project uses [Semantic Versioning](https://semver.org/). Before 1.0, minor versions may include breaking changes.

## [Unreleased]

## [0.1.0] - 2026-10-20

First public release.

### Added

- **`dws init`:** generate a `.devcontainer/` for a project.
  - Languages, AI agents (Claude Code, Codex) and tools, chosen interactively or with flags.
  - Templates, `--yes` and `--dry-run`.
- **`dws update`:** change a generated project with `--add-*`/`--remove-*` or full lists, then regenerate. Hand-edited files are detected and protected.
- **`dws rebuild`:** rebuild a project's container with the devcontainer CLI.
- **Global config** (`dws config`): defaults, favorites, custom registry entries and templates.
- **Registry management** (`dws tools`): Dev Container Features or install scripts, with dependencies, mounts, env, guard rules and sandbox domains.
- **`dws templates`:** built-in `go`, `go-web`, `python` and `fullstack` templates, plus your own.
- **SSH agent forwarding** for macOS (Docker Desktop / OrbStack) and Linux.
- **`dws creds aws`:** temporary credentials from SSO or assume-role profiles. Long-term keys are refused.
- **`dws creds github`:** fine-grained, expiring tokens only. `--create` opens a pre-filled token page.
- **Agent guard:** Claude Code managed-settings deny rules and a Codex env policy for tool credentials.
- **Opt-in Claude Code sandbox** (`--agent-sandbox`): OS-level credential and network limits for Claude's shell commands.
- **Agent skills** (`dws skills install`) for Claude Code and Codex.
- **Machine-readable output:** `--json` on `tools list` and `templates list`.
- **Docs:** user guides and a generated command reference.

[Unreleased]: https://github.com/parmar94/dev-workspace-cli/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/parmar94/dev-workspace-cli/releases/tag/v0.1.0
