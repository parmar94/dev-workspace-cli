## dws rebuild

Rebuild and restart a project's dev container

### Synopsis

Rebuild a project's dev container with the devcontainer CLI, replacing the
existing container. Uses "devcontainer" from PATH, or "npx @devcontainers/cli".

Named volumes (agent logins) are kept. Credentials passed with dws creds live
in the old container and are gone afterwards; pass them in again.

If the folder is open in VS Code, you can use "Dev Containers: Rebuild
Container" there instead.

```
dws rebuild [dir] [flags]
```

### Examples

```
  dws rebuild
  dws rebuild ./api --no-cache
```

### Options

```
  -h, --help       help for rebuild
      --no-cache   rebuild the image without Docker's build cache
```

### Options inherited from parent commands

```
      --config string   global config file (default ~/.config/dws/config.yaml, or $DWS_CONFIG)
```

### SEE ALSO

* [dws](dws.md)	 - Set up and manage project workspaces in dev containers

