package domain_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/FahadAkash/factorypulse/backend/internal/domain"
)

func TestEventEnvelope_Validation(t *testing.T) {
	validEv := domain.EventEnvelope{
		SchemaVersion: 1,
		EventID:       "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		EventType:     domain.EventMachineStateChanged,
		FactoryID:     "factory-01",
		LineID:        "line-refrig-01",
		MachineID:     "M-REF-01",
		RunID:         "run-001",
		Seq:           1,
		OccurredAt:    time.Now().UTC(),
		Payload:       json.RawMessage(`{"state":"RUNNING"}`),
	}

	if err := validEv.Validate(); err != nil {
		t.Fatalf("expected valid envelope, got error: %v", err)
	}

	invalidEv := validEv
	invalidEv.SchemaVersion = 2
	if err := invalidEv.Validate(); err == nil {
		t.Errorf("expected error for unsupported schema_version, got nil")
	}

	invalidEv = validEv
	invalidEv.EventID = ""
	if err := invalidEv.Validate(); err == nil {
		t.Errorf("expected error for empty event_id, got nil")
	}

	invalidEv = validEv
	invalidEv.Seq = 0
	if err := invalidEv.Validate(); err == nil {
		t.Errorf("expected error for seq <= 0, got nil")
	}
}
