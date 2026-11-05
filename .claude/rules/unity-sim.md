---
paths:
  - "simulator/**/Assets/**/*.cs"
---
# Unity simulator rules (C#)
- Simulation is **deterministic**: seeded `System.Random` per machine (seed = hash(runSeed, machineId)); simulation clock separate from `Time.time`; fixed step.
- Machines are **internally connected**: `MachineNode` -> `ProductBuffer` -> next `MachineNode`. Starved = upstream buffer empty; Blocked = downstream buffer full. A fault upstream must propagate as STARVED downstream (emit `upstream_cause_machine_id`).
- No Unity API off the main thread. Main thread enqueues events to a lock-free `ConcurrentQueue<SimEvent>`; a background publisher drains it to MQTT (QoS 1) with reconnect + bounded queue (drop-oldest for heartbeat/sensor ONLY, never for state/production).
- Zero per-frame allocations in `Update/FixedUpdate` (no LINQ, no string concat, reuse lists/StringBuilder).
- Event shape must match `contracts/event-schema.json` exactly; `seq` monotonic per machine; `event_id` = UUIDv5(run_id:machine_id:seq).
- Config via ScriptableObjects (`FactoryConfig`, `MachineConfig`), not hard-coded. Scenarios in JSON; each run writes `ground_truth.json` (known failures/bottleneck) for validation.
- Broker host/port from env vars `FP_MQTT_HOST`/`FP_MQTT_PORT`, default localhost:1883.
- Never hand-edit `.unity/.prefab/.meta`; tell the user which Editor steps to do.
