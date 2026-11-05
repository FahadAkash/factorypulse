---
name: windows-docker-dev
description: Windows 10/11 development specifics - PowerShell commands, Docker Desktop/WSL2, Mosquitto/Postgres/Redis compose, Unity-to-broker networking, Git Bash hook requirements, path/line-ending pitfalls.
---
# Windows dev environment
- Claude Code hooks run via **Git Bash** (install Git for Windows) and call `powershell -NoProfile -ExecutionPolicy Bypass -File ...`. Hooks are written for Windows PowerShell 5.1 (no PS7-only syntax).
- Stack up: `docker compose -f deploy/docker-compose.yml up -d` (add `--profile obs` for Prometheus/Grafana). Status: `docker compose -f deploy/docker-compose.yml ps`. Logs always `--tail 50 <service>`.
- Ports: 1883 MQTT, 5432 PG, 6379 Redis, 8080 Go API/metrics, 5173 Vite, 9090 Prom, 3000 Grafana. Port busy => `Get-NetTCPConnection -LocalPort 1883`.
- Unity (host) -> Mosquitto (container) via `localhost:1883` (published port). Go backend on host uses `localhost`; Go in container uses service names.
- Quick MQTT check without extra tools: add `mosquitto_sub` via `docker compose exec mosquitto mosquitto_sub -t "nova/#" -C 5`.
- Env: copy `.env.example` -> `.env` (never commit). Set vars for session: `$env:FP_PG_DSN="postgres://fp:fp_dev_only@localhost:5432/factorypulse"`.
- Go on Windows: `go build ./...`; binaries get `.exe`; `-race` needs MinGW (else use Docker/WSL).
- Execution policy errors: use `-ExecutionPolicy Bypass` per-invocation, never change machine policy.
- Keep the repo outside OneDrive-synced folders (file locking breaks Unity/Go builds).
