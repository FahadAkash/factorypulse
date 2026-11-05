# FactoryPulse — Agent Operating Manual

Real-time manufacturing operations intelligence for the fictional **Nova Appliances Ltd**.
Pipeline: **Unity simulator → MQTT (Mosquitto) → Go modular monolith → PostgreSQL + Redis → WebSocket → React/TS**.
Host OS: **Windows 10/11** (PowerShell 5.1+, Docker Desktop/WSL2, Git for Windows). Never emit bash-only or Linux-only commands.

## 0. Token Budget Protocol (HIGHEST PRIORITY — obey before anything else)
Cost is a hard constraint. Every response should be the cheapest one that is correct.
1. **One milestone per session.** Pick the first unchecked item in `docs/ROADMAP.md`. `/clear` between milestones; `/compact` at ~50% context.
2. **Search, then slice.** `Grep -n` / `Glob` first; `Read` with `offset`+`limit` (≤150 lines). Never read a whole file >300 lines. Never re-read a file already in context.
3. **Delegate noise to cheap subagents** (haiku): `explorer` (search), `test-runner` (tests/builds), `cost-auditor`. They return ≤15 lines.
4. **Run the narrowest check**: `go test ./internal/x -run TestName`, not `./...`. Use `scripts/run-quiet.ps1 "<cmd>"` so only failures reach context.
5. **Terse output.** Final answer ≤8 lines. No restating code/diffs you just wrote, no plan recaps, no apologies, no “I will now…”. Use `Edit` (diffs), never rewrite whole files.
6. **No speculative work**: no extra docs, scaffolding, refactors, or “nice to have” features beyond the ask. Ambiguous? Ask ONE question, don't explore.
7. **Think proportionally**: no extended thinking for routine edits; deep thinking only for concurrency, data-model, or architecture decisions.
8. **Skills load on demand** — invoke the matching skill instead of guessing; keep this file lean.
9. If a `[BUDGET]` message appears in context → **terse mode**: no subagents, no exploration, ≤3-line replies, ask the user before continuing.
10. Hooks block generated/vendor paths (Library, Temp, node_modules, bin, obj, .meta, .unity/.prefab…). Don't try to bypass; ask the user.

`TOKEN_COST_REPORT.md` + `docs/token-usage/usage-log.csv` are **auto-written by hooks** after every turn. Never edit them. Read only via `/cost` (head only).

## 1. Repository layout
```
backend/      Go module (modular monolith)  cmd/factorypulse, cmd/loadgen, internal/*, migrations/
simulator/    Unity project (FactoryPulseSim) — C# only; the user runs the Editor
web/          React + TypeScript (Vite, strict)
contracts/    event-schema.json, topics.md  ← SINGLE SOURCE OF TRUTH for Unity↔Go
deploy/       docker-compose.yml, mosquitto.conf, prometheus.yml
docs/         ROADMAP.md, business/PRD/HLD/LLD, token-usage/
scripts/      PowerShell helpers (run-quiet, install, verify-hooks)
.claude/      settings.json, rules/, skills/, agents/, commands/, hooks/, cost/
```

## 2. Commands (PowerShell)
```
docker compose -f deploy/docker-compose.yml up -d            # broker + pg + redis
cd backend; go build ./...; go vet ./...; go test ./... -short
powershell -File scripts/run-quiet.ps1 "cd backend; go test ./internal/ingest/..."
cd web; npm run typecheck; npm test
```
Race detector on Windows needs a C compiler (MinGW/TDM-GCC) → else run `go test -race` in Docker/WSL.
**Never launch Unity.** Edit C# only; tell the user what to press in the Editor.

## 3. Non-negotiable engineering invariants
- **Idempotency**: every event has `event_id`; insert with `ON CONFLICT (event_id) DO NOTHING`; aggregates update **only if the insert happened**, in the same transaction.
- **Ordering**: use `occurred_at` + per-machine `seq`; tolerate late/out-of-order; never let an old event overwrite newer state.
- **Raw events are immutable** and persisted — all metrics must be recomputable by replay.
- **One active incident per machine** — enforced by a partial unique index, not just code.
- **Bounded concurrency**: bounded channels + fixed worker pools; no goroutine-per-message; every goroutine has a `context` and an owner.
- **Redis is a cache** — everything in it must be rebuildable from PostgreSQL.
- **Contract first**: any event change updates `contracts/*`, Go structs, C# classes and a contract test in the same change (skill `event-contract`).
- **Honesty rule**: never invent benchmark/impact numbers. Report `[measured]` only from real runs; label assumptions as assumptions. Nova is fictional; do not claim Walton internals.
- Starved/Blocked machines are **consequences**, not incidents; root cause is the upstream STOPPED machine.

## 4. Definition of Done (per task)
`gofmt` clean (hook does it) · `go vet` · narrow tests pass · new behavior has a test (duplicate + out-of-order cases for ingestion) · no secrets · ROADMAP checkbox ticked · ≤8-line summary to user.

## 5. Where details live (invoke on demand)
Skills: `go-backend-core`, `go-mqtt-ingestion`, `go-postgres-redis`, `go-realtime-observability`, `go-testing-loadtest`, `unity-factory-simulator`, `event-contract`, `factory-domain`, `react-ops-dashboard`, `windows-docker-dev`, `token-budget`, `fde-validation`, `cost-report`.
Subagents: `architect`, `go-backend`, `unity-sim`, `db-engineer`, `frontend-dev`, `test-runner`, `code-reviewer`, `explorer`, `cost-auditor`.
Commands: `/cost`, `/milestone`, `/new-event`, `/replay-check`, `/lean-review`.
Path rules auto-load from `.claude/rules/`.
