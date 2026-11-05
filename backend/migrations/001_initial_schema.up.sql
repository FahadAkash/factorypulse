-- FactoryPulse Schema Migration 001: Initial Core Schema

CREATE TABLE IF NOT EXISTS factories (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS lines (
    id TEXT PRIMARY KEY,
    factory_id TEXT NOT NULL REFERENCES factories(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    sequence INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS machines (
    id TEXT PRIMARY KEY,
    line_id TEXT NOT NULL REFERENCES lines(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    station_order INT NOT NULL,
    ideal_cycle_ms INT NOT NULL,
    capacity_per_hour INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Append-only immutable raw events
CREATE TABLE IF NOT EXISTS events (
    event_id UUID PRIMARY KEY,
    factory_id TEXT NOT NULL,
    line_id TEXT NOT NULL,
    machine_id TEXT NOT NULL,
    run_id TEXT NOT NULL,
    event_type TEXT NOT NULL,
    seq BIGINT NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL,
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    trace_id TEXT,
    payload JSONB NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_events_machine_occurred ON events(machine_id, occurred_at);
CREATE INDEX IF NOT EXISTS idx_events_run_id ON events(run_id);

-- Dead letter table for invalid/malformed events
CREATE TABLE IF NOT EXISTS dead_letter_events (
    id BIGSERIAL PRIMARY KEY,
    raw BYTEA NOT NULL,
    reason TEXT NOT NULL,
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Machine operational state intervals for availability calculations
CREATE TABLE IF NOT EXISTS machine_state_intervals (
    id BIGSERIAL PRIMARY KEY,
    machine_id TEXT NOT NULL,
    state TEXT NOT NULL,
    reason_code TEXT,
    upstream_cause_machine_id TEXT,
    started_at TIMESTAMPTZ NOT NULL,
    ended_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_intervals_machine_range ON machine_state_intervals(machine_id, started_at);

-- Minute-bucket aggregates for OEE & production throughput
CREATE TABLE IF NOT EXISTS production_minute (
    machine_id TEXT NOT NULL,
    bucket TIMESTAMPTZ NOT NULL,
    good INT NOT NULL DEFAULT 0,
    rejected INT NOT NULL DEFAULT 0,
    run_ms BIGINT NOT NULL DEFAULT 0,
    PRIMARY KEY (machine_id, bucket)
);

-- Defect aggregation table
CREATE TABLE IF NOT EXISTS defects (
    machine_id TEXT NOT NULL,
    bucket TIMESTAMPTZ NOT NULL,
    defect_code TEXT NOT NULL,
    count INT NOT NULL DEFAULT 0,
    PRIMARY KEY (machine_id, bucket, defect_code)
);

-- Downtime incidents with partial unique index enforcing exactly 1 active incident per machine
CREATE TABLE IF NOT EXISTS incidents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    machine_id TEXT NOT NULL,
    status TEXT NOT NULL, -- 'open', 'acknowledged', 'resolved'
    opened_at TIMESTAMPTZ NOT NULL,
    acknowledged_at TIMESTAMPTZ,
    resolved_at TIMESTAMPTZ,
    assignee TEXT,
    root_cause TEXT,
    source_event_id TEXT NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS incidents_one_active 
ON incidents(machine_id) 
WHERE status IN ('open', 'acknowledged');

CREATE INDEX IF NOT EXISTS idx_incidents_opened_at ON incidents(opened_at DESC);

-- Planned shift definitions
CREATE TABLE IF NOT EXISTS shifts (
    id TEXT PRIMARY KEY,
    line_id TEXT NOT NULL REFERENCES lines(id) ON DELETE CASCADE,
    starts_at TIMESTAMPTZ NOT NULL,
    ends_at TIMESTAMPTZ NOT NULL,
    target_units INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

