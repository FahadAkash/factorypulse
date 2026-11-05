---
name: factory-domain
description: FactoryPulse domain logic - machine states, OEE, downtime incident lifecycle, bottleneck detection, quality analytics, shift forecasting. Use when implementing or reviewing business calculations.
---
# Domain logic

## Machine states
RUNNING, STOPPED (fault), STARVED, BLOCKED, IDLE, MAINTENANCE (planned). STARVED/BLOCKED are consequences; STOPPED is the root cause.

## OEE = A x P x Q (per machine, line, shift)
- Availability = run_time / planned_time. Planned time excludes MAINTENANCE/breaks. STOPPED/STARVED/BLOCKED count against availability (keep separate loss buckets so reports can show *why*).
- Performance = (total_units x ideal_cycle_time) / run_time, **capped at 1.0**.
- Quality = good / total (total=0 => undefined, show "n/a", never 0 or 100).
- Line OEE: compute from line output (last-station good units vs line planned capacity), not an average of machine OEEs.
- Guard every division; store components, not only the product.

## Incident lifecycle
`OPEN (stop event) -> ACKNOWLEDGED (technician) -> RESOLVED (machine RUNNING again; root cause recorded) `
- Created only on STOPPED (not STARVED/BLOCKED). Duplicate/late STOP while active => no new incident (partial unique index).
- Duration = resolved_at - opened_at using event time. Auto-resolve on RUNNING if stopped_at <= running_at.
- Record `source_event_id`, assignee, root_cause (enum + free text), timestamps for ack/discovery metrics.
- Notify via WS `incident` message (email/SMS out of scope for MVP).

## Bottleneck detection (line-level, sequence-aware)
For each station over a window W: `util = run_time/(W)`, `starved_ratio`, `blocked_ratio`, `throughput/capacity`.
Likely bottleneck = station with highest utilization and lowest throughput/capacity whose **upstream is often BLOCKED** and **downstream often STARVED**. Exclude windows with active STOPPED incident (that's downtime, not a constraint). Require persistence (e.g. >= 3 consecutive windows) to flag "repeated slowdown". Output a ranked list with evidence, not a single magic answer. Validate against `ground_truth.json`.

## Quality
Defect rate = rejected/total by machine, line, shift, defect_code; Pareto of defect codes; trend with rolling window; flag when rate exceeds baseline by configured sigma/threshold.

## Shift forecast
`remaining = target - produced; required_rate = remaining / hours_left; current_rate = rolling (e.g. last 30-60 min) good/hr; projected = produced + current_rate x hours_left`.
Status: ON_TRACK if projected >= target; AT_RISK if within 10%; BEHIND otherwise. Show the assumption (rate window) in the UI. Handle shift start (insufficient data) explicitly.

## Impact math honesty
Lost-production math (downtime x ideal rate) = *theoretical opportunity*, not recoverable output. Always label it so.
