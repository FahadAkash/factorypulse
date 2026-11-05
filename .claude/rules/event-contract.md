---
paths:
  - "contracts/**"
  - "backend/internal/ingest/**"
  - "simulator/**/Events/**"
---
# Contract rules
`contracts/event-schema.json` is the single source of truth. Changing an event requires, in ONE change: schema -> Go struct+validator -> C# class -> contract test (golden JSON fixture parsed by both) -> `topics.md` if topics change. Backward compatible additions only within a schema_version.
