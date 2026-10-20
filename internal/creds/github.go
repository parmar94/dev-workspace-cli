package creds

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// GitHubAPI is the API base URL; tests point it at a fake server.
var GitHubAPI = "https://api.github.com"

// GitHubToken is a validated fine-grained personal access token.
type GitHubToken struct {
	Value   string
	Login   string
	Expires string
}

// CheckGitHubToken accepts only fine-grained PATs that have an expiry, and
// confirms the token works.
func CheckGitHubToken(ctx context.Context, client *http.Client, token string) (GitHubToken, error) {
	token = strings.TrimSpace(token)
	switch {
	case token == "":
		return GitHubToken{}, errors.New("empty token")
	case strings.HasPrefix(token, "ghp_"):
		return GitHubToken{}, errors.New("classic personal access tokens are refused: they can't be limited to specific repositories. Create a fine-grained token (github_pat_...) with an expiry instead")
	case !strings.HasPrefix(token, "github_pat_"):
		return GitHubToken{}, errors.New("not a fine-grained personal access token (expected github_pat_...)")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, GitHubAPI+"/user", nil)
	if err != nil {
		return GitHubToken{}, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := client.Do(req)
	if err != nil {
		return GitHubToken{}, fmt.Errorf("check token with GitHub: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return GitHubToken{}, fmt.Errorf("GitHub rejected the token: %s", resp.Status)
	}
	exp := resp.Header.Get("github-authentication-token-expiration")
	if exp == "" {
		return GitHubToken{}, errors.New("token has no expiry; create one with an expiration date")
	}
	var user struct {
		Login string `json:"login"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&user)
	return GitHubToken{Value: token, Login: user.Login, Expires: exp}, nil
}

// InstallGitHub logs gh in with the token inside the container and sets gh
// up as git's credential helper.
func InstallGitHub(ctx context.Context, ctr *Container, t GitHubToken) error {
	if _, err := ctr.Exec(ctx, nil, "gh", "--version"); err != nil {
		return fmt.Errorf("gh is not installed in the container (add it with `dws init --tool gh --force` and rebuild): %w", err)
	}
	if _, err := ctr.Exec(ctx, strings.NewReader(t.Value+"\n"), "gh", "auth", "login", "--with-token"); err != nil {
		return err
	}
	_, err := ctr.Exec(ctx, nil, "gh", "auth", "setup-git")
	return err
}
