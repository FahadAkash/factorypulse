---
paths:
  - "backend/migrations/**"
  - "backend/internal/store/**"
---
# Database rules
- Migrations are forward-only, numbered, idempotent where possible; never edit an applied migration.
- `events` is append-only; `event_id` UNIQUE; index (machine_id, occurred_at); consider daily partitions.
- Active incident uniqueness via `CREATE UNIQUE INDEX ... ON incidents(machine_id) WHERE status IN ('open','acknowledged')`.
- Aggregates are minute-bucket upserts (`ON CONFLICT DO UPDATE`) executed in the same tx as the event insert, only when insert affected 1 row.
- Never `SELECT *` in app code; always bound queries (LIMIT / time range).
