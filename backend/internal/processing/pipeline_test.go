package processing_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/FahadAkash/factorypulse/backend/internal/domain"
	"github.com/FahadAkash/factorypulse/backend/internal/ingest"
	"github.com/FahadAkash/factorypulse/backend/internal/processing"
)

func TestPipelineHandler_NilStoresSafe(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	handler := processing.NewPipelineHandler(nil, nil, nil, nil, logger)

	valEv := &ingest.ValidatedEvent{
		Envelope: domain.EventEnvelope{
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
		},
	}

	err := handler.Handle(context.Background(), valEv)
	if err != nil {
		t.Fatalf("expected nil store handler to complete safely, got: %v", err)
	}

	err = handler.HandleDeadLetter(context.Background(), []byte("bad"), "reason")
	if err != nil {
		t.Fatalf("expected nil store dead letter to complete safely, got: %v", err)
	}
}
