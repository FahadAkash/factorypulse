---
name: react-ops-dashboard
description: React + TypeScript operations dashboard - role views, WebSocket client with resync, charts and incident board. Use for any web/ work.
---
# Ops dashboard
- Stack: Vite, React 18, TS strict, TanStack Query, lightweight chart lib (uPlot or Recharts), plain CSS modules/Tailwind. No heavy UI kits unless asked.
- Views: **Overview** (factory target vs actual, OEE, open incidents), **Line** (stations in sequence with state color, throughput vs capacity, bottleneck badge), **Machine** (state timeline, OEE A/P/Q, defects), **Incidents** (queue: open/ack/resolved, ack + root-cause form), **Shift report** (forecast + hourly table), **Quality**.
- `useLiveSocket()` hook: connect, exponential backoff (1s..30s), on reconnect refetch snapshots then apply deltas ignoring stale `seq`. Show "LIVE / RECONNECTING / STALE (>10s)" badge - managers must know if data is stale.
- Perf: memoize rows, batch state updates per animation frame, virtualize long tables, cap chart points.
- Status = color + icon + text. Time shown in factory local time (Asia/Dhaka) while API stays UTC.
- Test: Vitest for reducers/formatters; no snapshot floods.
