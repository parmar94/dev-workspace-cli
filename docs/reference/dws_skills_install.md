## dws skills install

Install bundled skills for Claude Code and/or Codex

### Synopsis

Install the bundled skills (default: all) so Claude Code and Codex can set up
dev containers, manage dws config and handle credentials for you.

  claude  user: ~/.claude/skills    project: <dir>/.claude/skills
  codex   user: ~/.agents/skills    project: <dir>/.agents/skills

Re-run with --force after upgrading dws to update installed skills.

```
dws skills install [skill]... [flags]
```

### Examples

```
  dws skills install
  dws skills install --agent claude --scope project --dir .
  dws skills install dws-credentials --force
```

### Options

```
  -a, --agent strings   agents to install for: claude, codex (default [claude,codex])
      --dir string      project directory for --scope project (default ".")
      --force           replace skills that are already installed
  -h, --help            help for install
      --scope string    user (all projects) or project (default "user")
```

### Options inherited from parent commands

```
      --config string   global config file (default ~/.config/dws/config.yaml, or $DWS_CONFIG)
```

### SEE ALSO

* [dws skills](dws_skills.md)	 - Install agent skills that teach Claude Code and Codex to use dws

