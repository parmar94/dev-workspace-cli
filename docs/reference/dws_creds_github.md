## dws creds github

Log gh in with a fine-grained, expiring personal access token

### Synopsis

Log gh in with a fine-grained personal access token and set it up as git's
credential helper. Classic tokens and tokens without an expiry are refused.

Create one at https://github.com/settings/personal-access-tokens/new limited
to the repositories and permissions this project needs.

```
dws creds github [dir] [flags]
```

### Examples

```
  dws creds github
  pbpaste | dws creds github --token-stdin
```

### Options

```
  -h, --help          help for github
      --token-stdin   read the token from stdin
```

### Options inherited from parent commands

```
      --config string   global config file (default ~/.config/dws/config.yaml, or $DWS_CONFIG)
      --user string     container user (default: remote_user from .devcontainer/dws.yaml)
```

### SEE ALSO

* [dws creds](dws_creds.md)	 - Pass short-lived credentials into a running dev container

