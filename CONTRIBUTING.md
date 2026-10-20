# Contributing

Thanks for your interest in dws! Bug reports, docs fixes, new registry entries and features are all welcome.

## Before you start

- **Bugs:** open an issue with the bug template, including `dws version`, your OS and Docker setup, and the exact command and output.
- **Features:** open an issue first so we can agree on the approach before you write code.
- **Security issues:** don't open a public issue. See [SECURITY.md](SECURITY.md).

## Development setup

You need Go (see the version in `go.mod`), Docker, and optionally the [devcontainer CLI](https://github.com/devcontainers/cli) for end-to-end checks.

```sh
git clone https://github.com/parmar94/dev-workspace-cli.git
cd dev-workspace-cli
make build test lint     # bin/dws, unit tests, go vet + gofmt
```

CI also runs [golangci-lint](https://golangci-lint.run) (`.golangci.yml`) and a GoReleaser snapshot build. To run golangci-lint locally:

```sh
docker run --rm -v "$PWD":/app -w /app golangci/golangci-lint:v2.15.0 golangci-lint run ./...
```

## Project layout

```
cmd/dws/                  main package
internal/cli/             cobra commands, prompts
internal/config/          global config
internal/registry/        catalog types, dependency resolution, guard rules
internal/devcontainer/    rendering .devcontainer/ (golden-file tests in testdata/)
internal/creds/           dws creds (docker exec, AWS, GitHub)
internal/assets/          embedded registry.yaml, templates, starter config, agent skills
docs/                     guides; docs/reference is generated
```

## Making changes

- **Tests:** add or update tests with every change.
  - Rendering changes need golden files: `go test ./internal/devcontainer -update`, then review the diff.
  - Commands that call Docker, git or other tools go through a swappable runner, so they can be tested without those tools.
- **Commands and flags:** run `make docs` to regenerate `docs/reference`. A test fails if it's stale. Also update the guides in `docs/` and, if agents use the command, the skills in `internal/assets/skills/`.
- **Registry entries:**
  - Prefer official [Dev Container Features](https://containers.dev/features) pinned to a major version.
  - Add `guard` rules for anything that stores credentials.
  - Add `domains` for package registries a language needs.
- **Security-sensitive code** (`internal/creds`, agent guard and sandbox): secrets must only travel over stdin, never in arguments, env vars or generated files. Explain your reasoning in the PR.
- Keep commits focused, with short imperative messages ("Add dws update").
- Add a line under `## [Unreleased]` in [CHANGELOG.md](CHANGELOG.md) for user-visible changes.

## Pull requests

1. Branch from `main`.
2. Make sure `make test lint` passes.
3. Open a PR using the template and describe how you tested it. For anything touching container behavior, include an end-to-end check with `devcontainer up`.

## Releases (maintainers)

1. Move `Unreleased` entries in `CHANGELOG.md` under a new version heading, and commit.
2. Tag and push: `git tag v0.2.0 && git push origin v0.2.0`.
3. The Release workflow runs GoReleaser and creates a **draft** GitHub release with binaries and checksums. Review it and publish.

By contributing you agree that your contributions are licensed under the [MIT License](LICENSE), and that you'll follow the [Code of Conduct](CODE_OF_CONDUCT.md).
