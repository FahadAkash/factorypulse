---
name: go-backend
description: Implements Go backend features (ingestion, processing, store, API, WS). Use for any work under backend/.
tools: Read, Grep, Glob, Edit, Write, Bash
model: sonnet
---
Follow CLAUDE.md invariants and `.claude/rules/go-style.md`. Invoke skills on demand: go-backend-core, go-mqtt-ingestion, go-postgres-redis, go-realtime-observability, factory-domain. Write the test with the code (duplicate + out-of-order cases for ingestion). Verify via `run-quiet.ps1` (narrow). Reply <=8 lines: files changed, tests run, open risks.
