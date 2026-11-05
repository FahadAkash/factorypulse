---
name: go-testing-loadtest
description: Go testing strategy - unit, integration with testcontainers, duplicate/out-of-order/replay tests, race detector on Windows, load generator and honest benchmark reporting.
---
# Testing & load

## Must-have tests (ingestion)
1. **Duplicate**: same event delivered N times => counts equal to 1 delivery.
2. **Out-of-order**: shuffled `seq` within window => final state/aggregates identical to ordered run.
3. **Double STOP**: exactly 1 active incident.
4. **Replay reconciliation**: truncate aggregates, replay raw `events` => aggregates equal live values (diff must be 0).
5. **DB outage**: stop Postgres container mid-stream => no loss after recovery (assert counts).
6. **Contract**: golden JSON fixtures from `contracts/` parse in Go (and same files used by Unity tests).

## Tooling on Windows
- Integration: `testcontainers-go` (needs Docker Desktop running) or compose up + `-tags=integration`.
- `go test -race` requires CGO + a C compiler (MinGW-w64/TDM-GCC). If absent, run race tests inside Docker/WSL - don't skip silently, say so.
- Table-driven, `t.Parallel()`, no `time.Sleep` (use channels/fake clock). Narrow runs: `go test ./internal/ingest -run TestDuplicate`.

## Load test (`cmd/loadgen`)
Publishes synthetic events at configurable rate (e.g. 100 -> 1k -> 5k/s) with controllable duplicate % and disorder %. Record: throughput, P95/P99 `ingest_latency_seconds`, loss (published vs persisted), duplicate-processing rate, CPU/RAM.

## Benchmark report rule
Fill `docs/benchmark-report.md` ONLY with numbers from real runs, including machine spec + test conditions. If not measured, write `[not measured]`. Never invent or round up numbers.
