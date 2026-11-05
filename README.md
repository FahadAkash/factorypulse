# FactoryPulse - AI-agent workspace (Windows)

Pre-configured Claude Code project: rules, skills, subagents, commands, Windows PowerShell hooks, and an automatic token-cost ledger.

## Setup (once)
1. Install Git for Windows (Claude Code runs hooks through Git Bash), Go, Docker Desktop, Node, Unity Hub, Claude Code.
2. `powershell -ExecutionPolicy Bypass -File scripts\install.ps1`
3. Open this folder: `claude`
4. `powershell -File scripts\verify-hooks.ps1`, then `powershell -File scripts\reset-cost.ps1`
5. Put the Unity project in `simulator\FactoryPulseSim`, the Go module in `backend\`, the React app in `web\`.

## Day-to-day
- Start a session -> hook injects branch + next milestone + today's spend.
- Work **one ROADMAP milestone per session**; `/clear` between milestones; `/compact` at ~50% context.
- `/cost` shows the cost report. Files updated after EVERY turn: `TOKEN_COST_REPORT.md`, `docs\token-usage\usage-log.csv`.
- Tune limits: `.claude\cost\budget.json` (set `hard_block: true` to stop prompts over budget; add `#budget-ok` to override). Rates: `.claude\cost\pricing.json` (verify against Anthropic's pricing).

## Token-saving design
Haiku subagents for search/tests - Sonnet default - thinking capped - hooks block Library/Temp/node_modules/scenes/binaries and log dumps - `run-quiet.ps1` trims command output - skills load on demand - ROADMAP-driven single-milestone sessions.

## Troubleshooting
Hooks silent? Ensure `git --version` works in a normal terminal (Git Bash needed). Tracker errors: `.claude\cost\tracker-errors.log`.
