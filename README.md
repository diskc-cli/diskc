# diskc

Find what is filling your Linux disk, how fast it is growing, and which process is writing it.

`diskc` is a read-only, SRE-focused CLI that combines the usual `df`, `df -i`, `du`, `find`, `lsof`, `ps`, and `/proc` investigation into one workflow.

## Quick Start

Build and run from source:

```bash
go build -o diskc ./cmd/diskc
./diskc
```

The default command scans `/`, samples growth for three seconds, and reports filesystem pressure, largest files/directories, growth rate, estimated time-to-full, and writable processes. Add `--deleted` to include deleted-open file detection.

`diskc` is read-only and does not require `sudo` or root privileges. Linux `/proc` permissions may limit process details for other users.

## Example Output

```text
$ ./diskc
Disk pressure: WARNING  92.4% used (38.0 GB free)
Inodes: 41.0% used  |  /
Trend: +2.4 GB/hour
Estimated filesystem full: ~35 minutes

Largest files
   74.8 GB  /var/log/payment/app.log  (log)  ↑ 18.2 MB/s  [1 writer(s)]
   42.1 GB  /tmp/upload-883120.bin   (temporary/cache)  ↑ 11.8 MB/s

Active writers
/var/log/payment/app.log
└─ PID 21819  payment-api
   ├─ exe      /opt/payment/bin/payment-api
   └─ service  payment.service

Potential issues
/var/log/payment/app.log
├─ rapid growth: 65.5 GB/hour
└─ filesystem full in ~35 minutes
```

## Common Commands

Inspect a directory or filesystem:

```bash
./diskc /var
./diskc /var/log --top 50 --depth 6
```

Measure growth over a custom window:

```bash
./diskc /var --sample 5s
```

Inspect multiple filesystems explicitly:

```bash
./diskc / /data /backup --sample 5s
```

Discover physical filesystems automatically:

```bash
./diskc --all --sample 5s
```

Virtual filesystems such as `/proc`, `/sys`, `tmpfs`, and cgroups are excluded. Separate filesystem results are never mixed.

Watch continuously until `Ctrl-C`:

```bash
./diskc -watch
./diskc --watch --interval 5s /data
```

The default refresh interval is three seconds between refresh starts. If a scan takes longer, the next refresh starts immediately.

Find deleted-but-open files:

```bash
./diskc --deleted
```

Example:

```text
Deleted but still open
   38.2 GB  /var/log/payment/app.log  ← payment-api (PID 21819)
```

## JSON Output

```bash
./diskc /data --top 50 --sample 3s --deleted --json > disk-report.json
```

For multiple filesystems, JSON output is an array:

```json
[
  {
    "filesystem": {
      "path": "/data",
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
]
```

With `--watch --json`, each refresh is emitted as a separate JSON document without terminal-control sequences.

## Installation

Requires Go 1.22 or newer:

```bash
go run ./cmd/diskc
go build -o diskc ./cmd/diskc
```

The module is organized for Homebrew. After publishing a tap:

```bash
brew tap your-org/diskc
brew install diskc
```

The formula builds `./cmd/diskc` and installs the resulting `diskc` binary.

## Flags

```text
--top N              number of large files (default: 20)
--depth N            maximum directory depth (default: 4)
--sample DURATION    measure growth (default: 3s; use 0 to disable)
--deleted            include deleted files held open by processes
--json               emit machine-readable JSON
--all                inspect all mounted physical filesystems
--watch              refresh continuously until interrupted
--interval DURATION  watch refresh interval (default: 3s)
```

## Disclaimer

Parts of this project, including source code, documentation, examples, and configuration, may be generated or assisted by artificial intelligence. The project is provided on an “AS IS” and “AS AVAILABLE” basis, without warranties of any kind.

Use `diskc` at your own risk. The authors, maintainers, contributors, and distributors are not liable for any damage, data loss, corruption, downtime, outage, security incident, or other loss arising from the use of all or any part of this project. Review, test, and validate the code and its output before using it in production.

## Contributing

The project uses only the Go standard library. Linux builds use filesystem statistics and `/proc` metadata.

Run formatting, tests, static analysis, and builds with:

```bash
gofmt -w cmd internal
go test ./...
go vet ./...
go build ./cmd/diskc
GOOS=linux GOARCH=amd64 go build -o diskc-linux ./cmd/diskc
```

Pull requests and issues are welcome at [github.com/diskc-cli/diskc](https://github.com/diskc-cli/diskc).
