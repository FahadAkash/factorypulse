package pg

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/FahadAkash/factorypulse/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrDuplicateEvent = errors.New("duplicate event ignored")

type DBTx interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Begin(ctx context.Context) (pgx.Tx, error)
}

type EventStore struct {
	pool *pgxpool.Pool
}

func NewEventStore(pool *pgxpool.Pool) *EventStore {
	return &EventStore{pool: pool}
}

func (s *EventStore) SaveEvent(ctx context.Context, env *domain.EventEnvelope) (bool, error) {
	if s.pool == nil {
		return false, errors.New("pg pool is nil")
	}

	query := `
		INSERT INTO events (
			event_id, factory_id, line_id, machine_id, run_id,
			event_type, seq, occurred_at, trace_id, payload
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (event_id) DO NOTHING
	`

	tag, err := s.pool.Exec(ctx, query,
		env.EventID,
		env.FactoryID,
		env.LineID,
		env.MachineID,
		env.RunID,
		string(env.EventType),
		env.Seq,
		env.OccurredAt,
		env.TraceID,
		env.Payload,
	)
	if err != nil {
		return false, fmt.Errorf("failed to insert event: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return false, ErrDuplicateEvent
	}

	return true, nil
}

func (s *EventStore) SaveDeadLetter(ctx context.Context, raw []byte, reason string) error {
	if s.pool == nil {
		return errors.New("pg pool is nil")
	}

	query := `
		INSERT INTO dead_letter_events (raw, reason, received_at)
		VALUES ($1, $2, $3)
	`
	_, err := s.pool.Exec(ctx, query, raw, reason, time.Now().UTC())
	return err
}
