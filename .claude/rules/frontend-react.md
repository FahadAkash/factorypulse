---
paths:
  - "web/**/*.ts"
  - "web/**/*.tsx"
---
# React/TS rules
- Vite + React + TypeScript `strict`; no `any`. Server state via TanStack Query; live updates via one WebSocket hook with exponential-backoff reconnect and resync-on-reconnect (refetch snapshot).
- Role views: Manager (target vs actual, OEE), Supervisor (lines, bottleneck), Technician (incident queue), Quality (defect rates).
- Status colors never the only signal (icon + text). Tables virtualized beyond 200 rows. Debounce chart updates (<=1/s).
- Types for API/WS messages generated from/validated against the contract; no ad-hoc shapes.
