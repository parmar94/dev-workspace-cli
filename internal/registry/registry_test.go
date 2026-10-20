package registry

import (
	"slices"
	"strings"
	"testing"
)

func names(es []Entry) []string {
	out := make([]string, len(es))
	for i, e := range es {
		out[i] = e.Name
	}
	return out
}

func TestBuiltinIsValid(t *testing.T) {
	r, err := Builtin()
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Validate(); err != nil {
		t.Fatal(err)
	}
	for _, k := range Kinds {
		if len(r.Names(k)) == 0 {
			t.Errorf("no built-in %s entries", k)
		}
	}
}

func TestBuiltinTemplatesReferenceKnownEntries(t *testing.T) {
	r, err := Builtin()
	if err != nil {
		t.Fatal(err)
	}
	tmpls, err := BuiltinTemplates()
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := tmpls["go"]; !ok {
		t.Fatal("missing built-in go template")
	}
	for name, tm := range tmpls {
		all := slices.Concat(tm.Languages, tm.Agents, tm.Tools)
		if _, _, err := r.Resolve(all); err != nil {
			t.Errorf("template %s: %v", name, err)
		}
	}
}

func TestResolveOrdersDependenciesFirst(t *testing.T) {
	r := &Registry{
		Languages: map[string]Item{"node": {Feature: "f/node"}},
		Agents:    map[string]Item{"codex": {Script: "x", Requires: []string{"node"}}},
		Tools:     map[string]Item{"gh": {Feature: "f/gh"}},
	}
	got, added, err := r.Resolve([]string{"gh", "codex"})
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"gh", "node", "codex"}; !slices.Equal(names(got), want) {
		t.Errorf("order = %v, want %v", names(got), want)
	}
	if want := []string{"node"}; !slices.Equal(added, want) {
		t.Errorf("added = %v, want %v", added, want)
	}

	// Explicitly requested dependencies are not reported as added.
	_, added, err = r.Resolve([]string{"node", "codex"})
	if err != nil {
		t.Fatal(err)
	}
	if len(added) != 0 {
		t.Errorf("added = %v, want none", added)
	}
}

func TestResolveErrors(t *testing.T) {
	r := &Registry{Tools: map[string]Item{
		"a": {Script: "x", Requires: []string{"b"}},
		"b": {Script: "x", Requires: []string{"a"}},
		"c": {Script: "x", Requires: []string{"missing"}},
	}}
	tests := map[string]struct {
		in   []string
		want string
	}{
		"unknown": {[]string{"nope"}, `"nope" is not in the registry`},
		"cycle":   {[]string{"a"}, "dependency cycle: a -> b -> a"},
		"missing": {[]string{"c"}, `"missing" (required by "c")`},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			_, _, err := r.Resolve(tc.in)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("err = %v, want containing %q", err, tc.want)
			}
		})
	}
}

func TestMergeReplacesAcrossSections(t *testing.T) {
	r := &Registry{Tools: map[string]Item{"jq": {Script: "apt jq"}}}
	r.Merge(&Registry{
		Languages: map[string]Item{"jq": {Feature: "f/jq"}},
		Tools:     map[string]Item{"rg": {Script: "apt rg"}},
	})
	e, ok := r.Lookup("jq")
	if !ok || e.Kind != KindLanguage || e.Feature != "f/jq" {
		t.Errorf("jq = %+v, %v", e, ok)
	}
	if _, ok := r.Tools["jq"]; ok {
		t.Error("old jq entry still in tools")
	}
	if _, ok := r.Lookup("rg"); !ok {
		t.Error("rg not merged")
	}
}

func TestGuardMerge(t *testing.T) {
	r, err := Builtin()
	if err != nil {
		t.Fatal(err)
	}
	aws, _ := r.Lookup("aws-cli")
	gh, _ := r.Lookup("gh")
	g := BaseGuard.Merge(aws.Guard, gh.Guard, aws.Guard)
	for _, want := range []string{"~/.ssh/**", "~/.aws/**", "~/.config/gh/**"} {
		if !slices.Contains(g.Paths, want) {
			t.Errorf("paths missing %s: %v", want, g.Paths)
		}
	}
	if n := len(g.Commands); n != len(BaseGuard.Commands)+len(aws.Guard.Commands)+len(gh.Guard.Commands) {
		t.Errorf("commands not de-duplicated: %v", g.Commands)
	}
	if (Guard{}).IsZero() != true || g.IsZero() {
		t.Error("IsZero wrong")
	}
}

func TestValidate(t *testing.T) {
	r := &Registry{
		Languages: map[string]Item{"x": {Feature: "f"}},
		Tools: map[string]Item{
			"x":    {Script: "s"},
			"both": {Feature: "f", Script: "s"},
			"none": {},
			"dep":  {Script: "s", Requires: []string{"ghost"}},
			"../x": {Script: "s"},
		},
	}
	err := r.Validate()
	if err == nil {
		t.Fatal("want error")
	}
	for _, want := range []string{`"x" is defined as both`, `"both": set exactly one`, `"none": set exactly one`, `unknown entry "ghost"`, `"../x": names may only`} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error missing %q:\n%v", want, err)
		}
	}
}
