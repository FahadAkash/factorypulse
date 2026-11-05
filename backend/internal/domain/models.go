package domain

import (
	"encoding/json"
	"errors"
	"time"
)

type EventType string

const (
	EventMachineStateChanged EventType = "machine.state_changed"
	EventProductionUnit      EventType = "production.unit"
	EventMachineHeartbeat    EventType = "machine.heartbeat"
	EventSensorReading       EventType = "sensor.reading"
)

type MachineState string

const (
	StateRunning     MachineState = "RUNNING"
	StateStopped     MachineState = "STOPPED"
	StateStarved     MachineState = "STARVED"
	StateBlocked     MachineState = "BLOCKED"
	StateIdle        MachineState = "IDLE"
	StateMaintenance MachineState = "MAINTENANCE"
)

type IncidentStatus string

const (
	IncidentStatusOpen         IncidentStatus = "open"
	IncidentStatusAcknowledged IncidentStatus = "acknowledged"
	IncidentStatusResolved     IncidentStatus = "resolved"
)

type EventEnvelope struct {
	SchemaVersion int             `json:"schema_version"`
	EventID       string          `json:"event_id"`
	EventType     EventType       `json:"event_type"`
	FactoryID     string          `json:"factory_id"`
	LineID        string          `json:"line_id"`
	MachineID     string          `json:"machine_id"`
	RunID         string          `json:"run_id"`
	Seq           int64           `json:"seq"`
	OccurredAt    time.Time       `json:"occurred_at"`
	TraceID       string          `json:"trace_id,omitempty"`
	Payload       json.RawMessage `json:"payload"`
}

type StatePayload struct {
	State                  MachineState `json:"state"`
	ReasonCode             string       `json:"reason_code,omitempty"`
	UpstreamCauseMachineID string       `json:"upstream_cause_machine_id,omitempty"`
}

type ProductionPayload struct {
	Outcome     string `json:"outcome"` // "good" or "rejected"
	DefectCode  string `json:"defect_code,omitempty"`
	CycleTimeMs int    `json:"cycle_time_ms"`
	ProductType string `json:"product_type,omitempty"`
}

type Incident struct {
	ID             string         `json:"id"`
	MachineID      string         `json:"machine_id"`
	Status         IncidentStatus `json:"status"`
	OpenedAt       time.Time      `json:"opened_at"`
	AcknowledgedAt *time.Time     `json:"acknowledged_at,omitempty"`
	ResolvedAt     *time.Time     `json:"resolved_at,omitempty"`
	Assignee       string         `json:"assignee,omitempty"`
	RootCause      string         `json:"root_cause,omitempty"`
	SourceEventID  string         `json:"source_event_id"`
}

func (e *EventEnvelope) Validate() error {
	if e.SchemaVersion != 1 {
		return errors.New("unsupported schema_version, must be 1")
	}
	if e.EventID == "" {
		return errors.New("event_id cannot be empty")
	}
	if e.EventType == "" {
		return errors.New("event_type cannot be empty")
	}
	if e.MachineID == "" {
		return errors.New("machine_id cannot be empty")
	}
	if e.Seq <= 0 {
		return errors.New("seq must be positive integer")
	}
	if e.OccurredAt.IsZero() {
		return errors.New("occurred_at timestamp is required")
	}
	return nil
}

