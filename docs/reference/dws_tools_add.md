## dws tools add

Add an entry to the registry, or override a built-in one

### Synopsis

Add an entry to the registry, or override a built-in one.

Give either --feature (a dev container feature reference) or --script /
--script-file (shell run as the remote user after the container is created).
In --mount, {{home}} expands to the remote user's home directory.

Note: saving rewrites the config file, so comments in it are not kept.

```
dws tools add <name> [flags]
```

### Examples

```
  dws tools add httpie --script "sudo apt-get update && sudo apt-get install -y httpie"
  dws tools add deno --kind language --feature ghcr.io/devcontainers-extra/features/deno:1
  dws tools add gemini --kind agent --requires node --script "npm i -g @google/gemini-cli" --favorite
```

### Options

```
  -d, --description string       short description
      --domain stringArray       network host this entry needs, pre-allowed in the Claude sandbox (repeatable)
      --env stringArray          container env KEY=value (repeatable)
      --favorite                 also add to favorites
      --feature string           dev container feature reference
      --guard-cmd stringArray    command prefix agents may not run (repeatable)
      --guard-env stringArray    env var name or glob hidden from agent shells (repeatable)
      --guard-path stringArray   credential path glob agents may not read, e.g. ~/.mytool/** (repeatable)
  -h, --help                     help for add
  -k, --kind string              language, agent or tool (default "tool")
      --mount stringArray        mount spec (repeatable)
      --option stringArray       feature option key=value (repeatable)
      --requires strings         entries this one needs, comma-separated
      --script string            install script
      --script-file string       read the install script from a file
```

### Options inherited from parent commands

```
      --config string   global config file (default ~/.config/dws/config.yaml, or $DWS_CONFIG)
```

### SEE ALSO

* [dws tools](dws_tools.md)	 - List and manage the languages, agents and tools in your registry

