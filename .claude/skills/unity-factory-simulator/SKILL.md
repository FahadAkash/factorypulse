---
name: unity-factory-simulator
description: Design and implement the Unity (C#) virtual factory - internally connected demo machines, buffers, starvation/blocking propagation, failure injection, deterministic clock, MQTT publisher and control channel. Use for any work in simulator/.
---
# Unity factory simulator (replaces a Go simulator)

Unity = source of machine events; Go backend only consumes MQTT. Target: Unity 2022.3 LTS or Unity 6, Windows x64, Mono/IL2CPP.

## Scene model (3 lines x 4 connected machines)
```
Line = Sheet Press -> Foaming -> Assembly -> Leak/QC Test      (refrigerator line)
MachineNode -> ProductBuffer -> MachineNode -> ProductBuffer -> ...
```
- `MachineConfig` (ScriptableObject): id, station order, ideal cycle time (ms), cycle jitter %, reject probability, defect codes, MTBF, MTTR.
- `ProductBuffer`: capacity, count, `TryPush/TryPop`. This is what makes machines *internally connected*.
- `MachineNode` state machine: `RUNNING, STOPPED(fault), STARVED(upstream buffer empty), BLOCKED(downstream full), IDLE, MAINTENANCE`.
  Each tick: if fault -> STOPPED; else if no input -> STARVED; else if output full -> BLOCKED; else work one cycle -> push unit -> emit `production.unit`.
- **Propagation**: a STOPPED machine drains downstream buffers -> downstream becomes STARVED; upstream fills -> BLOCKED. Emit `upstream_cause_machine_id` on those states. Only STOPPED creates an incident in the backend.
- `FailureInjector`: MTBF/MTTR via exponential distribution from a seeded RNG, plus scripted faults from scenario JSON, plus manual via control topic.
- Bottleneck scenarios: set one station's ideal cycle slower (e.g. Station B 43/h vs 100/h) and record it in `ground_truth.json`.

## Determinism & time
- `SimClock` advances by fixed step (e.g. 100 ms sim time) inside `FixedUpdate`; `timeScale` multiplies steps/tick (fast-forward 1x..60x) without changing results.
- Per-machine `System.Random(seed = Hash(runSeed, machineId))`. Same seed + scenario => identical event stream (needed for replay tests).
- `occurred_at` = sim clock mapped to a wall-clock epoch chosen at run start.

## Event emission
```csharp
public readonly struct SimEvent {
  public readonly string EventType, LineId, MachineId; public readonly long Seq; public readonly DateTime OccurredAt; public readonly string PayloadJson;
}
// MachineNode increments _seq (per machine), builds envelope per contracts/event-schema.json,
// event_id = UuidV5(runId + ":" + machineId + ":" + seq), enqueues to ConcurrentQueue<SimEvent>.
```
Build JSON with a reused `StringBuilder` or `JsonUtility` DTOs (no per-frame LINQ/allocations).

## MQTT publisher (background thread)
- Library: **MQTTnet** (.NET Standard 2.1 DLLs in `Assets/Plugins/`) - check the user's choice before adding packages.
- Thread: drains the queue, publishes QoS1 to topics in `contracts/topics.md`, reconnect with backoff, bounded outbox (drop-oldest only for heartbeat/sensor; never state/production).
- Broker from `FP_MQTT_HOST`/`FP_MQTT_PORT` (default `localhost:1883`). Windows Firewall may prompt once - allow private networks.
- Subscribe `nova/control/sim`: `start|stop|set_rate{line,factor}|fail_machine{id,reason,duration}|replay{runId}`; reply on `nova/control/ack`.
- Optional QoS-1 duplicate injection toggle (`duplicateRate`, `disorderRate`) to test backend idempotency/ordering.

## Outputs per run
`runs/{runId}/ground_truth.json` (known faults, bottleneck, expected counts) + `events.jsonl` (what was published) for reconciliation.

## Agent boundaries
Edit C# in `Assets/Scripts/**`. Do NOT edit `.unity/.prefab/.meta`/ProjectSettings. Provide the user a short "Editor steps" checklist (create ScriptableObjects, drag prefabs, wire references). Use asmdef files to keep compile times low. Never open Library/Temp.
