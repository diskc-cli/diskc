package docker

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadConfigDataRoot(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "daemon.json")
	if err := os.WriteFile(path, []byte(`{"data-root":"/srv/docker"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	config, err := readConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if config.DataRoot != "/srv/docker" {
		t.Fatalf("unexpected data root: %q", config.DataRoot)
	}
}

func TestUniquePaths(t *testing.T) {
	paths := uniquePaths([]PathReport{{Path: "/var/lib/docker", Kind: "data"}, {Path: "/var/lib/docker", Kind: "duplicate"}})
	if len(paths) != 1 || paths[0].Kind != "data" {
		t.Fatalf("unexpected paths: %#v", paths)
	}
}
