package ingest_test

import (
	"context"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/FahadAkash/factorypulse/backend/internal/ingest"
)

type mockHandler struct {
	mu          sync.Mutex
	handled     []*ingest.ValidatedEvent
	deadLetters [][]byte
}

func (m *mockHandler) Handle(ctx context.Context, event *ingest.ValidatedEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handled = append(m.handled, event)
	return nil
}

func (m *mockHandler) HandleDeadLetter(ctx context.Context, raw []byte, reason string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.deadLetters = append(m.deadLetters, raw)
	return nil
}

func TestValidateEvent_StateChanged(t *testing.T) {
	validRaw := []byte(`{
		"schema_version": 1,
		"event_id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		"event_type": "machine.state_changed",
		"factory_id": "factory-01",
		"line_id": "line-refrig-01",
		"machine_id": "M-REF-01",
		"run_id": "run-001",
		"seq": 1,
		"occurred_at": "2026-11-05T12:00:00Z",
		"payload": {
			"state": "STOPPED",
			"reason_code": "MOTOR_OVERLOAD"
		}
	}`)

	valEv, err := ingest.ValidateEvent(validRaw)
	if err != nil {
		t.Fatalf("expected valid state event, got error: %v", err)
	}
	if valEv.Envelope.MachineID != "M-REF-01" {
		t.Errorf("expected M-REF-01, got %s", valEv.Envelope.MachineID)
	}
}

func TestValidateEvent_StoppedMissingReasonCode(t *testing.T) {
	invalidRaw := []byte(`{
		"schema_version": 1,
		"event_id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		"event_type": "machine.state_changed",
		"factory_id": "factory-01",
		"line_id": "line-refrig-01",
		"machine_id": "M-REF-01",
		"run_id": "run-001",
		"seq": 1,
		"occurred_at": "2026-11-05T12:00:00Z",
		"payload": {
			"state": "STOPPED"
		}
	}`)

	_, err := ingest.ValidateEvent(invalidRaw)
	if err == nil {
		t.Fatalf("expected validation error for STOPPED without reason_code, got nil")
	}
}

func TestValidateEvent_ProductionUnit(t *testing.T) {
	validRaw := []byte(`{
		"schema_version": 1,
		"event_id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		"event_type": "production.unit",
		"factory_id": "factory-01",
		"line_id": "line-refrig-01",
		"machine_id": "M-REF-04",
		"run_id": "run-001",
		"seq": 2,
		"occurred_at": "2026-11-05T12:00:05Z",
		"payload": {
			"outcome": "good",
			"cycle_time_ms": 53500,
			"product_type": "refrigerator-standard"
		}
	}`)

	valEv, err := ingest.ValidateEvent(validRaw)
	if err != nil {
		t.Fatalf("expected valid production unit event, got error: %v", err)
	}
	if valEv.Envelope.Seq != 2 {
		t.Errorf("expected seq 2, got %d", valEv.Envelope.Seq)
	}
}

func TestWorkerPool_ExecutionAndDeadLetter(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	handler := &mockHandler{}
	pool := ingest.NewWorkerPool(4, 100, handler, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = pool.Start(ctx)
	}()

	time.Sleep(50 * time.Millisecond) // Let workers spin up

	validRaw := []byte(`{
		"schema_version": 1,
		"event_id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		"event_type": "machine.state_changed",
		"factory_id": "factory-01",
		"line_id": "line-refrig-01",
		"machine_id": "M-REF-01",
		"run_id": "run-001",
		"seq": 1,
		"occurred_at": "2026-11-05T12:00:00Z",
		"payload": {"state":"RUNNING"}
	}`)

	invalidRaw := []byte(`{"not_an_event": true}`)

	pool.Submit(validRaw)
	pool.Submit(invalidRaw)

	time.Sleep(100 * time.Millisecond)

	handler.mu.Lock()
	defer handler.mu.Unlock()

	if len(handler.handled) != 1 {
		t.Errorf("expected 1 handled event, got %d", len(handler.handled))
	}
	if len(handler.deadLetters) != 1 {
		t.Errorf("expected 1 dead letter, got %d", len(handler.deadLetters))
	}
}

