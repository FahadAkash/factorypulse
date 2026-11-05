---
name: token-budget
description: Detailed playbook for minimizing token cost in this repo - decision tree for reading, delegating, testing, replying and context hygiene. Use when a task looks large or when a [BUDGET] warning appears.
---
# Token playbook
**Decision tree**
1. Know the file? `Grep -n "symbol" backend/internal` -> `Read` offset/limit around hit. Don't know? `explorer` subagent.
2. Need to run something noisy? `test-runner` subagent or `scripts/run-quiet.ps1`.
3. Task >5 files or >1 milestone? Stop, split, tick ROADMAP only for finished parts.
4. Repeating context? Put stable facts into a skill/rule instead of re-explaining; `/clear` between milestones; `/compact "keep: current milestone, failing test, decisions"` at ~50%.
**Cost levers (high -> low impact)**: fewer files read > narrower tests > haiku delegation > shorter replies > thinking limits > batching tool calls.
**Never**: paste logs/diffs back, rewrite whole files, re-explain plans, read generated dirs, run `./...` when one package suffices, read TOKEN_COST_REPORT.md fully.
**Reply template**: `Done: <what> | Files: a, b | Tests: <cmd> PASS | Risk: <1 line or none>`.
**[BUDGET] terse mode**: no subagents, no exploration, <=3-line replies, ask before multi-step work.
Budgets/pricing: `.claude/cost/budget.json`, `pricing.json`.
