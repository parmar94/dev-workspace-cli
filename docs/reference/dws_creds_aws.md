## dws creds aws

Pass temporary AWS credentials from a host profile (SSO or assumed role)

```
dws creds aws [dir] [flags]
```

### Examples

```
  aws sso login --profile dev-readonly
  dws creds aws --profile dev-readonly --region eu-west-1
```

### Options

```
  -h, --help             help for aws
      --profile string   host AWS profile to export (required)
      --region string    default region to set in the container
```

### Options inherited from parent commands

```
      --config string   global config file (default ~/.config/dws/config.yaml, or $DWS_CONFIG)
      --user string     container user (default: remote_user from .devcontainer/dws.yaml)
```

### SEE ALSO

* [dws creds](dws_creds.md)	 - Pass short-lived credentials into a running dev container

