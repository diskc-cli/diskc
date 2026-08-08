package report

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/diskc/diskc/internal/disk"
	"github.com/diskc/diskc/internal/proc"
)

type Data struct {
	Filesystem  disk.Filesystem    `json:"filesystem"`
	Files       []disk.File        `json:"files"`
	Directories []disk.Directory   `json:"directories"`
	Deleted     []proc.DeletedFile `json:"deleted,omitempty"`
	Warnings    []string           `json:"warnings,omitempty"`
}

func Print(writer io.Writer, data Data, jsonOutput bool) error {
	return PrintMany(writer, []Data{data}, jsonOutput)
}

func PrintMany(writer io.Writer, data []Data, jsonOutput bool) error {
	if jsonOutput {
		encoder := json.NewEncoder(writer)
		encoder.SetIndent("", "  ")
		if len(data) == 1 {
			return encoder.Encode(data[0])
		}
		return encoder.Encode(data)
	}
	for index, item := range data {
		if len(data) > 1 {
			if index > 0 {
				fmt.Fprintln(writer)
			}
			fmt.Fprintf(writer, "===== %s =====\n", item.Filesystem.Path)
		}
		if err := text(writer, item); err != nil {
			return err
		}
	}
	return nil
}

func text(writer io.Writer, data Data) error {
	pressure := "OK"
	if data.Filesystem.UsedPercent >= 95 || data.Filesystem.InodePercent >= 95 {
		pressure = "CRITICAL"
	} else if data.Filesystem.UsedPercent >= 90 || data.Filesystem.InodePercent >= 90 {
		pressure = "WARNING"
	}
	fmt.Fprintf(writer, "Disk pressure: %s  %.1f%% used (%s free)\n", pressure, data.Filesystem.UsedPercent, formatBytes(data.Filesystem.FreeBytes))
	fmt.Fprintf(writer, "Inodes: %.1f%% used  |  %s\n", data.Filesystem.InodePercent, data.Filesystem.Path)
	if data.Filesystem.GrowthBytesPerSecond > 0 {
		fmt.Fprintf(writer, "Trend: +%s/hour\n", formatBytes(uint64(data.Filesystem.GrowthBytesPerSecond*3600)))
		if data.Filesystem.FullInSeconds > 0 {
			fmt.Fprintf(writer, "Estimated filesystem full: ~%s\n", duration(data.Filesystem.FullInSeconds))
		}
	}
	fmt.Fprintln(writer)
	fmt.Fprintln(writer, "Largest files")
	for _, file := range data.Files {
		growth := ""
		if file.Growth != 0 {
			growth = fmt.Sprintf("  ↑ %s/s", formatBytes(uint64(abs(file.Growth))))
		}
		writers := ""
		if len(file.Writers) > 0 {
			writers = fmt.Sprintf("  [%d writer(s)]", len(file.Writers))
		}
		fmt.Fprintf(writer, "%10s  %s  (%s)%s%s\n", formatBytes(uint64(file.Size)), file.Path, file.Kind, growth, writers)
	}
	if len(data.Directories) > 0 {
		fmt.Fprintln(writer, "\nLargest directories")
		for _, directory := range data.Directories {
			fmt.Fprintf(writer, "%10s  %s\n", formatBytes(uint64(directory.Size)), directory.Path)
		}
	}
	if hasWriters(data.Files) {
		fmt.Fprintln(writer, "\nActive writers")
		for _, file := range data.Files {
			for _, processWriter := range file.Writers {
				fmt.Fprintf(writer, "%s\n└─ PID %d  %s\n", file.Path, processWriter.PID, processWriter.Process)
				if processWriter.Exe != "" {
					fmt.Fprintf(writer, "   ├─ exe      %s\n", processWriter.Exe)
				}
				if processWriter.Service != "" {
					fmt.Fprintf(writer, "   └─ service  %s\n", processWriter.Service)
				}
			}
		}
	}
	if hasGrowth(data.Files) {
		fmt.Fprintln(writer, "\nPotential issues")
		for _, file := range data.Files {
			if file.Growth <= 0 {
				continue
			}
			fmt.Fprintf(writer, "%s\n├─ rapid growth: %s/hour\n", file.Path, formatBytes(uint64(file.Growth*3600)))
			if data.Filesystem.FullInSeconds > 0 {
				fmt.Fprintf(writer, "└─ filesystem full in ~%s\n", duration(data.Filesystem.FullInSeconds))
			}
		}
	}
	if len(data.Deleted) > 0 {
		fmt.Fprintln(writer, "\nDeleted but still open")
		for _, file := range data.Deleted {
			fmt.Fprintf(writer, "%10s  %s", formatBytes(uint64(file.Size)), file.Path)
			for _, processWriter := range file.Writers {
				fmt.Fprintf(writer, "  ← %s (PID %d)", processWriter.Process, processWriter.PID)
			}
			fmt.Fprintln(writer)
		}
	}
	if len(data.Warnings) > 0 {
		fmt.Fprintln(writer, "\nWarnings")
		for _, warning := range data.Warnings {
			fmt.Fprintf(writer, "- %s\n", warning)
		}
	}
	return nil
}

func formatBytes(value uint64) string {
	units := []string{"B", "KB", "MB", "GB", "TB", "PB"}
	number := float64(value)
	unit := 0
	for number >= 1024 && unit < len(units)-1 {
		number /= 1024
		unit++
	}
	if unit == 0 {
		return fmt.Sprintf("%d B", value)
	}
	return fmt.Sprintf("%.1f %s", number, units[unit])
}

func abs(value float64) float64 {
	if value < 0 {
		return -value
	}
	return value
}

func hasWriters(files []disk.File) bool {
	for _, file := range files {
		if len(file.Writers) > 0 {
			return true
		}
	}
	return false
}

func hasGrowth(files []disk.File) bool {
	for _, file := range files {
		if file.Growth > 0 {
			return true
		}
	}
	return false
}

func duration(seconds float64) string {
	minutes := int(seconds / 60)
	if minutes < 60 {
		return fmt.Sprintf("%d minutes", minutes)
	}
	return fmt.Sprintf("%d hours %d minutes", minutes/60, minutes%60)
}
