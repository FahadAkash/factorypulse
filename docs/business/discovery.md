# Nova Appliances Ltd — Business Discovery & Operating Assumptions

> **Document Status**: Approved Baseline (M1)  
> **Target System**: FactoryPulse Operations Intelligence  
> **Entity**: Nova Appliances Ltd (Fictional Discrete Manufacturer)

---

## 1. Executive Summary & Context

Nova Appliances Ltd is a mid-sized domestic appliance manufacturer producing refrigerators, washing machines, and countertop microwaves. Management is deploying **FactoryPulse** to transition from end-of-shift paper defect logs to real-time operations intelligence, enabling proactive bottleneck identification, automated downtime incident tracking, and live Overall Equipment Effectiveness (OEE) metrics.

### Core Simulation & Manufacturing Scale
- **Factory**: `factory-01` (Central Manufacturing Plant)
- **Production Lines**: 3 dedicated assembly lines
- **Total Machines**: 12 networked production stations (4 stations per line arranged in series)
- **Work Shift**: 8 hours (480 minutes total, 450 minutes net planned operating time)
- **Production Target**: 500 finished good units per line per shift (~62.5 units/hour, ~54s ideal line takt)

---

## 2. Production Topology & Line Specifications

Each line operates as a synchronous-flow assembly cell with intermediate physical buffers (conveyor queues) between stations.

```
Station 1 ──► [Buffer 1] ──► Station 2 ──► [Buffer 2] ──► Station 3 ──► [Buffer 3] ──► Station 4 (QC)
 (Forming)      (Queue)        (Assembly)      (Queue)         (Mount)        (Queue)       (Final Inspection)
```

### 2.1 Line 1: Refrigerator Assembly (`line-refrig-01`)
*Product Type*: `refrigerator-standard` | *Target Shift Output*: 500 units

| Machine ID | Station Name | Role | Ideal Cycle Time | Buffer Out | Capacity | Target Defect Rate |
|---|---|---|---|---|---|---|
| `M-REF-01` | Sheet Press | Outer casing stamping | 50s (50,000 ms) | `B-REF-01` | 10 units | < 0.8% |
| `M-REF-02` | Foaming Station | Polyurethane insulation injection | 52s (52,000 ms) | `B-REF-02` | 8 units | < 1.2% |
| `M-REF-03` | Compressor Mount | Mechanical mounting & wiring | 53s (53,000 ms) | `B-REF-03` | 12 units | < 0.5% |
| `M-REF-04` | Leak & QC Chamber | Helium leak detection & electrical test | 54s (54,000 ms) | Finished Goods | N/A | < 2.0% |

### 2.2 Line 2: Washing Machine Assembly (`line-wash-01`)
*Product Type*: `washer-frontload` | *Target Shift Output*: 500 units

| Machine ID | Station Name | Role | Ideal Cycle Time | Buffer Out | Capacity | Target Defect Rate |
|---|---|---|---|---|---|---|
| `M-WASH-01` | Drum Stamping | Stainless steel drum forming | 48s (48,000 ms) | `B-WASH-01` | 12 units | < 0.6% |
| `M-WASH-02` | Tub Subassembly | Plastic outer tub seal & bearing press | 51s (51,000 ms) | `B-WASH-02` | 8 units | < 1.0% |
| `M-WASH-03` | Direct Drive Mount | Motor, suspension, & harness install | 53s (53,000 ms) | `B-WASH-03` | 10 units | < 0.7% |
| `M-WASH-04` | Spin & Balance Test | Dynamic high-speed vibration & safety test | 54s (54,000 ms) | Finished Goods | N/A | < 1.8% |

### 2.3 Line 3: Microwave Assembly (`line-micro-01`)
*Product Type*: `microwave-countertop` | *Target Shift Output*: 500 units

| Machine ID | Station Name | Role | Ideal Cycle Time | Buffer Out | Capacity | Target Defect Rate |
|---|---|---|---|---|---|---|
| `M-MIC-01` | Chassis Press | Metal cavity and outer wrapper press | 46s (46,000 ms) | `B-MIC-01` | 15 units | < 0.5% |
| `M-MIC-02` | Cavity Welder | Automated seam welding & degrease | 50s (50,000 ms) | `B-MIC-02` | 10 units | < 0.9% |
| `M-MIC-03` | Magnetron Assembly | High-voltage transformer & magnetron mount | 52s (52,000 ms) | `B-MIC-03` | 10 units | < 0.6% |
| `M-MIC-04` | Safety & Leak Test | RF microwave containment & power check | 54s (54,000 ms) | Finished Goods | N/A | < 1.5% |

---

## 3. Shift Structure & Production Scheduling

* **Shift Duration**: 8.0 hours (28,800 seconds / 480 minutes)
* **Planned Breaks / Maintenance**:
  * Shift handover & inspection: 10 minutes at start
  * Scheduled operator break: 20 minutes mid-shift
  * **Net Planned Production Time**: 450 minutes (27,000 seconds)
* **Paced Cycle Target**:
  $$\text{Takt Time} = \frac{27,000\text{ seconds}}{500\text{ units}} = 54.0\text{ seconds/unit}$$
* **Shift Designations**:
  * Shift 1 (Morning): 06:00 – 14:00 UTC
  * Shift 2 (Afternoon): 14:00 – 22:00 UTC
  * Shift 3 (Night): 22:00 – 06:00 UTC

---

## 4. Machine State Taxonomy & Propagation Rules

Each machine reports one of six distinct states per `contracts/event-schema.json`:

| State | Classification | Description | Incident Trigger? |
|---|---|---|---|
| `RUNNING` | Active Output | Machine operating within normal cycle parameters | No (Resolves active incident) |
| `STOPPED` | Direct Loss | Unplanned stoppage due to internal fault or failure | **Yes** (Creates incident ticket) |
| `STARVED` | Indirect Loss | Station idle because upstream buffer is empty ($C = 0$) | No (Consequence of upstream stoppage) |
| `BLOCKED` | Indirect Loss | Station idle because downstream buffer is full ($C = C_{\max}$) | No (Consequence of downstream stoppage) |
| `IDLE` | Planned / Unscheduled | Machine operational but waiting for scheduled shift start | No |
| `MAINTENANCE` | Excluded Loss | Planned PM or tooling changeover; excluded from planned time | No |

### Propagation Rule & Root Cause Attribution
- When a machine transitions to `STOPPED` (root cause), it halts processing.
- Downstream buffers deplete: successive downstream stations transition to `STARVED` and emit `upstream_cause_machine_id = <failed_machine_id>`.
- Upstream buffers accumulate: successive upstream stations transition to `BLOCKED` and emit `upstream_cause_machine_id = <failed_machine_id>`.
- **System Invariant**: Only `STOPPED` triggers an incident in FactoryPulse. Starved/Blocked stations reflect operational propagation loss, not machine defects.

### Standard Failure Modes (`reason_code`)
- `MOTOR_OVERLOAD`: Thermal cutout on mechanical motor or drive
- `JAM`: Mechanical part misfeed or conveyor hangup
- `SENSOR_FAULT`: Optical/proximity sensor signal lost or corrupted
- `TOOL_WEAR`: Tooling tolerance breach requiring technician adjustment
- `HEATER_FAILURE`: Temperature band deviation (foaming, weld or seal station)
- `PNEUMATIC_PRESSURE_DROP`: Low line air pressure fault

---

## 5. Performance Indicators (KPIs) & Domain Formulas

### 5.1 Overall Equipment Effectiveness (OEE)
$$\text{OEE} = \text{Availability} \times \text{Performance} \times \text{Quality}$$

1. **Availability ($A$)**:
   $$A = \frac{\text{Run Time}}{\text{Planned Production Time}}$$
   *Unplanned downtime includes `STOPPED`, `STARVED`, and `BLOCKED` durations. Planned maintenance is excluded from planned time.*
2. **Performance ($P$)**:
   $$P = \min\left(1.0, \frac{\text{Total Units Produced} \times \text{Ideal Cycle Time}}{\text{Run Time}}\right)$$
   *Performance is strictly capped at 1.0 to prevent reporting artifacts from temporary buffer bursts.*
3. **Quality ($Q$)**:
   $$Q = \begin{cases} \frac{\text{Good Units}}{\text{Total Units}}, & \text{if Total Units} > 0 \\ \text{"n/a"}, & \text{if Total Units} = 0 \end{cases}$$
   *Guarded division; never defaults to 0% or 100% when no units have been produced.*

### 5.2 Line-Level OEE Rule
Line OEE is computed directly from **line output** (good finished units exiting Station 4 compared against total planned line capacity), **never** as an arithmetic mean of individual machine OEEs.

### 5.3 Reliability Metrics
- **MTBF (Mean Time Between Failures)**: $\frac{\text{Total Operating Run Time}}{\text{Number of Unplanned Failures}}$
- **MTTR (Mean Time to Resolve)**: $\frac{\text{Total Unplanned Downtime (STOPPED)}}{\text{Number of Resolved Incidents}}$

### 5.4 Shift Forecast Calculation
$$\text{Units Remaining} = \text{Target (500)} - \text{Units Produced}$$
$$\text{Required Rate} = \frac{\text{Units Remaining}}{\text{Hours Left in Shift}}$$
$$\text{Projected Total} = \text{Units Produced} + (\text{Current Rolling Rate} \times \text{Hours Left})$$
* **Status Thresholds**:
  * `ON_TRACK`: $\text{Projected Total} \ge \text{Target}$
  * `AT_RISK`: $\text{Projected Total} \ge 0.90 \times \text{Target}$ and $< \text{Target}$
  * `BEHIND`: $\text{Projected Total} < 0.90 \times \text{Target}$

---

## 6. Bottleneck Detection Model

Station bottleneck score is computed over rolling window $W$ (e.g. 15-minute sliding window):
- Station with highest sustained active utilization ($\text{Run Time} / W$) and lowest relative throughput ratio whose:
  - **Immediate upstream buffer is frequently full** (Upstream station `BLOCKED`).
  - **Immediate downstream buffer is frequently empty** (Downstream station `STARVED`).
- Excludes windows with active `STOPPED` incidents (downtime is tracked via incidents, not constraint theory).
- Requires persistence across $\ge 3$ consecutive evaluation windows before tagging station as an operational bottleneck.

---

## 7. Operational Honesty & Baseline Assumptions

> [!NOTE]
> Nova Appliances Ltd is a fictional business archetype created for architectural validation. Operational baselines reflect simulated engineering assumptions:
> - Baseline line OEE target: **85.0%** ($A \approx 94\%$, $P \approx 95\%$, $Q \ge 95\%$).
> - Theoretical lost production calculations represent **opportunity cost ceiling**, not guaranteed recoverable volume.
> - All empirical metrics reported in system dashboards will be prefixed with `[measured]` from simulation runs, preserving strict boundary between architectural model and synthetic data.

