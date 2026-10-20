// Package config loads and saves the dws global config.
package config

import (
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"

	"go.yaml.in/yaml/v3"

	"github.com/parmar94/dev-workspace-cli/internal/assets"
	"github.com/parmar94/dev-workspace-cli/internal/registry"
)

// Built-in fallbacks used when the config leaves a value unset.
const (
	DefaultBaseImage  = "mcr.microsoft.com/devcontainers/base:ubuntu"
	DefaultRemoteUser = "vscode"
)

// Defaults are the starting selections for `dws init`.
type Defaults struct {
	Languages  []string `yaml:"languages,omitempty"`
	Agents     []string `yaml:"agents,omitempty"`
	Tools      []string `yaml:"tools,omitempty"`
	SSHAgent   bool     `yaml:"ssh_agent,omitempty"`
	BaseImage  string   `yaml:"base_image,omitempty"`
	RemoteUser string   `yaml:"remote_user,omitempty"`
}

// Config is the global config file.
type Config struct {
	Defaults  Defaults                     `yaml:"defaults,omitempty"`
	Favorites []string                     `yaml:"favorites,omitempty"`
	Registry  registry.Registry            `yaml:"registry,omitempty"`
	Removed   []string                     `yaml:"removed,omitempty"`
	Templates map[string]registry.Template `yaml:"templates,omitempty"`
}

// Path returns the config file location: $DWS_CONFIG if set, else
// $XDG_CONFIG_HOME/dws/config.yaml, else ~/.config/dws/config.yaml.
func Path() (string, error) {
	if p := os.Getenv("DWS_CONFIG"); p != "" {
		return p, nil
	}
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, "dws", "config.yaml"), nil
}

// Load reads the config at path. A missing file yields an empty config.
func Load(path string) (*Config, error) {
	var c Config
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return &c, nil
	}
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return &c, nil
}

// Save writes c to path, creating parent directories. Comments in an
// existing file are not preserved.
func (c *Config) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return writeFile(path, data)
}

// WriteStarter writes the commented starter config to path. It fails if the
// file exists unless force is set.
func WriteStarter(path string, force bool) error {
	if !force {
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("%s already exists (use --force to overwrite)", path)
		}
	}
	return writeFile(path, assets.DefaultConfig)
}

func writeFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// EffectiveRegistry returns the built-in catalog with the user's entries
// merged in and removed entries dropped.
func (c *Config) EffectiveRegistry() (*registry.Registry, error) {
	r, err := registry.Builtin()
	if err != nil {
		return nil, err
	}
	r.Merge(&c.Registry)
	r.Remove(c.Removed...)
	if err := r.Validate(); err != nil {
		return nil, fmt.Errorf("invalid registry: %w", err)
	}
	return r, nil
}

// EffectiveTemplates returns built-in templates overlaid with user ones.
func (c *Config) EffectiveTemplates() (map[string]registry.Template, error) {
	t, err := registry.BuiltinTemplates()
	if err != nil {
		return nil, err
	}
	maps.Copy(t, c.Templates)
	return t, nil
}

// BaseImage returns the configured base image or the built-in default.
func (c *Config) BaseImage() string {
	if c.Defaults.BaseImage != "" {
		return c.Defaults.BaseImage
	}
	return DefaultBaseImage
}

// RemoteUser returns the configured remote user or the built-in default.
func (c *Config) RemoteUser() string {
	if c.Defaults.RemoteUser != "" {
		return c.Defaults.RemoteUser
	}
	return DefaultRemoteUser
}

// IsFavorite reports whether name is in the favorites list.
func (c *Config) IsFavorite(name string) bool {
	return slices.Contains(c.Favorites, name)
}
