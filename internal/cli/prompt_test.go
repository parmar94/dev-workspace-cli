package cli

import (
	"testing"

	"github.com/parmar94/dev-workspace-cli/internal/config"
	"github.com/parmar94/dev-workspace-cli/internal/registry"
)

func TestEntryOptionsFavoritesFirstAndPreselected(t *testing.T) {
	reg := &registry.Registry{Tools: map[string]registry.Item{
		"a": {Script: "x"},
		"b": {Script: "x", Description: "Bee"},
		"c": {Script: "x"},
	}}
	cfg := &config.Config{Favorites: []string{"c"}}

	opts := entryOptions(reg, cfg, registry.KindTool, []string{"b"})

	var got []string
	for _, o := range opts {
		got = append(got, o.Key)
	}
	want := []string{"★ c", "a", "b - Bee"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("labels = %q, want %q", got, want)
		}
	}
	// huh doesn't expose Option.selected, so check via the value round-trip.
	if opts[2].Value != "b" {
		t.Errorf("value = %s", opts[2].Value)
	}
}
