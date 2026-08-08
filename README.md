# diskc

`diskc` is a fast, read-only Linux CLI for answering: what is filling this disk, how fast is it growing, and which process is writing it?

It combines the incident workflow behind `df`, `df -i`, `du`, `find`, `lsof`, `/proc`, and file-growth sampling into one command.

## Features

- Reports byte pressure and inode pressure.
- Ranks large files and directories without crossing filesystem boundaries.
- Measures growth rate over a sampling window.
- Attributes open files to processes through `/proc/<pid>/fd`.
- Finds deleted files still held open by processes.
- Classifies logs, temporary/cache files, archives, and core dumps.
- Works without sudo or root; inaccessible data is skipped with warnings.
- Supports human-readable and JSON output.

## Install

Requires Go 1.22 or newer.

```bash
go run ./cmd/diskc
go build -o diskc ./cmd/diskc
./diskc --help
```

The binary does not need installation or root privileges:

```bash
./diskc /var --sample 5s --deleted
```

For process attribution, run as the same user that owns the processes when possible. Linux `/proc` permissions may hide other users' processes from an unprivileged account. `diskc` never elevates privileges.

## Homebrew

The module is organized for a Homebrew formula:

```ruby
class Diskc < Formula
  desc "Find what is filling a Linux filesystem"
  homepage "https://github.com/diskc/diskc"
  url "https://github.com/diskc/diskc/archive/refs/tags/v0.1.0.tar.gz"
  sha256 "REPLACE_WITH_RELEASE_SHA256"

  def install
    system "go", "build", *std_go_args(output: bin/"diskc"), "./cmd/diskc"
  end

  test do
    assert_match "Find what is filling", shell_output("#{bin}/diskc --help")
  end
end
```

After publishing a tap:

```bash
brew tap your-org/diskc
brew install diskc
```

## Usage

### Quick incident view

```text
$ diskc
Disk pressure: WARNING  92.4% used (38.0 GB free)
Inodes: 41.0% used  |  /

Largest files
   74.8 GB  /var/log/payment/app.log  (log)
   42.1 GB  /tmp/upload-883120.bin   (temporary/cache)
   28.4 GB  /home/service/core.29311 (core dump)

Largest directories
  128.4 GB  /var/log
   67.2 GB  /tmp
   41.5 GB  /home
```

The default path is `/`. Focus a scan with `diskc /var` or `diskc /data --top 50 --depth 6`.

### Find growing files and writers

`--sample` accepts Go duration syntax. It scans the largest files, waits, and measures them again:

```text
$ diskc /var --top 10 --sample 5s
Largest files
   74.8 GB  /var/log/payment/app.log  (log)  ↑ 18.2 MB/s  [1 writer(s)]
   31.6 GB  /var/log/nginx/access.log  (log)  ↑ 4.1 MB/s  [1 writer(s)]
```

At `18.2 MB/s`, the payment log grows about `1.1 GB/minute` or `65.5 GB/hour`.

### Find deleted-but-open files

This catches cases where `df` says the disk is full but `du` cannot find the space:

```text
$ diskc --deleted
Deleted but still open
   38.2 GB  /var/log/payment/app.log  ← payment-api (PID 21819)
```

The directory entry is gone, but the process must close its file descriptor before space is reclaimed.

### Export JSON

```bash
diskc /data --top 50 --sample 3s --deleted --json > disk-report.json
```

Example shape:

```json
{
  "filesystem": {
    "path": "/data",
    "total_bytes": 483183820800,
    "used_bytes": 446676598784,
    "free_bytes": 36507222016,
    "used_percent": 92.4,
    "inode_percent": 41
  },
  "files": [
    {
      "path": "/data/events.log",
      "size_bytes": 13314398617,
      "kind": "log",
      "growth_bytes_per_second": 4299161,
      "writers": [{"pid": 21819, "process": "payment-api", "exe": "/opt/payment/bin/payment-api", "uid": "1001"}]
    }
  ]
}
```

## Flags

```text
--top N             number of large files (default: 20)
--depth N           maximum directory depth (default: 4)
--sample DURATION   measure growth, for example 5s or 1m
--deleted           include deleted files held open by processes
--json              emit machine-readable JSON
```

## Development

```bash
gofmt -w cmd internal
go test ./...
go build ./cmd/diskc
GOOS=linux GOARCH=amd64 go build -o diskc-linux ./cmd/diskc
```

The project uses only the Go standard library. Linux builds use filesystem statistics and `/proc` metadata.
