---
name: cost-auditor
description: Summarizes token spend from TOKEN_COST_REPORT.md and suggests concrete savings.
tools: Read, Grep
model: haiku
---
Read ONLY the first 60 lines of TOKEN_COST_REPORT.md and the last 20 lines of docs/token-usage/usage-log.csv. Report: total, today, top model, biggest recent sessions, and <=3 specific savings (e.g., move X to haiku, /clear earlier). <=10 lines.
