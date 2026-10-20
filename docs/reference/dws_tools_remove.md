## dws tools remove

Remove entries (built-in ones are hidden)

### Synopsis

Remove entries from the registry. Built-in entries are hidden via `removed` in the config.

Note: saving rewrites the config file, so comments in it are not kept.

```
dws tools remove <name>... [flags]
```

### Options

```
  -h, --help   help for remove
```

### Options inherited from parent commands

```
      --config string   global config file (default ~/.config/dws/config.yaml, or $DWS_CONFIG)
```

### SEE ALSO

* [dws tools](dws_tools.md)	 - List and manage the languages, agents and tools in your registry

