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
		{Path: ".devcontainer/agent-policy/p.json", Data: []byte("{}\n")},
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

	// Force replaces the generated dirs, so stale scripts and policies disappear.
	if err := Write(root, files[:1], true); err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{".devcontainer/scripts/a.sh", ".devcontainer/agent-policy/p.json"} {
		if _, err := os.Stat(filepath.Join(root, p)); !os.IsNotExist(err) {
			t.Errorf("stale %s still present: %v", p, err)
		}
	}
}
