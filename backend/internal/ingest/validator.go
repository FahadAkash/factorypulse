package ingest

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/FahadAkash/factorypulse/backend/internal/domain"
)

const (
	MaxPayloadBytes = 16 * 1024 // 16 KB max per contract
)

var (
	ErrPayloadTooLarge   = errors.New("event payload exceeds maximum size of 16KB")
	ErrInvalidJSON       = errors.New("event is not valid JSON")
	ErrUnsupportedSchema = errors.New("unsupported schema_version (expected 1)")
	ErrMissingField      = errors.New("required field is missing")
	ErrInvalidEventType  = errors.New("invalid event_type")
)

type ValidatedEvent struct {
	Envelope domain.EventEnvelope
	Raw      []byte
}

func ValidateEvent(raw []byte) (*ValidatedEvent, error) {
	if len(raw) > MaxPayloadBytes {
		return nil, ErrPayloadTooLarge
	}

	var env domain.EventEnvelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidJSON, err.Error())
	}

	if env.SchemaVersion != 1 {
		return nil, ErrUnsupportedSchema
	}
	if env.EventID == "" {
		return nil, fmt.Errorf("%w: event_id", ErrMissingField)
	}
	if env.FactoryID == "" {
		return nil, fmt.Errorf("%w: factory_id", ErrMissingField)
	}
	if env.LineID == "" {
		return nil, fmt.Errorf("%w: line_id", ErrMissingField)
	}
	if env.MachineID == "" {
		return nil, fmt.Errorf("%w: machine_id", ErrMissingField)
	}
	if env.RunID == "" {
		return nil, fmt.Errorf("%w: run_id", ErrMissingField)
	}
	if env.Seq < 1 {
		return nil, errors.New("seq must be >= 1")
	}
	if env.OccurredAt.IsZero() {
		return nil, fmt.Errorf("%w: occurred_at", ErrMissingField)
	}

	switch env.EventType {
	case domain.EventMachineStateChanged:
		var statePayload domain.StatePayload
		if err := json.Unmarshal(env.Payload, &statePayload); err != nil {
			return nil, fmt.Errorf("invalid machine.state_changed payload: %w", err)
		}
		switch statePayload.State {
		case domain.StateRunning, domain.StateStopped, domain.StateStarved,
			domain.StateBlocked, domain.StateIdle, domain.StateMaintenance:
			// valid state
		default:
			return nil, fmt.Errorf("unknown machine state %q", statePayload.State)
		}
		if statePayload.State == domain.StateStopped && statePayload.ReasonCode == "" {
			return nil, errors.New("reason_code is required when state=STOPPED")
		}

	case domain.EventProductionUnit:
		var prodPayload domain.ProductionPayload
		if err := json.Unmarshal(env.Payload, &prodPayload); err != nil {
			return nil, fmt.Errorf("invalid production.unit payload: %w", err)
		}
		if prodPayload.Outcome != "good" && prodPayload.Outcome != "rejected" {
			return nil, fmt.Errorf("invalid outcome %q (must be 'good' or 'rejected')", prodPayload.Outcome)
		}
		if prodPayload.CycleTimeMs < 1 {
			return nil, errors.New("cycle_time_ms must be >= 1")
		}

	case domain.EventMachineHeartbeat, domain.EventSensorReading:
		// Acceptable standard types
	default:
		return nil, fmt.Errorf("%w: %q", ErrInvalidEventType, env.EventType)
	}

	return &ValidatedEvent{
		Envelope: env,
		Raw:      raw,
	}, nil
}
