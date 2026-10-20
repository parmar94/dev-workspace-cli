package creds

import (
	"context"
	"net/url"
	"strings"
	"testing"
)

func TestParseGitHubRepo(t *testing.T) {
	for in, want := range map[string]string{
		"git@github.com:parmar94/dev-workspace-cli.git":       "parmar94/dev-workspace-cli",
		"https://github.com/parmar94/dev-workspace-cli.git":   "parmar94/dev-workspace-cli",
		"https://github.com/parmar94/dev-workspace-cli":       "parmar94/dev-workspace-cli",
		"ssh://git@github.com/acme/api.git":                   "acme/api",
		"https://x-access-token@github.com/acme/web.site.git": "acme/web.site",
		"acme/api": "acme/api",
	} {
		o, r, err := ParseGitHubRepo(in)
		if err != nil || o+"/"+r != want {
			t.Errorf("%q -> %s/%s, %v; want %s", in, o, r, err, want)
		}
	}
	for _, bad := range []string{"git@gitlab.com:a/b.git", "justname", "", "a/"} {
		if _, _, err := ParseGitHubRepo(bad); err == nil {
			t.Errorf("%q: want error", bad)
		}
	}
}

func TestGitHubRepoFromRemote(t *testing.T) {
	r := &fakeRunner{answers: map[string]string{"git -C /p remote get-url origin": "git@github.com:acme/api.git\n"}}
	o, n, err := GitHubRepo(context.Background(), r, "/p")
	if err != nil || o != "acme" || n != "api" {
		t.Errorf("got %s/%s %v", o, n, err)
	}
	_, _, err = GitHubRepo(context.Background(), &fakeRunner{fail: map[string]bool{"git": true}}, "/p")
	if err == nil || !strings.Contains(err.Error(), "--repo") {
		t.Errorf("err = %v", err)
	}
}

func TestPATTemplateURL(t *testing.T) {
	tmpl, err := NewPATTemplate("acme", "api", "pr", 14, map[string]string{"issues": "write"})
	if err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(tmpl.URL())
	if err != nil {
		t.Fatal(err)
	}
	if u.Host != "github.com" || u.Path != "/settings/personal-access-tokens/new" {
		t.Errorf("url = %s", u)
	}
	q := u.Query()
	for k, want := range map[string]string{
		"name": "dws api", "target_name": "acme", "expires_in": "14",
		"contents": "write", "pull_requests": "write", "issues": "write",
	} {
		if q.Get(k) != want {
			t.Errorf("%s = %q, want %q", k, q.Get(k), want)
		}
	}
	if !strings.Contains(q.Get("description"), "acme/api") {
		t.Errorf("description = %q", q.Get("description"))
	}
	if tmpl.URL() != tmpl.URL() {
		t.Error("URL not deterministic")
	}

	long, _ := NewPATTemplate("o", strings.Repeat("r", 60), "read", 1, nil)
	if len(long.Name) > 40 {
		t.Errorf("name too long: %d", len(long.Name))
	}

	for _, tc := range []struct {
		access string
		days   int
		extra  map[string]string
	}{
		{"admin", 30, nil},
		{"pr", 0, nil},
		{"pr", 367, nil},
		{"pr", 30, map[string]string{"contents": "owner"}},
		{"pr", 30, map[string]string{"Bad-Name": "read"}},
	} {
		if _, err := NewPATTemplate("o", "r", tc.access, tc.days, tc.extra); err == nil {
			t.Errorf("%+v: want error", tc)
		}
	}
}
