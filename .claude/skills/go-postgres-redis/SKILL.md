---
name: go-postgres-redis
description: PostgreSQL schema, pgx usage, migrations, partitioning, minute-bucket aggregates, incident uniqueness and Redis state design. Use for DB or cache work.
---
# PostgreSQL + Redis

## Core tables (UTC timestamptz everywhere)
- `factories, lines(id, factory_id, name, sequence), machines(id, line_id, name, station_order, ideal_cycle_ms, capacity_per_hour)`
- `events(event_id uuid PK, machine_id, event_type, seq, occurred_at, received_at, payload jsonb)` - append-only; index `(machine_id, occurred_at)`; partition by day when >10M rows.
- `dead_letter_events(id, raw bytea, reason, received_at)`
- `machine_state_intervals(machine_id, state, reason_code, started_at, ended_at)` - source for availability/OEE.
- `production_minute(machine_id, bucket timestamptz, good int, rejected int, run_ms bigint, PRIMARY KEY(machine_id,bucket))`
- `defects(machine_id, bucket, defect_code, count)`
- `incidents(id, machine_id, status, opened_at, acknowledged_at, resolved_at, assignee, root_cause, source_event_id)`
- `shifts(id, line_id, starts_at, ends_at, target_units)`

## Invariants in SQL
```sql
CREATE UNIQUE INDEX incidents_one_active ON incidents(machine_id) WHERE status IN ('open','acknowledged');
```
Duplicate STOP => `INSERT ... ON CONFLICT DO NOTHING` returns 0 rows => no second incident.

## Aggregate upsert (same tx as event insert)
```sql
INSERT INTO production_minute(machine_id,bucket,good,rejected)
VALUES($1,date_trunc('minute',$2),$3,$4)
ON CONFLICT (machine_id,bucket) DO UPDATE
SET good=production_minute.good+EXCLUDED.good, rejected=production_minute.rejected+EXCLUDED.rejected;
```
Hourly/shift views roll up from minute buckets - never scan raw events for dashboards.

## pgx
`pgxpool` with `MaxConns` sized to workers; `pgx.BeginFunc` for tx; batch (`pgx.Batch`) for bulk loads; `COPY` for replay/loadgen. Always set statement/context timeouts.

## Redis (cache only, rebuildable from PG)
- `fp:m:{machine}:state` HASH {state, since, reason, last_seq, last_at}
- `fp:m:{machine}:hb` STRING with TTL 15s (offline detection)
- `fp:line:{line}:shift` HASH counters (good, rejected) for fast dashboard
- Fan-out to WS: Pub/Sub channel `fp:updates` (lossy OK; snapshot endpoint resyncs).
- Startup: rebuild hot keys from PG if missing.
