---
name: go-backend-core
description: Go modular-monolith structure, config, logging, graceful shutdown, bounded worker pools and error handling for FactoryPulse. Use when creating or restructuring any Go code under backend/.
---
# Go backend core

## Layout
```
backend/
  cmd/factorypulse/main.go     # wiring only
  cmd/loadgen/main.go          # synthetic event generator (benchmarks)
  internal/{config,domain,ingest,processing,incident,oee,bottleneck,quality,shift,store/{pg,redis},api,ws,telemetry}
  migrations/
```
`domain` = pure types + rules (no I/O). Other packages depend inward on `domain`. Split into services only with a measured reason.

## Lifecycle (errgroup + signal)
```go
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer stop()
g, ctx := errgroup.WithContext(ctx)
g.Go(func() error { return ingest.Run(ctx) })
g.Go(func() error { return api.Serve(ctx) })
if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) { slog.Error("fatal", "err", err); os.Exit(1) }
```
(Windows: `os.Interrupt` works for Ctrl+C; SIGTERM is accepted but never delivered - Docker stop uses it in-container.)

## Bounded worker pool (no goroutine-per-message)
```go
jobs := make(chan Event, 4096)              // bounded => backpressure
for i := 0; i < workers; i++ {
  g.Go(func() error {
    for { select {
      case <-ctx.Done(): return nil
      case ev := <-jobs: if err := h.Handle(ctx, ev); err != nil { /* retry policy or DLQ */ } } }
  })
}
```
Shard by `machine_id` hash across N channels when per-machine ordering matters.

## Conventions
- Config: env vars -> struct (`FP_PG_DSN`, `FP_REDIS_ADDR`, `FP_MQTT_URL`), validate at start, fail fast.
- Errors: wrap with `%w`; classify `ErrInvalidEvent` (DLQ, no retry) vs transient (retry with backoff+jitter, cap).
- Time via injected `Clock` interface for testability.
- Health: `/healthz` (process) and `/readyz` (pg+redis+mqtt reachable).
