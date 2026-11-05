---
name: fde-validation
description: FDE-style validation for FactoryPulse - customer-discovery docs, baseline vs intervention experiments, ground-truth accuracy checks and the business impact report, with strict honesty rules.
---
# FDE validation
Deliverables: business discovery (Nova assumptions), current vs proposed incident workflow, experiment plan, measured results, limitations, impact report.
**Honesty rules**: Nova is fictional; assumptions are labeled assumptions; don't claim Walton internals; tables in docs/prompts with "illustrative" numbers must be replaced by measurements; "recoverable" production != theoretical lost units.
**Experiments (measure with Unity ground truth)**
| Experiment | Measure |
|---|---|
| Event->dashboard latency | `event_to_dashboard_seconds` P50/P95/P99 |
| Incident detection/ack | opened_at - stop occurred_at; ack - opened_at (simulated technician delays: baseline manual vs system) |
| Bottleneck accuracy | precision/recall vs `ground_truth.json` over N seeded runs |
| Idempotency | counts under X% duplicate injection == 0 error |
| Replay | per-metric diff after replay == 0 |
| Load | throughput, P99, loss, CPU/RAM per rate step |
Baseline simulation: model manual notification delays with documented distributions (assumptions), run N seeds, report mean +/- spread. State test machine specs. Report limitations (simulated data, no real operators).
