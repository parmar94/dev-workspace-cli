## dws creds github

Log gh in with a fine-grained, expiring personal access token

### Synopsis

Log gh in with a fine-grained personal access token and set it up as git's
credential helper. Classic tokens and tokens without an expiry are refused.

With --create, dws opens GitHub's "new token" page pre-filled for this
project: name, expiry, the repository owner and a permission preset. GitHub
can't pre-select repositories, so choose "Only select repositories" and pick
this one, generate the token, and paste it at the prompt.

Access presets: read (contents:read), push (contents:write),
pr (contents:write + pull_requests:write). Add more with --perm.

```
dws creds github [dir] [flags]
```

### Examples

```
  dws creds github --create                 # guided: opens a pre-filled token page
  dws creds github --create --access push --days 7 --perm issues=write
  dws creds github                          # paste an existing token
  pbpaste | dws creds github --token-stdin
```

### Options

```
      --access string      with --create: permission preset: read, push or pr (default "pr")
      --create             open GitHub's token page pre-filled for this project first
      --days int           with --create: days until the token expires (1-366) (default 30)
  -h, --help               help for github
      --no-browser         with --create: print the URL instead of opening a browser
      --perm stringArray   with --create: extra permission name=read|write|admin (repeatable)
      --repo string        with --create: owner/name (default: from the git remote "origin")
      --token-stdin        read the token from stdin
```

### Options inherited from parent commands

```
      --config string   global config file (default ~/.config/dws/config.yaml, or $DWS_CONFIG)
      --user string     container user (default: remote_user from .devcontainer/dws.yaml)
```

### SEE ALSO

* [dws creds](dws_creds.md)	 - Pass short-lived credentials into a running dev container

