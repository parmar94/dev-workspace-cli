package creds

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// GitHubWeb is the GitHub web base URL; tests point it elsewhere.
var GitHubWeb = "https://github.com"

// AccessPresets are permission sets for common dev-container work. Metadata
// read access is always included by GitHub.
var AccessPresets = map[string]map[string]string{
	"read": {"contents": "read"},
	"push": {"contents": "write"},
	"pr":   {"contents": "write", "pull_requests": "write"},
}

// PATTemplate describes a fine-grained token to pre-fill on GitHub's
// "new token" page. GitHub can't pre-select repositories, so the user picks
// the repository on the page.
type PATTemplate struct {
	Name        string
	Description string
	Owner       string
	Days        int
	Permissions map[string]string
}

// NewPATTemplate builds a template for owner/repo using an access preset
// plus extra permissions.
func NewPATTemplate(owner, repo, access string, days int, extra map[string]string) (PATTemplate, error) {
	preset, ok := AccessPresets[access]
	if !ok {
		return PATTemplate{}, fmt.Errorf("unknown access preset %q (want read, push or pr)", access)
	}
	if days < 1 || days > 366 {
		return PATTemplate{}, fmt.Errorf("--days must be between 1 and 366, got %d", days)
	}
	perms := map[string]string{}
	for k, v := range preset {
		perms[k] = v
	}
	for k, v := range extra {
		if !validPermName.MatchString(k) {
			return PATTemplate{}, fmt.Errorf("invalid permission name %q", k)
		}
		if !slices.Contains([]string{"read", "write", "admin"}, v) {
			return PATTemplate{}, fmt.Errorf("permission %s: level must be read, write or admin, got %q", k, v)
		}
		perms[k] = v
	}
	name := "dws " + repo
	if len(name) > 40 {
		name = name[:40]
	}
	return PATTemplate{
		Name:        name,
		Description: fmt.Sprintf("dws dev container for %s/%s. Limit Repository access to this repository only.", owner, repo),
		Owner:       owner,
		Days:        days,
		Permissions: perms,
	}, nil
}

var validPermName = regexp.MustCompile(`^[a-z_]+$`)

// URL returns the pre-filled token creation URL. Parameters are in a fixed
// order so the URL is stable.
func (t PATTemplate) URL() string {
	q := []string{
		"name=" + url.QueryEscape(t.Name),
		"description=" + url.QueryEscape(t.Description),
		"target_name=" + url.QueryEscape(t.Owner),
		"expires_in=" + strconv.Itoa(t.Days),
	}
	keys := make([]string, 0, len(t.Permissions))
	for k := range t.Permissions {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	for _, k := range keys {
		q = append(q, k+"="+url.QueryEscape(t.Permissions[k]))
	}
	return GitHubWeb + "/settings/personal-access-tokens/new?" + strings.Join(q, "&")
}

var remotePattern = regexp.MustCompile(`github\.com[:/]([^/]+)/([^/]+?)(\.git)?/?$`)

// GitHubRepo returns the owner and repository of dir's "origin" remote.
func GitHubRepo(ctx context.Context, r Runner, dir string) (owner, repo string, err error) {
	out, err := r.Run(ctx, nil, "git", "-C", dir, "remote", "get-url", "origin")
	if err != nil {
		return "", "", fmt.Errorf("read the git remote (pass --repo owner/name instead): %w", err)
	}
	return ParseGitHubRepo(strings.TrimSpace(string(out)))
}

// ParseGitHubRepo accepts "owner/name" or a GitHub remote URL.
func ParseGitHubRepo(s string) (owner, repo string, err error) {
	if m := remotePattern.FindStringSubmatch(s); m != nil {
		return m[1], m[2], nil
	}
	if o, r, ok := strings.Cut(s, "/"); ok && o != "" && r != "" && !strings.ContainsAny(o+r, ":/ ") {
		return o, r, nil
	}
	return "", "", fmt.Errorf("%q is not a GitHub repository (want owner/name or a github.com remote)", s)
}
