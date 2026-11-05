# FactoryPulse — Product Requirements Document (PRD)

> **Document Status**: Approved Baseline (M2)  
> **System**: FactoryPulse Real-Time Operations Intelligence  
> **Target Deployment**: Nova Appliances Ltd (`factory-01`, 3 lines, 12 machines)

---

## 1. Product Overview & Objectives

FactoryPulse is an industrial IoT and manufacturing operations intelligence platform designed to ingest high-frequency machine events, automate downtime incident tracking, pinpoint line bottlenecks, and deliver real-time operational visibility to plant personnel.

### Target Personas
1. **Line Operator**: Needs immediate feedback on station state, buffer levels, and current cycle pacing.
2. **Maintenance Technician**: Needs instant push notification of machine stoppages with specific fault reason codes, plus one-click acknowledgement and root-cause recording.
3. **Shift Supervisor**: Needs live shift pacing (target vs actual), active incident overview, and early bottleneck alerts to reallocate labor or material.
4. **Plant Operations Director**: Needs shift/line OEE rollups, MTTR/MTBF trends, and audit-grade historical replay reports.

---

## 2. Functional Requirements (FR)

### Ingestion & Stream Processing
- **FR-1.1 (MQTT Envelope Ingestion)**: System shall subscribe to MQTT topics matching `nova/{factory}/{line}/{machine}/+` using a shared consumer group (`$share/fp/nova/+/+/+/#`).
- **FR-1.2 (Schema Validation)**: Ingested messages must be strictly validated against `contracts/event-schema.json` v1 before downstream processing. Malformed messages must be rejected and logged to a dead-letter metric without crashing the worker pool.
- **FR-1.3 (Idempotent Persistence)**: All raw events must be persisted to the `events` table with `ON CONFLICT (event_id) DO NOTHING`. Replayed or duplicated events must yield zero secondary mutations.
- **FR-1.4 (Out-of-Order Handling)**: Per-machine sequence numbers (`seq`) and event timestamps (`occurred_at`) must govern state transitions. Older events arriving after newer events must never overwrite current machine state in Redis or database.

### State & Incident Management
- **FR-2.1 (Machine State Cache)**: Current state (`RUNNING`, `STOPPED`, `STARVED`, `BLOCKED`, `IDLE`, `MAINTENANCE`), last sequence, and last transition timestamp must be cached in Redis with sub-millisecond retrieval.
- **FR-2.2 (Single Active Incident Guarantee)**: When a machine enters `STOPPED`, an incident ticket must be opened if no active incident exists for that machine. Enforced at the database level via a partial unique index (`WHERE status != 'RESOLVED'`).
- **FR-2.3 (Incident Triage Lifecycle)**:
  - Transition from `OPEN` to `ACKNOWLEDGED` when a technician takes ownership.
  - Transition from `ACKNOWLEDGED` to `RESOLVED` automatically when machine emits `RUNNING` (with `occurred_at >= opened_at`) or manually via API with root-cause reason code.
- **FR-2.4 (Consequence Attribution)**: Stations entering `STARVED` or `BLOCKED` must record `upstream_cause_machine_id` pointing to the offending station, without opening secondary incidents.

### Production & OEE Analytics
- **FR-3.1 (Production Aggregation)**: Minute-by-minute buckets of total units, good units, and reject counts per station, line, and product type.
- **FR-3.2 (OEE Calculation)**:
  - Availability ($A$): $\text{Run Time} / \text{Planned Time}$, penalizing `STOPPED`, `STARVED`, and `BLOCKED` states while excluding planned `MAINTENANCE`.
  - Performance ($P$): $(\text{Total Units} \times \text{Ideal Cycle Time}) / \text{Run Time}$, capped strictly at $1.0$.
  - Quality ($Q$): $\text{Good Units} / \text{Total Units}$; guarded against division by zero (returning `n/a`).
- **FR-3.3 (Line-Level OEE Integrity)**: Line OEE must be derived from final assembly output (Station 4 good units vs planned line capacity), not an arithmetic average of machine OEEs.
- **FR-3.4 (Defect Tracking & Pareto)**: Group rejected units by machine, line, shift, and `defect_code` to generate defect Pareto distributions and trigger alarms if defect rate exceeds configured threshold.

### Bottleneck & Shift Intelligence
- **FR-4.1 (Bottleneck Detection)**: Continually compute utilization and starved/blocked ratios across rolling 15-minute windows. Flag the station with highest active utilization whose upstream is blocked and downstream is starved, requiring persistence over $\ge 3$ consecutive windows.
- **FR-4.2 (Shift Forecasting)**:
  $$\text{Projected} = \text{Produced} + (\text{Current Rolling Rate} \times \text{Remaining Shift Hours})$$
  Display status indicator: `ON_TRACK` ($\ge \text{target}$), `AT_RISK` ($90\%\dots99.9\%$), or `BEHIND` ($< 90\%$).

### Frontend & Observability
- **FR-5.1 (Real-Time WebSocket Feed)**: Push live machine state transitions, production unit tallies, and incident state changes to connected browsers within 100ms of persistence.
- **FR-5.2 (Operator & Supervisor Dashboard)**:
  - Factory Overview: Real-time throughput, active incidents, line OEE cards.
  - Line View: 4-station topological diagram showing live state colors, buffer occupancy bars, and bottleneck badges.
  - Incident Board: Kanban/list view of `OPEN`, `ACKNOWLEDGED`, and `RESOLVED` incidents.
- **FR-5.3 (Historical Replay Engine)**: Ability to trigger event replay from raw immutable ledger for a specific `run_id` and audit metric reconciliation.

---

## 3. Non-Functional Requirements (NFR)

| ID | Category | Requirement Specification |
|---|---|---|
| **NFR-1** | **Latency** | End-to-end latency from machine event publication to WebSocket browser render must be $P50 < 100\text{ms}$, $P95 < 500\text{ms}$, $P99 < 1.0\text{s}$. |
| **NFR-2** | **Throughput** | Ingestion pipeline must sustain $\ge 1,000\text{ events/second}$ per single backend instance without unbounded queue growth. |
| **NFR-3** | **Bounded Concurrency** | Zero dynamic per-message goroutine spawning. Ingestion must use fixed worker pools and bounded Go channels. |
| **NFR-4** | **Idempotency & Ordering** | System must maintain $100\%$ metric reconciliation accuracy under simulated duplicate delivery rates up to $20\%$ and disordered arrivals up to $10\%$. |
| **NFR-5** | **Cache Resilience** | Redis is strictly an ephemeral cache. Total Redis loss must not corrupt data; all live state must be rebuildable from PostgreSQL within 30 seconds. |
| **NFR-6** | **Observability** | Expose standard `/metrics` endpoint scraping at 5-second intervals with Prometheus counter/histogram metrics (`events_ingested_total`, `event_to_dashboard_seconds`, `incidents_active`). |
| **NFR-7** | **Platform Compatibility** | Go backend, React frontend, and deploy scripts must natively run on Windows 10/11 (PowerShell 5.1+, WSL2/Docker) and Linux container hosts. |

---

## 4. Verification & Acceptance Criteria Matrix

| Requirement | Verification Method | Pass Criteria |
|---|---|---|
| **FR-1.3 (Idempotency)** | Re-publish identical `events.jsonl` log file twice to MQTT | Total unique rows in `events` table equals exact single-run count; 0 duplicate key errors. |
| **FR-2.2 (Single Incident)** | Inject 5 consecutive `STOPPED` events for `M-REF-01` before resolution | Exactly 1 incident created in `incidents` table. |
| **FR-3.2 (OEE Capping)** | Simulate burst unit outputs under short intervals | Performance metric never exceeds 100.0%. |
| **FR-4.1 (Bottleneck)** | Run Unity scenario with `M-REF-02` artificially slowed | Engine accurately identifies `M-REF-02` as top bottleneck matching `ground_truth.json`. |
| **NFR-1 (Latency)** | Measure timestamp delta between simulator `occurred_at` and UI render | $P95 < 500\text{ms}$ under normal operating load. |
| **NFR-2 (Throughput)** | Execute loadgen tool at 1,000 msgs/sec for 5 minutes | Memory remains bounded; 0 dropped state/production events. |
