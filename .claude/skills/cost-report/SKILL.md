---
name: cost-report
description: How the automatic token/cost ledger works and how to maintain, audit or troubleshoot it. Use when the user asks about spend, wants to change pricing/budget, or the report isn't updating.
---
# Token & cost ledger
- Hooks `Stop`, `SubagentStop`, `PreCompact`, `SessionEnd` run `.claude/hooks/cost-tracker.ps1` -> parses the session transcript (+ subagent transcripts), dedupes by message id, prices per model.
- Files (auto-written, never hand-edit): `TOKEN_COST_REPORT.md` (rewritten each turn), `docs/token-usage/usage-log.csv` (one delta row per turn), `.claude/cost/ledger.json` (state).
- Config: `.claude/cost/pricing.json` (per-million rates by model family; verify against Anthropic's pricing), `.claude/cost/budget.json` (daily/session limits, warn ratio, `hard_block`).
- Budget gate: `UserPromptSubmit` hook injects `[BUDGET]` warnings; with `hard_block:true` prompts are rejected until `#budget-ok` is included.
- Troubleshoot: `.claude/cost/tracker-errors.log`; run `scripts/verify-hooks.ps1`; ensure Git Bash installed; transcript path comes from hook stdin. Reset history: `scripts/reset-cost.ps1`.
- Costs are API-equivalent estimates (subscription users: treat as a usage proxy).
- To read: `/cost` (head only). Never load the whole CSV.
