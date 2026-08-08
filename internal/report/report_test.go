package report

import (
	"bytes"
	"strings"
	"testing"

	"github.com/diskc/diskc/internal/disk"
)

func TestPrintJSON(t *testing.T) {
	var output bytes.Buffer
	data := Data{Filesystem: disk.Filesystem{Path: "/", FreeBytes: 1024}, Files: []disk.File{{Path: "/tmp/app.log", Size: 2048, Kind: "log"}}}
	if err := Print(&output, data, true); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), `"/tmp/app.log"`) || !strings.Contains(output.String(), `"size_bytes": 2048`) {
		t.Fatalf("unexpected JSON: %s", output.String())
	}
}
