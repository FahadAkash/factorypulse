# Roadmap (one milestone per Claude session; tick when Definition of Done met)
## Phase 1 - Business & domain
- [x] M1 Business discovery doc (Nova Appliances assumptions: 3 lines, 12 machines, 8h shifts, 500 units/shift)
- [x] M2 Workflow docs (current vs FactoryPulse incident flow) + PRD requirements (FR/NFR)
## Phase 2 - Unity simulator
- [x] M3 Unity: FactoryConfig ScriptableObject + 3 lines x 4 connected machines + buffers
- [x] M4 Unity: machine state machine, starve/block propagation, seeded failure injector
- [x] M5 Unity: MQTT publisher (background thread, outbox) + control topic + ground_truth.json
## Phase 3 - Go backend
- [x] M6 Go skeleton, config, logging, graceful shutdown, migrations
- [x] M7 MQTT consumer -> validate -> bounded worker pool
- [x] M8 Idempotent persistence (events table) + Redis machine state
## Phase 4 - Intelligence
- [ ] M9 Minute-bucket production aggregation
- [ ] M10 Downtime incident lifecycle (partial unique index)
- [ ] M11 OEE (machine/line/shift) + quality analytics
- [ ] M12 Bottleneck detection (validated against ground truth) + shift forecast
## Phase 5 - Frontend
- [ ] M13 REST + WebSocket API
- [ ] M14 React dashboard (overview, line, machine) + incident board
## Phase 6 - Reliability & observability
- [ ] M15 Failure-injection + replay reconciliation tests
- [ ] M16 Prometheus/Grafana + OpenTelemetry end-to-end trace
- [ ] M17 Load test (loadgen) -> real benchmark report
## Phase 7 - FDE validation
- [ ] M18 Baseline vs intervention experiments (measured, not assumed)
- [ ] M19 Business impact report + limitations + demo video script
