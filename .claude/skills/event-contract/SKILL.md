---
name: event-contract
description: Keeps the MQTT event contract synchronized across Unity (C#), Go and tests. Use whenever an event type, field, topic or schema version changes.
---
# Event contract workflow
Source of truth: `contracts/event-schema.json` + `contracts/topics.md`.
Checklist (all in one change):
1. Edit schema (additive within version; breaking => `schema_version` bump + Go supports N and N-1).
2. Go: struct + `Validate()` in `backend/internal/ingest`; update decoder switch.
3. C#: DTO + emitter in `simulator/**/Assets/Scripts/Events`.
4. Golden fixture in `contracts/fixtures/<event_type>.json`; Go test parses+validates it; C# test serializes and compares field set.
5. Update `topics.md` if topic/QoS/retain changes.
6. Mention migration impact (DB columns/aggregates) to `db-engineer`.
Rules: `event_id` deterministic UUIDv5(run:machine:seq); `occurred_at` RFC3339 UTC; unknown fields rejected by Go validator (strict) in tests, tolerated in prod decode (log+metric).
