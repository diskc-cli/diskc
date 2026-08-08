package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/diskc/diskc/internal/disk"
	"github.com/diskc/diskc/internal/proc"
	"github.com/diskc/diskc/internal/report"
)

func main() {
	var (
		top     int
		depth   int
		sample  time.Duration
		deleted bool
		jsonOut bool
	)
	flag.IntVar(&top, "top", 20, "number of large files to report")
	flag.IntVar(&depth, "depth", 4, "maximum directory depth to scan")
	flag.DurationVar(&sample, "sample", 0, "measure growth over this duration, for example 5s")
	flag.BoolVar(&deleted, "deleted", false, "include deleted files still held open by processes")
	flag.BoolVar(&jsonOut, "json", false, "emit machine-readable JSON")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: diskc [path] [flags]\n\nFind what is filling a Linux filesystem.\n\nFlags:\n")
		flag.PrintDefaults()
	}
	args := os.Args[1:]
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		args = append(args[1:], args[0])
	}
	if err := flag.CommandLine.Parse(args); err != nil {
		os.Exit(2)
	}

	if top < 1 || depth < 0 || sample < 0 {
		fmt.Fprintln(os.Stderr, "diskc: top must be positive; depth and sample cannot be negative")
		os.Exit(2)
	}
	root := "/"
	if flag.NArg() > 0 {
		root = flag.Arg(0)
	}

	filesystem, err := disk.FilesystemUsage(root)
	if err != nil {
		fatal(err)
	}
	files, directories, warnings, err := disk.Scan(root, filesystem.Device, depth, top)
	if err != nil {
		fatal(err)
	}
	if sample > 0 {
		if err := disk.MeasureGrowth(files, sample); err != nil {
			warnings = append(warnings, err.Error())
		}
	}
	proc.AttributeWriters(files)

	var openDeleted []proc.DeletedFile
	if deleted {
		openDeleted = proc.DeletedFiles()
	}
	output := report.Data{Filesystem: filesystem, Files: files, Directories: directories, Deleted: openDeleted, Warnings: warnings}
	if err := report.Print(os.Stdout, output, jsonOut); err != nil {
		fatal(err)
	}
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "diskc: %v\n", err)
	os.Exit(2)
}
