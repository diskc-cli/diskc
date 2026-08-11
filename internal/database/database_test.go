package database

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParsePostgreSQLConfig(t *testing.T) {
	root := t.TempDir()
	config := filepath.Join(root, "postgresql.conf")
	content := "data_directory = '/srv/postgres/data'\nport = 5432\narchive_mode = on\nshared_buffers = 1GB\n"
	if err := os.WriteFile(config, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	settings, paths := parseConfig(config, "postgresql")
	if settings["port"] != "5432" || settings["archive_mode"] != "on" || settings["shared_buffers"] != "1GB" {
		t.Fatalf("unexpected settings: %#v", settings)
	}
	if len(paths.data) != 1 || paths.data[0] != "/srv/postgres/data" {
		t.Fatalf("unexpected data paths: %#v", paths.data)
	}
}

func TestParseRedisRelativePaths(t *testing.T) {
	root := t.TempDir()
	config := filepath.Join(root, "redis.conf")
	if err := os.WriteFile(config, []byte("dir ./data\nappendonly yes\nappendfilename appendonly.aof\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	settings, paths := parseConfig(config, "redis")
	if settings["appendonly"] != "yes" || settings["appendfilename"] != "appendonly.aof" {
		t.Fatalf("unexpected settings: %#v", settings)
	}
	if len(paths.data) != 1 || paths.data[0] != filepath.Join(root, "./data") {
		t.Fatalf("unexpected data paths: %#v", paths.data)
	}
}
