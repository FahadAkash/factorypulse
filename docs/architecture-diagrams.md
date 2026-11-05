# FactoryPulse — Complete System Architecture & Script Map

This document provides a comprehensive visual and technical map of the entire FactoryPulse project, detailing every script, component, data pipeline, and database relationship.

---

## 1. High-Level End-to-End Pipeline

```mermaid
flowchart LR
    subgraph Unity["1. Virtual Factory Simulator (Unity C#)"]
        direction TB
        FC["FactoryTopology\n(Clock + Outbox)"]
        LT["3 x LineTopology\n(Ref, Wash, Micro)"]
        MN["12 x MachineNode\n(State Machine)"]
        PB["9 x ProductBuffer\n(Conveyor Queues)"]
        FI["FailureInjector\n(Exponential MTBF/MTTR)"]
        MB["MqttSimulationBridge\n(QoS 1 Background Worker)"]
        
        FC --> LT --> MN
        MN <--> PB
        FI -.->|injects STOPPED| MN
        MN -->|SimEvent| FC
        FC --> MB
    end

    subgraph MQTT["2. Message Broker"]
        MOSQ["Eclipse Mosquitto 2.0\n(Port 1883)\nTopics: nova/+/+/+/#"]
    end

    subgraph Go["3. Go Modular Monolith (cmd/factorypulse)"]
        direction TB
        subgraph Ingest["Ingestion Layer"]
            MC["MQTT Consumer\n($share/fp/...)"]
            VAL["Schema Validator\n(16KB limit, contract v1)"]
            WP["Bounded Worker Pool\n(4096 chan, fixed workers)"]
            MC --> VAL --> WP
        end

        subgraph Core["Processing & Storage"]
            PIPE["PipelineHandler"]
            PG_ST["PostgreSQL EventStore\n(ON CONFLICT DO NOTHING)"]
            RD_ST["Redis StateStore\n(Lua ordering guard)"]
            AGG["AggregateStore\n(production_minute, defects)"]
            WP --> PIPE
            PIPE --> PG_ST
            PIPE --> RD_ST
            PIPE --> AGG
        end

        subgraph Intel["Intelligence Services"]
            INC["Incident Service\n(One active per machine)"]
            OEE_C["OEE Calculator\n(A x P x Q, capped at 1.0)"]
            BOT["Bottleneck Detector\n(Sequence & buffer aware)"]
            SHF["Shift Forecaster\n(Rolling hourly pace)"]
            QLT["Quality Analytics\n(Defect Pareto)"]
            PIPE --> INC
            AGG -.-> OEE_C & BOT & SHF & QLT
        end

        subgraph Delivery["Delivery Layer"]
            WSH["WebSocket Hub\n(Bounded 256 send buffers)"]
            REST["REST API Server\n(/api/v1/factory, /incidents)"]
            PROM["Prometheus Telemetry\n(/metrics)"]
            INC & PIPE --> WSH
        end
    end

    subgraph Storage["4. Persistence Layer"]
        PG[("PostgreSQL 16\n(Port 5432)")]
        RD[("Redis 7 Cache\n(Port 6379)")]
    end

    subgraph Frontend["5. Operations Dashboard (React 18 / TS)"]
        direction TB
        HOOK["useLiveSocket()\n(Auto-reconnect + STALE >10s)"]
        HDR["Header\n(Live status pill)"]
        OVR["Overview\n(OEE + Pace)"]
        LNV["LineView\n(Topological stations + buffers)"]
        IBD["IncidentBoard\n(Ack / Resolve actions)"]
        SFC["ShiftForecastCard\n(On track / at risk)"]
        
        HOOK --> HDR & OVR & LNV & IBD & SFC
    end

    subgraph Obs["6. Observability"]
        PRM["Prometheus Scraper\n(5s intervals)"]
        GRAF["Grafana Dashboard\n(Throughput, Latency, Incidents)"]
        PRM --> GRAF
    end

    MB -->|MQTT QoS 1| MOSQ
    MOSQ -->|Shared Sub| MC
    PG_ST & AGG & INC <--> PG
    RD_ST <--> RD
    WSH -->|WS Push <5ms| HOOK
    REST <--> Frontend
    PROM --> PRM
```

---

## 2. Unity Simulator: Architecture & Script Map

**Directory**: `simulator/FactoryPulseSim/Assets/Scripts/`  
**Assembly Definition**: `FactoryPulseSim.asmdef`

```mermaid
classDiagram
    class FactoryConfig {
        +string factoryId
        +LineConfig[] lines
        +int shiftDurationMinutes
        +int plannedProductionMinutes
        +int runSeed
    }

    class LineConfig {
        +string lineId
        +string productType
        +int targetShiftUnits
        +MachineConfig[] machines
        +int[] bufferCapacities
    }

    class MachineConfig {
        +string machineId
        +int stationOrder
        +int idealCycleTimeMs
        +float cycleJitterPercent
        +float rejectProbability
        +string[] defectCodes
        +float mtbfSeconds
        +float mttrSeconds
    }

    class FactoryTopology {
        +SimClock clock
        +LineTopology[] lines
        +ConcurrentQueue~SimEvent~ eventOutbox
        +InitializeFactory()
        +InjectFault()
    }

    class LineTopology {
        +MachineNode[] machines
        +ProductBuffer[] buffers
        +InitializeTopology()
        +InitializeSimulation()
    }

    class MachineNode {
        +MachineConfig config
        +ProductBuffer upstreamBuffer
        +ProductBuffer downstreamBuffer
        +MachineOperationalState currentState
        +FixedUpdate()
        +SetState()
        +EmitProductionEvent()
    }

    class ProductBuffer {
        +string bufferId
        +int capacity
        +int currentCount
        +TryPush() bool
        +TryPop() bool
    }

    class FailureInjector {
        +UpdateFailureStatus() bool
        +InjectManualFault()
    }

    class MqttSimulationBridge {
        +Thread workerThread
        +WorkerLoop()
        +PublishEvent()
    }

    FactoryConfig *-- LineConfig
    LineConfig *-- MachineConfig
    FactoryTopology *-- LineTopology
    LineTopology *-- MachineNode
    LineTopology *-- ProductBuffer
    MachineNode o-- ProductBuffer
    MachineNode *-- FailureInjector
    FactoryTopology --> MqttSimulationBridge
```

### Script Inventory (Simulator)

| Path | Primary Role |
|---|---|
| [`Config/MachineConfig.cs`](file:///F:/coding/_LinkdingProject/Episode2/FactoryPulse/simulator/FactoryPulseSim/Assets/Scripts/Config/MachineConfig.cs) | ScriptableObject defining station cycle time, jitter, defect probabilities, and MTBF/MTTR. |
| [`Config/LineConfig.cs`](file:///F:/coding/_LinkdingProject/Episode2/FactoryPulse/simulator/FactoryPulseSim/Assets/Scripts/Config/LineConfig.cs) | ScriptableObject holding 4 connected machines and 3 intermediate buffer capacities. |
| [`Config/FactoryConfig.cs`](file:///F:/coding/_LinkdingProject/Episode2/FactoryPulse/simulator/FactoryPulseSim/Assets/Scripts/Config/FactoryConfig.cs) | Master ScriptableObject holding 3 line configs, 8h shift params, and seed. |
| [`Core/ProductBuffer.cs`](file:///F:/coding/_LinkdingProject/Episode2/FactoryPulse/simulator/FactoryPulseSim/Assets/Scripts/Core/ProductBuffer.cs) | Inter-station queue managing `TryPush()` and `TryPop()` with starvation/blocking signals. |
| [`Core/MachineNode.cs`](file:///F:/coding/_LinkdingProject/Episode2/FactoryPulse/simulator/FactoryPulseSim/Assets/Scripts/Core/MachineNode.cs) | MonoBehaviour station executing deterministic unit cycles, failure checks, and event emission. |
| [`Core/LineTopology.cs`](file:///F:/coding/_LinkdingProject/Episode2/FactoryPulse/simulator/FactoryPulseSim/Assets/Scripts/Core/LineTopology.cs) | Wires 4 machines in series with 3 buffers: `M1 -> B1 -> M2 -> B2 -> M3 -> B3 -> M4`. |
| [`Core/FactoryTopology.cs`](file:///F:/coding/_LinkdingProject/Episode2/FactoryPulse/simulator/FactoryPulseSim/Assets/Scripts/Core/FactoryTopology.cs) | Root orchestrator managing master simulation clock and outbox queue. |
| [`Core/SimClock.cs`](file:///F:/coding/_LinkdingProject/Episode2/FactoryPulse/simulator/FactoryPulseSim/Assets/Scripts/Core/SimClock.cs) | Deterministic clock advancing by fixed delta with 1x to 60x timescale fast-forward. |
| [`Core/SimEvent.cs`](file:///F:/coding/_LinkdingProject/Episode2/FactoryPulse/simulator/FactoryPulseSim/Assets/Scripts/Core/SimEvent.cs) | Immutable struct envelope matching `contracts/event-schema.json`. |
| [`Core/FailureInjector.cs`](file:///F:/coding/_LinkdingProject/Episode2/FactoryPulse/simulator/FactoryPulseSim/Assets/Scripts/Core/FailureInjector.cs) | Generates exponential MTBF/MTTR distributions from seeded `System.Random`. |
| [`Network/MqttSimulationBridge.cs`](file:///F:/coding/_LinkdingProject/Episode2/FactoryPulse/simulator/FactoryPulseSim/Assets/Scripts/Network/MqttSimulationBridge.cs) | Background thread publisher draining outbox and logging to `runs/{runId}/events.jsonl`. |
| [`Analytics/GroundTruthRecorder.cs`](file:///F:/coding/_LinkdingProject/Episode2/FactoryPulse/simulator/FactoryPulseSim/Assets/Scripts/Analytics/GroundTruthRecorder.cs) | Writes `runs/{runId}/ground_truth.json` for validation and bottleneck checks. |
| [`Utils/UuidV5.cs`](file:///F:/coding/_LinkdingProject/Episode2/FactoryPulse/simulator/FactoryPulseSim/Assets/Scripts/Utils/UuidV5.cs) | RFC 4122 UUIDv5 generator for deterministic `run_id:machine_id:seq` deduplication. |

---

## 3. Go Backend Modular Monolith Architecture

**Module**: `github.com/FahadAkash/factorypulse/backend`  
**Entry Points**: `cmd/factorypulse` (production server), `cmd/loadgen` (load benchmark tool)

```mermaid
flowchart TD
    subgraph CMD["cmd/"]
        MAIN["cmd/factorypulse/main.go\n(Root Wiring & Signal Lifecycle)"]
        LOAD["cmd/loadgen/main.go\n(Benchmark Load Generator)"]
    end

    subgraph Internal["internal/"]
        CFG["config\n(Env parser)"]
        DOM["domain\n(Pure models & types)"]
        
        subgraph Stream["ingest & processing"]
            ING_C["ingest/consumer.go\n(paho.mqtt)"]
            ING_V["ingest/validator.go\n(Contract validator)"]
            ING_W["ingest/worker_pool.go\n(Bounded pool)"]
            PROC["processing/pipeline.go\n(PipelineHandler)"]
        end

        subgraph Store["store/"]
            PG_EV["store/pg/events.go\n(EventStore)"]
            PG_AG["store/pg/aggregates.go\n(AggregateStore)"]
            RD_ST["store/redis/state.go\n(StateStore + Lua)"]
        end

        subgraph Services["Domain Intelligence"]
            INC["incident/service.go\n(IncidentService)"]
            OEE["oee/calculator.go\n(OEE Math)"]
            BOT["bottleneck/detector.go\n(Constraint Ranks)"]
            SHF["shift/forecast.go\n(Pace Forecaster)"]
            QLT["quality/analytics.go\n(Pareto Analysis)"]
        end

        subgraph WebLayer["api & ws"]
            API_S["api/server.go\n(HTTP Server)"]
            API_R["api/routes.go\n(Route Handlers)"]
            WS_H["ws/hub.go\n(WebSocket Hub)"]
        end

        subgraph Obs["telemetry"]
            TEL_L["telemetry/logging.go\n(slog JSON)"]
            TEL_M["telemetry/metrics.go\n(Prometheus gauges/histograms)"]
        end
    end

    MAIN --> CFG & TEL_L & WS_H & INC & API_S
    ING_C --> ING_W
    ING_W --> ING_V --> PROC
    PROC --> PG_EV & PG_AG & RD_ST & INC
    API_S --> API_R
    API_R --> WS_H & INC
```

### Script Inventory (Go Backend)

| Package | File | Key Functions & Responsibilities |
|---|---|---|
| `cmd/factorypulse` | [`main.go`](file:///F:/coding/_LinkdingProject/Episode2/FactoryPulse/backend/cmd/factorypulse/main.go) | Service wiring, `errgroup` lifecycle, and OS signal shutdown. |
| `cmd/loadgen` | [`main.go`](file:///F:/coding/_LinkdingProject/Episode2/FactoryPulse/backend/cmd/loadgen/main.go) | Ingestion benchmark generator (measured 1,602.88 ev/s). |
| `internal/config` | [`config.go`](file:///F:/coding/_LinkdingProject/Episode2/FactoryPulse/backend/internal/config/config.go) | Validates `FP_PORT`, `FP_PG_DSN`, `FP_REDIS_ADDR`, `FP_MQTT_URL`. |
| `internal/domain` | [`models.go`](file:///F:/coding/_LinkdingProject/Episode2/FactoryPulse/backend/internal/domain/models.go) | `EventEnvelope`, `StatePayload`, `ProductionPayload`, `Incident`. |
| `internal/ingest` | [`consumer.go`](file:///F:/coding/_LinkdingProject/Episode2/FactoryPulse/backend/internal/ingest/consumer.go) | MQTT client with `$share/fp/nova/+/+/+/#` subscription and auto-reconnect. |
| `internal/ingest` | [`validator.go`](file:///F:/coding/_LinkdingProject/Episode2/FactoryPulse/backend/internal/ingest/validator.go) | Contract validator (16KB cap, valid states, required fields). |
| `internal/ingest` | [`worker_pool.go`](file:///F:/coding/_LinkdingProject/Episode2/FactoryPulse/backend/internal/ingest/worker_pool.go) | Fixed worker pool reading from bounded 4,096 job channel with backpressure. |
| `internal/processing` | [`pipeline.go`](file:///F:/coding/_LinkdingProject/Episode2/FactoryPulse/backend/internal/processing/pipeline.go) | Orchestrates idempotent PG insert, Redis cache update, and incident trigger. |
| `internal/store/pg` | [`events.go`](file:///F:/coding/_LinkdingProject/Episode2/FactoryPulse/backend/internal/store/pg/events.go) | `SaveEvent()` with `ON CONFLICT (event_id) DO NOTHING` + Dead letter table. |
| `internal/store/pg` | [`aggregates.go`](file:///F:/coding/_LinkdingProject/Episode2/FactoryPulse/backend/internal/store/pg/aggregates.go) | Minute-bucket upsert (`production_minute` and `defects`). |
| `internal/store/redis` | [`state.go`](file:///F:/coding/_LinkdingProject/Episode2/FactoryPulse/backend/internal/store/redis/state.go) | Atomic Lua script preventing out-of-order state regression (`fp:m:{id}:state`). |
| `internal/incident` | [`service.go`](file:///F:/coding/_LinkdingProject/Episode2/FactoryPulse/backend/internal/incident/service.go) | Partial unique index incident lifecycle (`Open`, `Acknowledge`, `Resolve`, `AutoResolve`). |
| `internal/oee` | [`calculator.go`](file:///F:/coding/_LinkdingProject/Episode2/FactoryPulse/backend/internal/oee/calculator.go) | Computes $A \times P \times Q$ with performance strictly capped at 1.0 and zero-division guards. |
| `internal/quality` | [`analytics.go`](file:///F:/coding/_LinkdingProject/Episode2/FactoryPulse/backend/internal/quality/analytics.go) | Defect rate calculation, Pareto descending sort, and threshold alerts. |
| `internal/bottleneck` | [`detector.go`](file:///F:/coding/_LinkdingProject/Episode2/FactoryPulse/backend/internal/bottleneck/detector.go) | Constraint detection based on upstream blocked and downstream starved buffers. |
| `internal/shift` | [`forecast.go`](file:///F:/coding/_LinkdingProject/Episode2/FactoryPulse/backend/internal/shift/forecast.go) | Computes projected shift output, pacing, and `ON_TRACK`/`AT_RISK`/`BEHIND` status. |
| `internal/ws` | [`hub.go`](file:///F:/coding/_LinkdingProject/Episode2/FactoryPulse/backend/internal/ws/hub.go) | Bounded WebSocket broadcast hub (256-message client buffer, slow-client drop protection). |
| `internal/api` | [`server.go`](file:///F:/coding/_LinkdingProject/Episode2/FactoryPulse/backend/internal/api/server.go) | HTTP server with graceful shutdown, `/healthz`, and `/readyz`. |
| `internal/api` | [`routes.go`](file:///F:/coding/_LinkdingProject/Episode2/FactoryPulse/backend/internal/api/routes.go) | Endpoints: `/ws`, `/api/v1/factory`, `/api/v1/incidents`, `/api/v1/shift/forecast`, `/metrics`. |
| `internal/telemetry` | [`logging.go`](file:///F:/coding/_LinkdingProject/Episode2/FactoryPulse/backend/internal/telemetry/logging.go) | Structured JSON `slog` configuration. |
| `internal/telemetry` | [`metrics.go`](file:///F:/coding/_LinkdingProject/Episode2/FactoryPulse/backend/internal/telemetry/metrics.go) | Prometheus metrics (`ingest_received_total`, `ingest_latency_seconds`, etc.). |

---

## 4. Frontend Dashboard: Component Tree

**Directory**: `web/src/`  
**Tech**: React 18 + TypeScript strict + Vite

```mermaid
graph TD
    APP["App.tsx\n(Master Dashboard & Live State)"]
    HOOK["useLiveSocket.ts\n(WebSocket with auto-reconnect)"]
    HDR["Header.tsx\n(Plant title + Live status badge)"]
    OVR["Overview.tsx\n(Factory pace, Line OEE cards, Active incidents)"]
    SFC["ShiftForecastCard.tsx\n(Target, Projected, Current vs Required Pace)"]
    IBD["IncidentBoard.tsx\n(Active tickets, Acknowledge & Resolve actions)"]
    LNV["3 x LineView.tsx\n(Sequential Stations & Intermediate Buffers)"]
    
    APP --> HOOK
    APP --> HDR
    APP --> OVR
    APP --> SFC
    APP --> IBD
    APP --> LNV
```

### Script Inventory (Frontend)

| File | Purpose |
|---|---|
| [`src/types/index.ts`](file:///F:/coding/_LinkdingProject/Episode2/FactoryPulse/web/src/types/index.ts) | TypeScript interfaces matching Go domain models (`StationData`, `LineData`, `IncidentData`, `ShiftForecastData`). |
| [`src/hooks/useLiveSocket.ts`](file:///F:/coding/_LinkdingProject/Episode2/FactoryPulse/web/src/hooks/useLiveSocket.ts) | Custom hook managing WebSocket lifecycle, backoff reconnect (1s..30s), and `STALE (>10s)` detection. |
| [`src/components/Header.tsx`](file:///F:/coding/_LinkdingProject/Episode2/FactoryPulse/web/src/components/Header.tsx) | Top bar showing facility identity, current shift, and dynamic status pill (`LIVE` / `RECONNECTING` / `STALE`). |
| [`src/components/Overview.tsx`](file:///F:/coding/_LinkdingProject/Episode2/FactoryPulse/web/src/components/Overview.tsx) | High-level KPI cards: Shift production volume, Line 1-3 OEE rollups, active incident counter. |
| [`src/components/LineView.tsx`](file:///F:/coding/_LinkdingProject/Episode2/FactoryPulse/web/src/components/LineView.tsx) | Topological station cards with state colors (`RUNNING`, `STOPPED`, `STARVED`, `BLOCKED`), buffer bars, and bottleneck badges. |
| [`src/components/IncidentBoard.tsx`](file:///F:/coding/_LinkdingProject/Episode2/FactoryPulse/web/src/components/IncidentBoard.tsx) | Operational incident queue with status badges and direct `Acknowledge` and `Resolve` action handlers. |
| [`src/components/ShiftForecastCard.tsx`](file:///F:/coding/_LinkdingProject/Episode2/FactoryPulse/web/src/components/ShiftForecastCard.tsx) | Shift forecasting widget showing required pace vs current pace and `ON_TRACK` / `AT_RISK` / `BEHIND` indicators. |
| [`src/App.tsx`](file:///F:/coding/_LinkdingProject/Episode2/FactoryPulse/web/src/App.tsx) | Root application tying together components, initial data loading, and live websocket updates. |
| [`src/App.test.ts`](file:///F:/coding/_LinkdingProject/Episode2/FactoryPulse/web/src/App.test.ts) | Vitest test suite validating plant totals, bottleneck logic, and status thresholds. |

---

## 5. PostgreSQL Database Schema & Invariants

**Migration**: [`backend/migrations/001_initial_schema.up.sql`](file:///F:/coding/_LinkdingProject/Episode2/FactoryPulse/backend/migrations/001_initial_schema.up.sql)

```mermaid
erDiagram
    factories ||--o{ lines : "has"
    lines ||--o{ machines : "contains"
    lines ||--o{ shifts : "schedules"
    machines ||--o{ events : "generates"
    machines ||--o{ machine_state_intervals : "logs"
    machines ||--o{ production_minute : "aggregates"
    machines ||--o{ defects : "records"
    machines ||--o{ incidents : "triggers"

    factories {
        text id PK
        text name
        timestamptz created_at
    }

    lines {
        text id PK
        text factory_id FK
        text name
        int sequence
    }

    machines {
        text id PK
        text line_id FK
        text name
        int station_order
        int ideal_cycle_ms
        int capacity_per_hour
    }

    events {
        uuid event_id PK
        text factory_id
        text line_id
        text machine_id
        text run_id
        text event_type
        bigint seq
        timestamptz occurred_at
        timestamptz received_at
        text trace_id
        jsonb payload
    }

    incidents {
        uuid id PK
        text machine_id
        text status
        timestamptz opened_at
        timestamptz acknowledged_at
        timestamptz resolved_at
        text assignee
        text root_cause
        text source_event_id
    }

    production_minute {
        text machine_id PK
        timestamptz bucket PK
        int good
        int rejected
        bigint run_ms
    }

    defects {
        text machine_id PK
        timestamptz bucket PK
        text defect_code PK
        int count
    }
```

> [!IMPORTANT]
> **Key Database Invariant**:  
> `CREATE UNIQUE INDEX incidents_one_active ON incidents(machine_id) WHERE status IN ('open', 'acknowledged');`  
> Guarantees that duplicate `STOPPED` events or cascade failures never create duplicate active incident tickets for the same machine.

---

## 6. Incident Lifecycle State Machine

```mermaid
stateDiagram-v2
    [*] --> RUNNING : Machine active & processing
    
    RUNNING --> STOPPED : Physical Fault (e.g. MOTOR_OVERLOAD)\n[Emits machine.state_changed]
    
    state STOPPED {
        [*] --> OPEN : Event received by Go backend\n[Creates incident ticket]
        OPEN --> ACKNOWLEDGED : Technician clicks Acknowledge\n[Assignee recorded]
        ACKNOWLEDGED --> RESOLVED : Machine restarts or manual resolution\n[Auto-resolves on RUNNING]
        OPEN --> RESOLVED : Machine restarts without manual ack
    }
    
    STOPPED --> RUNNING : Fault Repaired\n[Auto-resolves incident]
    
    RUNNING --> STARVED : Upstream buffer empty (C = 0)\n[Carries upstream_cause_machine_id]
    STARVED --> RUNNING : Upstream delivers unit
    
    RUNNING --> BLOCKED : Downstream buffer full (C = C_max)\n[Carries upstream_cause_machine_id]
    BLOCKED --> RUNNING : Downstream pulls unit
```
