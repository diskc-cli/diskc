package report

import (
	"bytes"
	"strings"
	"testing"

	"github.com/diskc/diskc/internal/disk"
	"github.com/diskc/diskc/internal/health"
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

func TestPrintIncludesGrowthAndDiagnostics(t *testing.T) {
	var output bytes.Buffer
	data := Data{
		Filesystem: disk.Filesystem{Path: "/", FreeBytes: 1024},
		Files:      []disk.File{{Path: "/large.log", Size: 2048, Kind: "log"}},
		Growing:    []disk.File{{Path: "/small-fast.log", Size: 10, Kind: "log", Growth: 512}},
		Findings:   []health.Finding{{Severity: "warning", Kind: "device-io", Message: "sda: 99% busy"}},
	}
	if err := Print(&output, data, false); err != nil {
		t.Fatal(err)
	}
	text := output.String()
	for _, expected := range []string{"Fastest growth", "/small-fast.log", "System diagnostics", "sda: 99% busy"} {
		if !strings.Contains(text, expected) {
			t.Fatalf("expected %q in output: %s", expected, text)
		}
	}
}
