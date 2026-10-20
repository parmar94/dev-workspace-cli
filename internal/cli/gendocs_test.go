package cli

import (
	"os"
	"path/filepath"
	"testing"
)

// TestReferenceUpToDate fails when docs/reference no longer matches the
// command tree. Fix with `make docs`.
func TestReferenceUpToDate(t *testing.T) {
	want := filepath.Join("..", "..", "docs", "reference")
	got := t.TempDir()
	if err := genDocs(got); err != nil {
		t.Fatal(err)
	}
	gotFiles, _ := filepath.Glob(filepath.Join(got, "*.md"))
	wantFiles, _ := filepath.Glob(filepath.Join(want, "*.md"))
	if len(gotFiles) != len(wantFiles) {
		t.Fatalf("docs/reference has %d files, generator produces %d; run `make docs`", len(wantFiles), len(gotFiles))
	}
	for _, f := range gotFiles {
		g, _ := os.ReadFile(f)
		w, err := os.ReadFile(filepath.Join(want, filepath.Base(f)))
		if err != nil || string(g) != string(w) {
			t.Errorf("docs/reference/%s is stale; run `make docs`", filepath.Base(f))
		}
	}
}
