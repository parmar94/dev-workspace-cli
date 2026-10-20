package config

import (
	"path/filepath"
	"slices"
	"testing"

	"github.com/parmar94/dev-workspace-cli/internal/registry"
)

func TestPath(t *testing.T) {
	t.Setenv("DWS_CONFIG", "")
	t.Setenv("XDG_CONFIG_HOME", "/xdg")
	if p, _ := Path(); p != "/xdg/dws/config.yaml" {
		t.Errorf("XDG path = %s", p)
	}
	t.Setenv("DWS_CONFIG", "/custom.yaml")
	if p, _ := Path(); p != "/custom.yaml" {
		t.Errorf("override path = %s", p)
	}
}

func TestLoadMissingIsEmpty(t *testing.T) {
	c, err := Load(filepath.Join(t.TempDir(), "nope.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if c.BaseImage() != DefaultBaseImage || c.RemoteUser() != DefaultRemoteUser {
		t.Errorf("fallbacks not applied: %s %s", c.BaseImage(), c.RemoteUser())
	}
}

func TestStarterRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dws", "config.yaml")
	if err := WriteStarter(path, false); err != nil {
		t.Fatal(err)
	}
	if err := WriteStarter(path, false); err == nil {
		t.Error("second write without force should fail")
	}
	c, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(c.Defaults.Agents, "claude") || !c.Defaults.SSHAgent {
		t.Errorf("unexpected defaults: %+v", c.Defaults)
	}
	if _, err := c.EffectiveRegistry(); err != nil {
		t.Errorf("starter config registry invalid: %v", err)
	}
}

func TestEffectiveRegistryAndTemplates(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	c := &Config{
		Registry: registry.Registry{Tools: map[string]registry.Item{
			"httpie": {Script: "apt httpie"},
		}},
		Removed:   []string{"fzf"},
		Templates: map[string]registry.Template{"mine": {Languages: []string{"go"}}},
	}
	if err := c.Save(path); err != nil {
		t.Fatal(err)
	}
	c, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	r, err := c.EffectiveRegistry()
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := r.Lookup("httpie"); !ok {
		t.Error("custom tool missing")
	}
	if _, ok := r.Lookup("fzf"); ok {
		t.Error("removed tool still present")
	}
	if _, ok := r.Lookup("go"); !ok {
		t.Error("built-in language missing")
	}

	tm, err := c.EffectiveTemplates()
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := tm["mine"]; !ok {
		t.Error("user template missing")
	}
	if _, ok := tm["go"]; !ok {
		t.Error("built-in template missing")
	}
}
