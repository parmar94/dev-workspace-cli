# Security policy

dws handles credentials (SSH agent sockets, AWS and GitHub tokens) and generates policies that limit what AI agents can do in containers, so security reports are very welcome.

## Reporting a vulnerability

**Please don't open a public issue.** Report it privately through GitHub: **Security → Report a vulnerability** on this repository ([private vulnerability reporting](https://docs.github.com/en/code-security/security-advisories/guidance-on-reporting-and-writing-information-about-vulnerabilities/privately-reporting-a-security-vulnerability)).

Include:
- the dws version (`dws version`), OS and Docker setup
- steps to reproduce, and what an attacker gains
- any suggested fix

You can expect an acknowledgement within a week. Fixes are released as soon as practical and credited in the advisory unless you prefer otherwise.

## Supported versions

Only the latest release gets security fixes.

## Scope

In scope, for example:
- a secret reaching a command line, environment variable, log, generated file or the project directory
- `dws creds` accepting credentials it should refuse (long-term AWS keys, classic or non-expiring GitHub tokens)
- generated agent guard or sandbox policies that don't apply rules they claim to apply
- path traversal or command injection through config, registry entries or flags

Out of scope (documented limitations, see [docs/credentials.md](docs/credentials.md)):
- an AI agent running as the same user in the same container reading credentials that tools in that container can read
- agents with `sudo` removing the agent policy
- deny rules being bypassed by scripts. The opt-in sandbox exists for this.
- vulnerabilities in Docker, Dev Container Features, Claude Code, Codex or other third-party tools. Report those upstream.
