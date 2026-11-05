---
description: Verify metrics reconcile after replaying raw events
---
Use skill `go-testing-loadtest` (replay reconciliation). Run the replay test via `scripts/run-quiet.ps1`. Report: events replayed, duplicates ignored, per-metric diff vs live aggregates (expect 0). <=6 lines.
