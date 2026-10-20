package registry

import (
	"fmt"
	"io/fs"
	"path"
	"strings"

	"go.yaml.in/yaml/v3"

	"github.com/parmar94/dev-workspace-cli/internal/assets"
)

// Template is a named preset of languages, agents and tools.
type Template struct {
	Description string   `yaml:"description,omitempty"`
	Languages   []string `yaml:"languages,omitempty"`
	Agents      []string `yaml:"agents,omitempty"`
	Tools       []string `yaml:"tools,omitempty"`
	SSHAgent    *bool    `yaml:"ssh_agent,omitempty"`
	BaseImage   string   `yaml:"base_image,omitempty"`
}

// BuiltinTemplates returns the templates embedded in the binary, keyed by
// file name without extension.
func BuiltinTemplates() (map[string]Template, error) {
	files, err := fs.Glob(assets.Templates, "templates/*.yaml")
	if err != nil {
		return nil, err
	}
	out := make(map[string]Template, len(files))
	for _, f := range files {
		data, err := fs.ReadFile(assets.Templates, f)
		if err != nil {
			return nil, err
		}
		var t Template
		if err := yaml.Unmarshal(data, &t); err != nil {
			return nil, fmt.Errorf("parse template %s: %w", f, err)
		}
		out[strings.TrimSuffix(path.Base(f), ".yaml")] = t
	}
	return out, nil
}
