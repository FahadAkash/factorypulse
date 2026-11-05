---
name: go-realtime-observability
description: WebSocket hub design, Prometheus metrics, OpenTelemetry tracing from MQTT to dashboard, and Grafana provisioning. Use for real-time delivery or observability work.
---
# Real-time delivery + observability

## WebSocket (github.com/coder/websocket)
- One `Hub` goroutine owns client set. Each client has a **bounded** send channel (e.g. 256); if full -> disconnect that client (slow consumer must never block the hub).
- Message envelope: `{type, ts, seq, data, trace_id}`; types: `machine_state`, `production_delta`, `incident`, `line_summary`.
- Throttle: coalesce per-machine updates to <=4/s. Client resync: on (re)connect, fetch REST snapshot then apply WS deltas with `seq` guard.
- Ping/pong + read limit + origin check (dev: localhost only).

## Tracing (OpenTelemetry Go)
Span chain: `mqtt.receive -> ingest.validate -> pg.tx(insert+aggregate) -> redis.set -> ws.broadcast`.
Carry `trace_id` in event envelope (optional field) so Unity-originated events can be correlated. OTLP exporter -> collector/Jaeger (add to compose under `obs` profile when needed).

## Metrics (Prometheus client_golang) - names are part of the contract
`ingest_received_total`, `ingest_duplicate_total`, `ingest_invalid_total`, `ingest_latency_seconds` (histogram), `db_tx_duration_seconds`, `ws_clients`, `ws_dropped_clients_total`, `incident_open` (gauge), `event_to_dashboard_seconds` (histogram; key FDE metric).
Expose `/metrics` on :8080. Keep label cardinality low (line, event_type - not machine_id unless <100 machines).

## Grafana
Provision datasource + dashboards from files in `deploy/grafana/` (no click-ops). Panels: throughput, P95/P99 latency, duplicates, queue depth, open incidents.
