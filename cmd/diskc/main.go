package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/diskc/diskc/internal/disk"
	"github.com/diskc/diskc/internal/mount"
	"github.com/diskc/diskc/internal/proc"
	"github.com/diskc/diskc/internal/report"
)

func main() {
	var (
		top     int
		depth   int
		sample  = 3 * time.Second
		deleted bool
		jsonOut bool
		all     bool
	)
	flag.IntVar(&top, "top", 20, "number of large files to report")
	flag.IntVar(&depth, "depth", 4, "maximum directory depth to scan")
	flag.DurationVar(&sample, "sample", sample, "measure growth over this duration; use 0 to disable, for example 5s")
	flag.BoolVar(&deleted, "deleted", false, "include deleted files still held open by processes")
	flag.BoolVar(&jsonOut, "json", false, "emit machine-readable JSON")
	flag.BoolVar(&all, "all", false, "inspect all mounted physical filesystems")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: diskc [path] [flags]\n\nFind what is filling a Linux filesystem.\n\nFlags:\n")
		flag.PrintDefaults()
	}
	args := reorderArgs(os.Args[1:])
	if err := flag.CommandLine.Parse(args); err != nil {
		os.Exit(2)
	}

	if top < 1 || depth < 0 || sample < 0 {
		fmt.Fprintln(os.Stderr, "diskc: top must be positive; depth and sample cannot be negative")
		os.Exit(2)
	}
	roots := make([]string, 0, flag.NArg()+1)
	for index := 0; index < flag.NArg(); index++ {
		roots = append(roots, flag.Arg(index))
	}
	if all {
		mounts, err := mount.All()
		if err != nil {
			fatal(err)
		}
		roots = append(roots, mounts...)
	}
	if len(roots) == 0 {
		roots = append(roots, "/")
	}
	outputs := make([]report.Data, 0, len(roots))
	for _, root := range unique(roots) {
		filesystem, err := disk.FilesystemUsage(root)
		if err != nil {
			fmt.Fprintf(os.Stderr, "diskc: skipping %s: %v\n", root, err)
			continue
		}
		files, directories, warnings, err := disk.Scan(root, filesystem.Device, depth, top)
		if err != nil {
			fmt.Fprintf(os.Stderr, "diskc: skipping %s: %v\n", root, err)
			continue
		}
		if sample > 0 {
			if err := disk.MeasureGrowth(files, sample); err != nil {
				warnings = append(warnings, err.Error())
			}
		}
		disk.UpdateTrend(files, &filesystem)
		proc.AttributeWriters(files)
		var openDeleted []proc.DeletedFile
		if deleted {
			openDeleted = proc.DeletedFiles()
		}
		outputs = append(outputs, report.Data{Filesystem: filesystem, Files: files, Directories: directories, Deleted: openDeleted, Warnings: warnings})
	}
	if len(outputs) == 0 {
		fatal(fmt.Errorf("no filesystems could be inspected"))
	}
	if err := report.PrintMany(os.Stdout, outputs, jsonOut); err != nil {
		fatal(err)
	}
}

func reorderArgs(args []string) []string {
	pathCount := 0
	for pathCount < len(args) && !strings.HasPrefix(args[pathCount], "-") {
		pathCount++
	}
	if pathCount == 0 {
		return args
	}
	return append(append([]string{}, args[pathCount:]...), args[:pathCount]...)
}

func unique(paths []string) []string {
	seen := make(map[string]bool, len(paths))
	result := make([]string, 0, len(paths))
	for _, path := range paths {
		if !seen[path] {
			seen[path] = true
			result = append(result, path)
		}
	}
	return result
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "diskc: %v\n", err)
	os.Exit(2)
}
