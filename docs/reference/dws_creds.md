## dws creds

Pass short-lived credentials into a running dev container

### Synopsis

Pass short-lived credentials from the host into a running dev container.

Long-term secrets stay on the host. Credentials are streamed over stdin into
the container's filesystem (not the project or a volume), so a rebuild wipes
them.

### Options

```
  -h, --help          help for creds
      --user string   container user (default: remote_user from .devcontainer/dws.yaml)
```

### Options inherited from parent commands

```
      --config string   global config file (default ~/.config/dws/config.yaml, or $DWS_CONFIG)
```

### SEE ALSO

* [dws](dws.md)	 - Set up and manage project workspaces in dev containers
* [dws creds aws](dws_creds_aws.md)	 - Pass temporary AWS credentials from a host profile (SSO or assumed role)
* [dws creds github](dws_creds_github.md)	 - Log gh in with a fine-grained, expiring personal access token

