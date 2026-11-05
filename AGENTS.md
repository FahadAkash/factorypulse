# AGENTS.md (for non-Claude coding agents)
Read `CLAUDE.md` first — it is authoritative. Summary: Windows host (PowerShell), Unity C# simulator → MQTT → Go modular monolith → PostgreSQL/Redis → WebSocket → React/TS.
Be token-frugal: search before reading, read slices, narrowest tests, ≤8-line answers, one ROADMAP milestone per session.
Contract in `contracts/` is the source of truth. Idempotent ingestion, one active incident per machine, no invented benchmark numbers.
