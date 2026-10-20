package creds

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// AWSCreds are temporary credentials as printed by
// `aws configure export-credentials --format process`.
type AWSCreds struct {
	AccessKeyID     string `json:"AccessKeyId"`
	SecretAccessKey string `json:"SecretAccessKey"`
	SessionToken    string `json:"SessionToken"`
	Expiration      string `json:"Expiration"`
}

// ExportAWS resolves profile on the host into temporary credentials. It
// refuses long-term keys, which have no session token or expiry.
func ExportAWS(ctx context.Context, r Runner, profile string) (AWSCreds, error) {
	out, err := r.Run(ctx, nil, "aws", "configure", "export-credentials", "--profile", profile, "--format", "process")
	if err != nil {
		return AWSCreds{}, fmt.Errorf("export credentials for profile %q (for SSO, run `aws sso login --profile %s` first): %w", profile, profile, err)
	}
	var c AWSCreds
	if err := json.Unmarshal(out, &c); err != nil {
		return AWSCreds{}, fmt.Errorf("parse aws output: %w", err)
	}
	if c.AccessKeyID == "" || c.SecretAccessKey == "" {
		return AWSCreds{}, fmt.Errorf("aws returned no credentials for profile %q", profile)
	}
	if c.SessionToken == "" || c.Expiration == "" {
		return AWSCreds{}, fmt.Errorf("profile %q uses long-term access keys; only temporary credentials (SSO or an assumed role) are passed into containers", profile)
	}
	return c, nil
}

// Expires parses the expiry time.
func (c AWSCreds) Expires() (time.Time, error) {
	return time.Parse(time.RFC3339, c.Expiration)
}

// InstallAWS writes c as the default profile in the container's ~/.aws.
func InstallAWS(ctx context.Context, ctr *Container, c AWSCreds, region string) error {
	var b strings.Builder
	b.WriteString("# Temporary credentials from `dws creds aws`.\n")
	fmt.Fprintf(&b, "# Expires %s.\n[default]\n", c.Expiration)
	fmt.Fprintf(&b, "aws_access_key_id = %s\n", c.AccessKeyID)
	fmt.Fprintf(&b, "aws_secret_access_key = %s\n", c.SecretAccessKey)
	fmt.Fprintf(&b, "aws_session_token = %s\n", c.SessionToken)
	if err := ctr.WriteHomeFile(ctx, ".aws/credentials", []byte(b.String())); err != nil {
		return err
	}
	if region == "" {
		return nil
	}
	return ctr.WriteHomeFile(ctx, ".aws/config", []byte("[default]\nregion = "+region+"\n"))
}
