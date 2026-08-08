package disk

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestScanRanksFilesAndClassifiesThem(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "var", "log"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "var", "log", "app.log"), []byte("12345"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "small"), []byte("1"), 0o644); err != nil {
		t.Fatal(err)
	}
	filesystem, err := FilesystemUsage(root)
	if err != nil {
		t.Fatal(err)
	}
	files, directories, warnings, err := Scan(root, filesystem.Device, 4, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(warnings) != 0 || len(files) != 2 || len(directories) == 0 {
		t.Fatalf("unexpected scan result: files=%d directories=%d warnings=%v", len(files), len(directories), warnings)
	}
	if files[0].Kind != "log" || files[0].Path != filepath.Join(root, "var", "log", "app.log") {
		t.Fatalf("unexpected top file: %+v", files[0])
	}
}

func TestMeasureGrowth(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "growing.log")
	if err := os.WriteFile(path, []byte("start"), 0o644); err != nil {
		t.Fatal(err)
	}
	files := []File{{Path: path, Size: 5, Kind: "log"}}
	go func() {
		time.Sleep(20 * time.Millisecond)
		file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0)
		if err == nil {
			_, _ = file.WriteString("growth")
			_ = file.Close()
		}
	}()
	if err := MeasureGrowth(files, 50*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if files[0].Growth <= 0 {
		t.Fatalf("expected positive growth, got %f", files[0].Growth)
	}
}
