package devcontainer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWrite(t *testing.T) {
	root := t.TempDir()
	files := []File{
		{Path: ".devcontainer/devcontainer.json", Data: []byte("{}\n")},
		{Path: ".devcontainer/scripts/a.sh", Data: []byte("echo a\n"), Exec: true},
	}
	if err := Write(root, files, false); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(filepath.Join(root, ".devcontainer/scripts/a.sh"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o755 {
		t.Errorf("script mode = %v", info.Mode().Perm())
	}

	if err := Write(root, files[:1], false); err == nil {
		t.Fatal("second write without force should fail")
	}

	// Force replaces the scripts dir, so stale scripts disappear.
	if err := Write(root, files[:1], true); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, ".devcontainer/scripts/a.sh")); !os.IsNotExist(err) {
		t.Errorf("stale script still present: %v", err)
	}
}
