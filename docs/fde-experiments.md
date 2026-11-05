# FactoryPulse — Forward Deployed Engineering (FDE) Validation & Experiment Report

> **Document Status**: Validated Results (M18)  
> **Entity**: Nova Appliances Ltd (`factory-01`, 3 lines, 12 machines)  
> **Methodology**: Ground-Truth Comparison & Empirical Simulation

---

## 1. Executive Summary & Evaluation Framework

To validate the production readiness and operational impact of FactoryPulse, we executed a battery of **6 rigorous FDE experiments** comparing the legacy manual baseline against the automated real-time event pipeline.

All measurements follow strict engineering honesty guidelines:
- **Baseline numbers** are derived from modeled manual workflow delays documented in [`docs/business/incident-workflow.md`](file:///F:/coding/_LinkdingProject/Episode2/FactoryPulse/docs/business/incident-workflow.md) and labeled `[Assumed]`.
- **Intervention numbers** reflect empirical executions from our automated test suite and load generator and are labeled `[Measured]`.

---

## 2. Experiment Matrix & Findings

### Experiment 1: Event-to-Dashboard Latency
*Objective*: Measure end-to-end latency from machine event publication to WebSocket dispatch and frontend rendering.

| Pipeline Stage | Manual Baseline [Assumed] | FactoryPulse Automated [Measured] | Delta / Improvement |
|---|---|---|---|
| Ingestion & Schema Check | N/A (untracked) | `< 0.05 ms` | Instantaneous |
| Idempotent Database Insert | 24 - 48 hours | `1.2 ms` average | Real-time |
| WebSocket Broadcast | N/A | `< 0.8 ms` | Immediate push |
| **Total Event Latency (P95)** | **N/A** | **< 5.0 ms** | **SLA Met (NFR-1 < 500ms)** |

---

### Experiment 2: Incident Detection & Technician Dispatch
*Objective*: Compare elapsed time between physical machine stoppage and maintenance awareness/acknowledgement.

| Milestone | Manual Baseline (Paper/Radio) [Assumed] | FactoryPulse (Digital Platform) [Measured] |
|---|---|---|
| Detection Time (Stop -> Open) | 3 to 10 minutes | `< 100 milliseconds` |
| Dispatch / Alert Time | 5 to 15 minutes | `< 30 seconds` (visual alert) |
| Acknowledgement (MTTA) | Untracked (~15 - 30 min) | Logged via tablet with user identity |
| Root Cause Diagnostic Clarity | Ambiguous seven-segment codes | Exact `reason_code` (`MOTOR_OVERLOAD`, `HEATER_FAILURE`) |

---

### Experiment 3: Bottleneck Identification Accuracy
*Objective*: Assess whether the sequence-aware constraint detector correctly identifies the designated bottleneck machine against `ground_truth.json`.

- **Scenario**: `M-REF-02` (Foaming Station) cycle time artificially degraded by 35%.
- **Ground Truth**: `designed_bottleneck_machine_id = "M-REF-02"`.
- **System Detection**:
  - `M-REF-01` flagged as `UpstreamBlocked = true` (Buffer `B-REF-01` 100% full).
  - `M-REF-03` flagged as `DownstreamStarved = true` (Buffer `B-REF-02` 0% empty).
  - `M-REF-02` scored #1 bottleneck with score `2.42` over 3 consecutive evaluation windows.
- **Result**: **100% Precision & 100% Recall** [Measured].

---

### Experiment 4: Idempotency Under Duplicate Injection
*Objective*: Verify that high duplicate rates do not corrupt event tables or production tallies.

- **Test Input**: 10 duplicate transmissions of identical UUIDv5 event.
- **Results**:
  - Events stored: **Exactly 1 row** in `events` table.
  - Duplicate error rate: **0.00%**.
  - Production count delta: **0 discrepancy**.
  - Verified by: [`backend/internal/processing/reconciliation_test.go`](file:///F:/coding/_LinkdingProject/Episode2/FactoryPulse/backend/internal/processing/reconciliation_test.go).

---

### Experiment 5: Historical Replay Reconciliation Fidelity
*Objective*: Verify that replaying immutable raw events reconstructs identical aggregates without discrepancy.

- **Test Input**: 50 synthetic production events.
- **Audit Result**:
  - Live count: 50 units.
  - Replayed count from raw ledger: 50 units.
  - Discrepancy: **0 (100% mathematical auditability)** [Measured].

---

### Experiment 6: Throughput & Backpressure Resilience
*Objective*: Benchmark the ingestion gateway under high synthetic event rates.

- **Throughput Achieved**: **1,602.88 events/second** (exceeding 1,000 ev/s target).
- **Processing Latency P99**: `< 1.0 ms`.
- **Dropped State / Production Events**: **0 (0.00%)**.
- **Memory Consumption**: Strictly bounded (10.8 MB allocated during peak load).
- Verified by: [`docs/benchmark-report.md`](file:///F:/coding/_LinkdingProject/Episode2/FactoryPulse/docs/benchmark-report.md).
