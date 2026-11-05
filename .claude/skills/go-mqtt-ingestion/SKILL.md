---
name: go-mqtt-ingestion
description: MQTT consumption with paho (Mosquitto), validation, idempotency, out-of-order handling, reconnects and backpressure. Use for any ingestion pipeline work.
---
# MQTT ingestion (github.com/eclipse/paho.mqtt.golang)

## Client setup
- `SetCleanSession(false)` + stable `ClientID` (persistent session), `SetAutoReconnect(true)`, `SetConnectRetry(true)`.
- **Re-subscribe inside `OnConnect`** (reconnect loses subscriptions otherwise).
- Subscribe QoS 1 to `$share/fp/nova/+/+/+/#` (shared sub => horizontal scale; Mosquitto 2 supports it).
- The handler must be **fast**: copy payload, push to bounded channel, return. Never do DB work in the callback. If channel is full: block briefly then metric `ingest_backpressure_total++` (don't drop state/production events).

## Pipeline
`receive -> decode+validate -> dedupe/persist (tx) -> aggregate -> state(redis) -> publish to ws hub`
1. **Validate** against contract (required fields, enum, seq>=1, payload <=16KB). Invalid => `dead_letter_events` table with reason; never crash.
2. **Idempotent persist**:
```go
tag, err := tx.Exec(ctx, `INSERT INTO events(event_id, machine_id, event_type, seq, occurred_at, payload)
  VALUES($1,$2,$3,$4,$5,$6) ON CONFLICT (event_id) DO NOTHING`, ...)
if tag.RowsAffected() == 0 { return ErrDuplicate } // counts must NOT change
// ...aggregate upsert in the SAME tx, then Commit
```
3. **Ordering**: keep `last_seq`/`last_occurred_at` per machine (Redis + PG). State change applies only if `(occurred_at, seq)` is newer. Late events still get stored and still count in production aggregates (bucket by `occurred_at`), but never regress current state. Allow a configurable lateness window (default 5 min) before closing a minute bucket for dashboards; replay can always recompute.
4. **Transient DB failure**: retry with exponential backoff+jitter (cap), keep message in-flight; if exhausted, park in an on-disk/PG retry table. Never ack-and-lose.
5. Heartbeat TTL in Redis marks machines OFFLINE when missing.

## Metrics to emit
`ingest_received_total`, `ingest_invalid_total`, `ingest_duplicate_total`, `ingest_latency_seconds` (occurred->persisted), `ingest_queue_depth`.
