package creds

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type call struct {
	name  string
	args  []string
	stdin string
}

// fakeRunner records calls and answers from a table keyed by command prefix.
type fakeRunner struct {
	calls   []call
	answers map[string]string
	fail    map[string]bool
}

func (f *fakeRunner) Run(_ context.Context, stdin io.Reader, name string, args ...string) ([]byte, error) {
	c := call{name: name, args: args}
	if stdin != nil {
		b, _ := io.ReadAll(stdin)
		c.stdin = string(b)
	}
	f.calls = append(f.calls, c)
	line := name + " " + strings.Join(args, " ")
	for prefix := range f.fail {
		if strings.HasPrefix(line, prefix) {
			return nil, errors.New("failed")
		}
	}
	for prefix, out := range f.answers {
		if strings.HasPrefix(line, prefix) {
			return []byte(out), nil
		}
	}
	return nil, nil
}

// assertNoSecretInArgs fails if secret appears in any command argument.
func (f *fakeRunner) assertNoSecretInArgs(t *testing.T, secret string) {
	t.Helper()
	for _, c := range f.calls {
		for _, a := range c.args {
			if strings.Contains(a, secret) {
				t.Errorf("secret passed as argument to %s: %q", c.name, c.args)
			}
		}
	}
}

func TestFindContainer(t *testing.T) {
	r := &fakeRunner{answers: map[string]string{"docker ps": "abc123\n"}}
	c, err := FindContainer(context.Background(), r, t.TempDir(), "vscode")
	if err != nil || c.ID != "abc123" {
		t.Fatalf("got %+v, %v", c, err)
	}
	if !strings.HasPrefix(r.calls[0].args[3], "label=devcontainer.local_folder=/") {
		t.Errorf("filter = %q", r.calls[0].args[3])
	}

	_, err = FindContainer(context.Background(), &fakeRunner{}, t.TempDir(), "vscode")
	if err == nil || !strings.Contains(err.Error(), "start it first") {
		t.Errorf("err = %v", err)
	}
}

func TestAWS(t *testing.T) {
	const temp = `{"Version":1,"AccessKeyId":"ASIAX","SecretAccessKey":"s3cr3t","SessionToken":"tok3n","Expiration":"2026-10-20T08:00:00+00:00"}`
	r := &fakeRunner{answers: map[string]string{"aws configure export-credentials": temp}}
	c, err := ExportAWS(context.Background(), r, "dev")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.Expires(); err != nil {
		t.Errorf("expiry: %v", err)
	}
	ctr := &Container{ID: "abc", User: "vscode", r: r}
	if err := InstallAWS(context.Background(), ctr, c, "eu-west-1"); err != nil {
		t.Fatal(err)
	}
	r.assertNoSecretInArgs(t, "s3cr3t")
	r.assertNoSecretInArgs(t, "tok3n")
	written := r.calls[1]
	if !strings.Contains(written.stdin, "aws_session_token = tok3n") || written.args[len(written.args)-1] != ".aws/credentials" {
		t.Errorf("credentials write = %+v", written)
	}
	if got := strings.Join(written.args[:5], " "); got != "exec -u vscode -i abc" {
		t.Errorf("exec args = %q", got)
	}
	if !strings.Contains(r.calls[2].stdin, "region = eu-west-1") {
		t.Errorf("config write = %+v", r.calls[2])
	}

	static := `{"Version":1,"AccessKeyId":"AKIAX","SecretAccessKey":"s"}`
	_, err = ExportAWS(context.Background(), &fakeRunner{answers: map[string]string{"aws": static}}, "prod")
	if err == nil || !strings.Contains(err.Error(), "long-term access keys") {
		t.Errorf("static keys: err = %v", err)
	}
	_, err = ExportAWS(context.Background(), &fakeRunner{fail: map[string]bool{"aws": true}}, "sso")
	if err == nil || !strings.Contains(err.Error(), "aws sso login") {
		t.Errorf("failure hint: err = %v", err)
	}
}

func TestGitHub(t *testing.T) {
	expiry := "2026-11-20 00:00:00 UTC"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Header.Get("Authorization") {
		case "Bearer github_pat_good":
			w.Header().Set("github-authentication-token-expiration", expiry)
			_, _ = w.Write([]byte(`{"login":"octo"}`))
		case "Bearer github_pat_noexp":
			_, _ = w.Write([]byte(`{"login":"octo"}`))
		default:
			w.WriteHeader(http.StatusUnauthorized)
		}
	}))
	defer srv.Close()
	old := GitHubAPI
	GitHubAPI = srv.URL
	defer func() { GitHubAPI = old }()

	ctx := context.Background()
	for token, want := range map[string]string{
		"ghp_classic":       "classic personal access tokens are refused",
		"gho_oauth":         "not a fine-grained",
		"github_pat_noexp":  "no expiry",
		"github_pat_revoke": "GitHub rejected",
		"":                  "empty token",
	} {
		if _, err := CheckGitHubToken(ctx, srv.Client(), token); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("%q: err = %v, want %q", token, err, want)
		}
	}

	tok, err := CheckGitHubToken(ctx, srv.Client(), " github_pat_good\n")
	if err != nil {
		t.Fatal(err)
	}
	if tok.Login != "octo" || tok.Expires != expiry {
		t.Errorf("token = %+v", tok)
	}

	r := &fakeRunner{}
	if err := InstallGitHub(ctx, &Container{ID: "abc", User: "vscode", r: r}, tok); err != nil {
		t.Fatal(err)
	}
	r.assertNoSecretInArgs(t, "github_pat_good")
	if r.calls[1].stdin != "github_pat_good\n" {
		t.Errorf("login stdin = %q", r.calls[1].stdin)
	}

	r = &fakeRunner{fail: map[string]bool{"docker exec -u vscode abc gh --version": true}}
	if err := InstallGitHub(ctx, &Container{ID: "abc", User: "vscode", r: r}, tok); err == nil || !strings.Contains(err.Error(), "gh is not installed") {
		t.Errorf("missing gh: err = %v", err)
	}
}
