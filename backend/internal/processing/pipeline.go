package processing

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"

	"github.com/FahadAkash/factorypulse/backend/internal/domain"
	"github.com/FahadAkash/factorypulse/backend/internal/ingest"
	"github.com/FahadAkash/factorypulse/backend/internal/store/pg"
	"github.com/FahadAkash/factorypulse/backend/internal/store/redis"
)

type PipelineHandler struct {
	events *pg.EventStore
	states *redis.StateStore
	logger *slog.Logger
}

func NewPipelineHandler(events *pg.EventStore, states *redis.StateStore, logger *slog.Logger) *PipelineHandler {
	return &PipelineHandler{
		events: events,
		states: states,
		logger: logger,
	}
}

func (h *PipelineHandler) Handle(ctx context.Context, valEv *ingest.ValidatedEvent) error {
	env := &valEv.Envelope

	// 1. Idempotent persistence in PostgreSQL
	if h.events != nil {
		inserted, err := h.events.SaveEvent(ctx, env)
		if err != nil {
			if errors.Is(err, pg.ErrDuplicateEvent) {
				h.logger.Debug("Duplicate event received; skipping secondary processing", "event_id", env.EventID)
				return nil
			}
			return err
		}
		if !inserted {
			return nil
		}
	}

	// 2. Machine state update in Redis
	if env.EventType == domain.EventMachineStateChanged && h.states != nil {
		var statePayload domain.StatePayload
		if err := json.Unmarshal(env.Payload, &statePayload); err == nil {
			updated, err := h.states.UpdateMachineState(
				ctx,
				env.MachineID,
				string(statePayload.State),
				statePayload.ReasonCode,
				env.Seq,
				env.OccurredAt,
			)
			if err != nil {
				h.logger.Warn("Failed to update Redis machine state", "machine_id", env.MachineID, "error", err)
			} else if !updated {
				h.logger.Debug("Out-of-order state change rejected by Redis", "machine_id", env.MachineID, "seq", env.Seq)
			}
		}
	}

	return nil
}

func (h *PipelineHandler) HandleDeadLetter(ctx context.Context, raw []byte, reason string) error {
	if h.events != nil {
		return h.events.SaveDeadLetter(ctx, raw, reason)
	}
	return nil
}
