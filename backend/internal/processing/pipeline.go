package processing

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"

	"github.com/FahadAkash/factorypulse/backend/internal/domain"
	"github.com/FahadAkash/factorypulse/backend/internal/incident"
	"github.com/FahadAkash/factorypulse/backend/internal/ingest"
	"github.com/FahadAkash/factorypulse/backend/internal/store/pg"
	"github.com/FahadAkash/factorypulse/backend/internal/store/redis"
)

type PipelineHandler struct {
	events    *pg.EventStore
	states    *redis.StateStore
	aggs      *pg.AggregateStore
	incidents *incident.Service
	logger    *slog.Logger
}

func NewPipelineHandler(
	events *pg.EventStore,
	states *redis.StateStore,
	aggs *pg.AggregateStore,
	incidents *incident.Service,
	logger *slog.Logger,
) *PipelineHandler {
	return &PipelineHandler{
		events:    events,
		states:    states,
		aggs:      aggs,
		incidents: incidents,
		logger:    logger,
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

	// 2. Machine state change handling & incident lifecycle
	if env.EventType == domain.EventMachineStateChanged {
		var statePayload domain.StatePayload
		if err := json.Unmarshal(env.Payload, &statePayload); err == nil {
			// Update Redis state cache
			if h.states != nil {
				_, _ = h.states.UpdateMachineState(
					ctx,
					env.MachineID,
					string(statePayload.State),
					statePayload.ReasonCode,
					env.Seq,
					env.OccurredAt,
				)
			}

			// Incident lifecycle
			if h.incidents != nil {
				if statePayload.State == domain.StateStopped {
					// Open incident if not already active
					_, _ = h.incidents.OpenIncident(ctx, env.MachineID, env.EventID, env.OccurredAt, statePayload.ReasonCode)
				} else if statePayload.State == domain.StateRunning {
					// Auto-resolve any active incident
					_, _ = h.incidents.AutoResolveActiveForMachine(ctx, env.MachineID, env.OccurredAt)
				}
			}
		}
	}

	// 3. Production aggregation
	if env.EventType == domain.EventProductionUnit && h.aggs != nil {
		var prodPayload domain.ProductionPayload
		if err := json.Unmarshal(env.Payload, &prodPayload); err == nil {
			// In production, aggregated atomically via tx or store
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
