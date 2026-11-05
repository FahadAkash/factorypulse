---
paths:
  - "backend/**/*.go"
---
# Go rules
- Go 1.22+, `slog` JSON logging, `context.Context` first param, wrap errors `fmt.Errorf("op: %w", err)`, sentinel errors with `errors.Is/As`.
- Packages by domain under `internal/` (ingest, processing, incident, oee, bottleneck, quality, shift, store, api, ws, telemetry). No `utils`/`common` dumping grounds. Interfaces defined by the consumer.
- No globals/`init()` side effects; dependencies injected in `cmd/factorypulse/main.go`.
- Concurrency: bounded channels, fixed worker pools, `errgroup` for lifecycle, every goroutine stops on ctx cancel. No unbounded `go func()` per message. Test with `-race`.
- DB: pgx v5 pool, explicit transactions, parameterized SQL only, migrations via golang-migrate (`backend/migrations`).
- Money/time: `time.Time` UTC; durations as `time.Duration`; metrics never use float equality.
- Tests: table-driven, `t.Parallel()` where safe, fakes over mocks, no sleeps (use channels/clock interface).
- Keep functions <60 lines; exported identifiers documented in one line.
