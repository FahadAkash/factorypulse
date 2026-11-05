# Windows host rules (always on)
- Use PowerShell syntax: `;` not `&&` (PS 5.1), `$env:VAR`, `Get-Content -Tail`, `Select-String`. No `grep/sed/awk/cat/rm -rf/export`.
- Paths: forward slashes are fine in tools/Go; use `Join-Path` in scripts. Mind MAX_PATH (`core.longpaths=true`).
- Line endings LF (`.gitattributes`). Never write BOM; UTF-8 no BOM.
- Docker Desktop = WSL2; from containers reach host via `host.docker.internal`. Unity (on host) reaches broker at `localhost:1883`.
- Long-running processes (docker compose up, go run, npm run dev) must be started detached/in background and never tailed unboundedly.
- Never run `Unity.exe`; never install global tools without asking.
