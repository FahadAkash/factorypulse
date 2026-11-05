package incident

import (
	"context"
	"fmt"
	"time"

	"github.com/FahadAkash/factorypulse/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

// OpenIncident creates a new incident only if the machine does not already have an active one.
func (s *Service) OpenIncident(
	ctx context.Context,
	machineID string,
	sourceEventID string,
	occurredAt time.Time,
	reasonCode string,
) (*domain.Incident, error) {
	if s.pool == nil {
		return nil, nil
	}

	query := `
		INSERT INTO incidents (machine_id, status, opened_at, root_cause, source_event_id)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (machine_id) WHERE status IN ('open', 'acknowledged') DO NOTHING
		RETURNING id, machine_id, status, opened_at, acknowledged_at, resolved_at, assignee, root_cause, source_event_id
	`

	var inc domain.Incident
	err := s.pool.QueryRow(ctx, query,
		machineID,
		string(domain.IncidentStatusOpen),
		occurredAt,
		reasonCode,
		sourceEventID,
	).Scan(
		&inc.ID,
		&inc.MachineID,
		&inc.Status,
		&inc.OpenedAt,
		&inc.AcknowledgedAt,
		&inc.ResolvedAt,
		&inc.Assignee,
		&inc.RootCause,
		&inc.SourceEventID,
	)
	if err != nil {
		// If rows affected is 0 because of ON CONFLICT DO NOTHING, pgx returns pgx.ErrNoRows
		return nil, nil // active incident already exists
	}

	return &inc, nil
}

// AcknowledgeIncident transitions an open incident to acknowledged status.
func (s *Service) AcknowledgeIncident(
	ctx context.Context,
	incidentID string,
	assignee string,
	ackAt time.Time,
) error {
	if s.pool == nil {
		return nil
	}

	query := `
		UPDATE incidents
		SET status = $1, assignee = $2, acknowledged_at = $3
		WHERE id = $4 AND status = 'open'
	`
	_, err := s.pool.Exec(ctx, query, string(domain.IncidentStatusAcknowledged), assignee, ackAt, incidentID)
	return err
}

// ResolveIncident manually closes an incident with a root cause reason.
func (s *Service) ResolveIncident(
	ctx context.Context,
	incidentID string,
	rootCause string,
	resolvedAt time.Time,
) error {
	if s.pool == nil {
		return nil
	}

	query := `
		UPDATE incidents
		SET status = $1, root_cause = $2, resolved_at = $3
		WHERE id = $4 AND status IN ('open', 'acknowledged')
	`
	_, err := s.pool.Exec(ctx, query, string(domain.IncidentStatusResolved), rootCause, resolvedAt, incidentID)
	return err
}

// AutoResolveActiveForMachine automatically resolves active incidents when a machine enters RUNNING state.
func (s *Service) AutoResolveActiveForMachine(
	ctx context.Context,
	machineID string,
	resolvedAt time.Time,
) (int64, error) {
	if s.pool == nil {
		return 0, nil
	}

	query := `
		UPDATE incidents
		SET status = $1, resolved_at = $2
		WHERE machine_id = $3 AND status IN ('open', 'acknowledged') AND opened_at <= $2
	`
	tag, err := s.pool.Exec(ctx, query, string(domain.IncidentStatusResolved), resolvedAt, machineID)
	if err != nil {
		return 0, fmt.Errorf("failed to auto-resolve active incident: %w", err)
	}

	return tag.RowsAffected(), nil
}

// ListActiveIncidents retrieves all incidents currently open or acknowledged.
func (s *Service) ListActiveIncidents(ctx context.Context) ([]domain.Incident, error) {
	if s.pool == nil {
		return nil, nil
	}

	query := `
		SELECT id, machine_id, status, opened_at, acknowledged_at, resolved_at, assignee, root_cause, source_event_id
		FROM incidents
		WHERE status IN ('open', 'acknowledged')
		ORDER BY opened_at DESC
	`
	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query active incidents: %w", err)
	}
	defer rows.Close()

	var list []domain.Incident
	for rows.Next() {
		var inc domain.Incident
		if err := rows.Scan(
			&inc.ID,
			&inc.MachineID,
			&inc.Status,
			&inc.OpenedAt,
			&inc.AcknowledgedAt,
			&inc.ResolvedAt,
			&inc.Assignee,
			&inc.RootCause,
			&inc.SourceEventID,
		); err != nil {
			return nil, err
		}
		list = append(list, inc)
	}
	return list, nil
}
