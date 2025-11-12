# Bitcask Inspired KV Store: Append-Only Logs Meet In-Memory Hash Indexes

A lightweight, high-performance key-value store inspired by [Bitcask](https://riak.com/assets/bitcask-intro.pdf).
Built for simplicity, speed, and durability—perfect for learning system design, experimenting with storage engines, or using as a fast KV store in your projects.

### Implementation

![Header](https://github.com/himakhaitan/logkv-store/blob/main/resources/header.png)

Read the full write-up: [Building a Bitcask Inspired KV Store: Append-Only Logs Meet In-Memory Hash Indexes](https://himakhaitan.substack.com/p/building-a-bitcask-inspired-kv-store)

The design prioritizes simplicity and correctness for educational clarity; production deployments would add compaction/merge, checksums, file rotation thresholds, and crash-consistency hardening.

## Run Locally

### Prerequisites

- Go 1.24+
- macOS/Linux recommended (Windows should work with minor path adjustments)

### Clone

```bash
git clone https://github.com/himakhaitan/logkv-store.git
cd logkv-store
```

### Build

Build the server and CLI binaries:

```bash
go build -o bin/logkvd ./cmd/logkvd
go build -o bin/logkv ./cmd/logkv-cli
```

### Run the server

```bash
# Optional: customize bind address (default ":8080")
export LOGKV_ADDR=":8080"

./bin/logkvd
```

By default, data is persisted under the `data/` directory at the project root. 

### Test

To run all unit tests in the project:

```bash
go test ./...
```

To measure how much of your code is covered by tests:

```bash
go test ./... -cover
```

For a detailed function-by-function coverage report:

```bash
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out


#Optional If you want to see which lines are covered (green/red):

go tool cover -html=coverage.out
```

### Interacting

- An HTTP server is started for basic operations.
- A CLI binary (`bin/logkv`) is available for local interactions.

Note: This README intentionally omits specific CLI commands and API route details per project guidelines.

## Configuration

- Address: set `LOGKV_ADDR` (e.g., `:8080`).
- Data directory: defaults to `data/` (see `pkg/config/config.go`).

## Limitations (Current)

- No background compaction/merge to reclaim space.
- No checksums or CRC verification on records.
- Fixed single-process access pattern; no WAL fencing/locks for multi-process writers.
- Entire KeyDir must fit in memory.
- Basic HTTP surface; minimal validation.

## Contributing

Issues and PRs are welcome. Keep changes focused and well-documented. Favor clarity over cleverness.

## Feature Ideas You Can Add Next

- Compaction/merge job: coalesce latest values, drop tombstones, rewrite segments.
- Active segment rotation based on size/time thresholds.
- Checksums (CRC32/XXH3) for each record; verify on read.
- Bloom filter per segment to reduce unnecessary disk reads.
- Snapshot and restore utilities.
- Compression (per record or per segment).
- TTL/expiry and lazy cleanup.
- Metrics and tracing (Prometheus/OpenTelemetry).
- File-level and process-level locking for crash safety and multi-process protection.
- Configurable data directory via env/flags and structured config file.
- Batch operations (multi-set, multi-get) with pipelining.
- Simple authentication/authorization layer for the HTTP surface.
- Graceful schema evolution for record headers.
- Fsync strategies (always, interval, buffered) tuned via config.

## Inspired by Bitcask

![bitcask](https://github.com/himakhaitan/logkv-store/blob/main/resources/bitcask.png)

This project draws inspiration from the Bitcask storage model. See the original paper and community resources for deeper background.

## Explore More

