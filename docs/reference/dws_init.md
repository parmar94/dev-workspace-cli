## dws init

Generate a .devcontainer for a project

### Synopsis

Generate a .devcontainer for a project (default: current directory).

Values come from, lowest to highest priority: global config defaults, the
--template, then flags. A template or a list flag (--lang, --agent, --tool)
replaces the list below it rather than adding to it.

```
dws init [dir] [flags]
```

### Examples

```
  dws init
  dws init --template go-web --yes
  dws init ./api --lang go --agent claude,codex --tool gh,jq --ssh-agent --yes --dry-run
```

### Options

```
  -a, --agent strings       AI agents, comma-separated (claude, codex)
      --agent-guard         deny agents access to credential files and commands of selected tools (use --agent-guard=false to skip) (default true)
      --base-image string   base image (default from config)
      --dry-run             print the files instead of writing them
      --force               overwrite an existing .devcontainer
      --git-config          mount the host ~/.gitconfig and include it
  -h, --help                help for init
  -l, --lang strings        languages, comma-separated
      --name string         container name (default: directory name)
  -p, --port ints           ports to forward, comma-separated
      --ssh-agent           forward the host SSH agent
      --ssh-mode string     how to forward the SSH agent: mac or linux (default: detect host)
  -t, --template name       start from a named template (see dws templates list)
      --tool strings        extra tools, comma-separated
  -y, --yes                 don't prompt; use defaults, template and flags
```

### Options inherited from parent commands

```
      --config string   global config file (default ~/.config/dws/config.yaml, or $DWS_CONFIG)
```

### SEE ALSO

* [dws](dws.md)	 - Set up and manage project workspaces in dev containers

