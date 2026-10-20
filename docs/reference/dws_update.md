## dws update

Change a project's dev container and regenerate it

### Synopsis

Change the languages, agents, tools or options of a project generated with
dws init, and regenerate .devcontainer/ from its dws.yaml.

Start from the recorded selection and adjust it with --add-*/--remove-*, or
replace a whole list with --lang/--agent/--tool. With no changes on a
terminal, you're prompted with the current selection. With no changes and
--yes, files are regenerated with the current registry (useful after
upgrading dws or editing your registry).

Files you edited by hand since dws generated them are not overwritten unless
you pass --force.

```
dws update [dir] [flags]
```

### Examples

```
  dws update --add-tool ripgrep,terraform --remove-agent codex
  dws update --add-lang python --dry-run
  dws update --yes --rebuild          # regenerate with the current registry, then rebuild
```

### Options

```
      --add-agent strings      add agents
      --add-lang strings       add languages
      --add-tool strings       add tools
  -a, --agent strings          AI agents, comma-separated (claude, codex)
      --agent-guard            deny agents access to credential files and commands of selected tools (use --agent-guard=false to skip) (default true)
      --agent-sandbox          turn on Claude Code's OS-level sandbox for its shell commands (blocks credential reads, limits network)
      --base-image string      base image (default from config)
      --dry-run                show what would change without writing
      --force                  overwrite generated files you edited by hand
      --git-config             mount the host ~/.gitconfig and include it
  -h, --help                   help for update
  -l, --lang strings           languages, comma-separated
      --name string            container name (default: directory name)
      --no-cache               with --rebuild, skip Docker's build cache
  -p, --port ints              ports to forward, comma-separated
      --rebuild                rebuild the container after updating
      --remove-agent strings   remove agents
      --remove-lang strings    remove languages
      --remove-tool strings    remove tools
      --ssh-agent              forward the host SSH agent
      --ssh-mode string        how to forward the SSH agent: mac or linux (default: detect host)
      --tool strings           extra tools, comma-separated
  -y, --yes                    don't prompt
```

### Options inherited from parent commands

```
      --config string   global config file (default ~/.config/dws/config.yaml, or $DWS_CONFIG)
```

### SEE ALSO

* [dws](dws.md)	 - Set up and manage project workspaces in dev containers

