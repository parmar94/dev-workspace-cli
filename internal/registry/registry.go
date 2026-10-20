// Package registry is the catalog of languages, agents and tools dws can
// install, plus the templates that bundle them.
package registry

import (
	"errors"
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"

	"go.yaml.in/yaml/v3"

	"github.com/parmar94/dev-workspace-cli/internal/assets"
)

// Kind is the catalog section an entry belongs to.
type Kind string

const (
	KindLanguage Kind = "language"
	KindAgent    Kind = "agent"
	KindTool     Kind = "tool"
)

// Kinds lists every kind in display order.
var Kinds = []Kind{KindLanguage, KindAgent, KindTool}

// Item describes how to install one thing. Exactly one of Feature or Script
// must be set.
type Item struct {
	Description string            `yaml:"description,omitempty"`
	Feature     string            `yaml:"feature,omitempty"`
	Options     map[string]any    `yaml:"options,omitempty"`
	Script      string            `yaml:"script,omitempty"`
	Requires    []string          `yaml:"requires,omitempty"`
	Mounts      []string          `yaml:"mounts,omitempty"`
	Env         map[string]string `yaml:"env,omitempty"`
	Guard       Guard             `yaml:"guard,omitempty"`
}

// Guard lists credential paths and commands that AI agents should be denied
// when this entry is installed. It is a guardrail, not a security boundary.
type Guard struct {
	// Paths are globs agents may not read, e.g. "~/.aws/**".
	Paths []string `yaml:"paths,omitempty"`
	// Commands are command prefixes agents may not run, e.g. "gh auth token".
	Commands []string `yaml:"commands,omitempty"`
	// Env are environment variable names or globs hidden from agent shells.
	Env []string `yaml:"env,omitempty"`
}

// IsZero reports whether g has no rules. yaml uses it for omitempty.
func (g Guard) IsZero() bool {
	return len(g.Paths) == 0 && len(g.Commands) == 0 && len(g.Env) == 0
}

// BaseGuard applies whenever agent guard rules are generated.
var BaseGuard = Guard{
	Paths:    []string{"~/.ssh/**", "**/.env", "**/.env.*"},
	Commands: []string{"printenv", "env"},
}

// Merge returns the union of g and others, keeping first-seen order.
func (g Guard) Merge(others ...Guard) Guard {
	out := Guard{}
	for _, x := range append([]Guard{g}, others...) {
		out.Paths = appendNew(out.Paths, x.Paths...)
		out.Commands = appendNew(out.Commands, x.Commands...)
		out.Env = appendNew(out.Env, x.Env...)
	}
	return out
}

func appendNew(s []string, vs ...string) []string {
	for _, v := range vs {
		if !slices.Contains(s, v) {
			s = append(s, v)
		}
	}
	return s
}

// Entry is an Item together with its name and kind.
type Entry struct {
	Name string
	Kind Kind
	Item
}

// Registry is the full catalog. Names are unique across all sections.
type Registry struct {
	Languages map[string]Item `yaml:"languages,omitempty"`
	Agents    map[string]Item `yaml:"agents,omitempty"`
	Tools     map[string]Item `yaml:"tools,omitempty"`
}

// Builtin returns the catalog embedded in the binary.
func Builtin() (*Registry, error) {
	var r Registry
	if err := yaml.Unmarshal(assets.Registry, &r); err != nil {
		return nil, fmt.Errorf("parse built-in registry: %w", err)
	}
	return &r, nil
}

// Section returns the map for kind, creating it if needed.
func (r *Registry) Section(kind Kind) map[string]Item {
	var m *map[string]Item
	switch kind {
	case KindLanguage:
		m = &r.Languages
	case KindAgent:
		m = &r.Agents
	case KindTool:
		m = &r.Tools
	default:
		panic("registry: unknown kind " + string(kind))
	}
	if *m == nil {
		*m = map[string]Item{}
	}
	return *m
}

// Merge overlays o onto r. Entries in o replace same-named entries in r,
// including ones in a different section.
func (r *Registry) Merge(o *Registry) {
	if o == nil {
		return
	}
	for _, k := range Kinds {
		for name, item := range o.Section(k) {
			r.Remove(name)
			r.Section(k)[name] = item
		}
	}
}

// Remove deletes the named entries from every section.
func (r *Registry) Remove(names ...string) {
	for _, k := range Kinds {
		for _, n := range names {
			delete(r.Section(k), n)
		}
	}
}

// Lookup finds an entry by name in any section.
func (r *Registry) Lookup(name string) (Entry, bool) {
	for _, k := range Kinds {
		if item, ok := r.Section(k)[name]; ok {
			return Entry{Name: name, Kind: k, Item: item}, true
		}
	}
	return Entry{}, false
}

// Names returns the sorted entry names for kind.
func (r *Registry) Names(kind Kind) []string {
	return slices.Sorted(maps.Keys(r.Section(kind)))
}

// Names become script file names, so keep them simple.
var validName = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]*$`)

// Validate checks that names are well-formed and unique, each item has exactly one install
// method, and every requirement exists.
func (r *Registry) Validate() error {
	var errs []error
	seen := map[string]Kind{}
	for _, k := range Kinds {
		for _, name := range r.Names(k) {
			item := r.Section(k)[name]
			if prev, dup := seen[name]; dup {
				errs = append(errs, fmt.Errorf("%q is defined as both %s and %s", name, prev, k))
			}
			seen[name] = k
			if !validName.MatchString(name) {
				errs = append(errs, fmt.Errorf("%s %q: names may only use a-z, 0-9, '.', '_' and '-'", k, name))
			}
			if (item.Feature == "") == (item.Script == "") {
				errs = append(errs, fmt.Errorf("%s %q: set exactly one of feature or script", k, name))
			}
		}
	}
	for _, k := range Kinds {
		for _, name := range r.Names(k) {
			for _, req := range r.Section(k)[name].Requires {
				if _, ok := seen[req]; !ok {
					errs = append(errs, fmt.Errorf("%s %q requires unknown entry %q", k, name, req))
				}
			}
		}
	}
	return errors.Join(errs...)
}

// Resolve expands names with their requirements and returns entries with
// dependencies ordered before their dependents. added lists the entries that
// were pulled in only as requirements.
func (r *Registry) Resolve(names []string) (entries []Entry, added []string, err error) {
	requested := map[string]bool{}
	for _, n := range names {
		requested[n] = true
	}

	const (
		visiting = 1
		done     = 2
	)
	state := map[string]int{}

	var visit func(name string, path []string) error
	visit = func(name string, path []string) error {
		switch state[name] {
		case done:
			return nil
		case visiting:
			return fmt.Errorf("dependency cycle: %s", strings.Join(append(path, name), " -> "))
		}
		e, ok := r.Lookup(name)
		if !ok {
			if len(path) > 0 {
				return fmt.Errorf("%q (required by %q) is not in the registry", name, path[len(path)-1])
			}
			return fmt.Errorf("%q is not in the registry", name)
		}
		state[name] = visiting
		for _, req := range e.Requires {
			if err := visit(req, append(path, name)); err != nil {
				return err
			}
		}
		state[name] = done
		entries = append(entries, e)
		if !requested[name] {
			added = append(added, name)
		}
		return nil
	}

	for _, n := range names {
		if err := visit(n, nil); err != nil {
			return nil, nil, err
		}
	}
	return entries, added, nil
}
