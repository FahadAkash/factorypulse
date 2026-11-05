package pg

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductionBucket struct {
	MachineID string    `json:"machine_id"`
	Bucket    time.Time `json:"bucket"`
	Good      int       `json:"good"`
	Rejected  int       `json:"rejected"`
	RunMs     int64     `json:"run_ms"`
}

type AggregateStore struct {
	pool *pgxpool.Pool
}

func NewAggregateStore(pool *pgxpool.Pool) *AggregateStore {
	return &AggregateStore{pool: pool}
}

// UpsertProductionMinute updates or increments minute-bucket metrics atomically within an existing tx.
func (s *AggregateStore) UpsertProductionMinuteTx(
	ctx context.Context,
	tx pgx.Tx,
	machineID string,
	occurredAt time.Time,
	goodDelta int,
	rejectedDelta int,
	cycleMs int64,
) error {
	query := `
		INSERT INTO production_minute (machine_id, bucket, good, rejected, run_ms)
		VALUES ($1, date_trunc('minute', $2::timestamptz), $3, $4, $5)
		ON CONFLICT (machine_id, bucket) DO UPDATE
		SET good = production_minute.good + EXCLUDED.good,
		    rejected = production_minute.rejected + EXCLUDED.rejected,
		    run_ms = production_minute.run_ms + EXCLUDED.run_ms;
	`
	_, err := tx.Exec(ctx, query, machineID, occurredAt, goodDelta, rejectedDelta, cycleMs)
	if err != nil {
		return fmt.Errorf("failed to upsert production_minute: %w", err)
	}
	return nil
}

// RecordDefectTx increments the defect counter for a given defect code in the minute bucket.
func (s *AggregateStore) RecordDefectTx(
	ctx context.Context,
	tx pgx.Tx,
	machineID string,
	occurredAt time.Time,
	defectCode string,
) error {
	if defectCode == "" {
		return nil
	}

	query := `
		INSERT INTO defects (machine_id, bucket, defect_code, count)
		VALUES ($1, date_trunc('minute', $2::timestamptz), $3, 1)
		ON CONFLICT (machine_id, bucket, defect_code) DO UPDATE
		SET count = defects.count + 1;
	`
	_, err := tx.Exec(ctx, query, machineID, occurredAt, defectCode)
	if err != nil {
		return fmt.Errorf("failed to record defect: %w", err)
	}
	return nil
}

// GetProductionBuckets retrieves minute buckets for a machine within a given time range.
func (s *AggregateStore) GetProductionBuckets(
	ctx context.Context,
	machineID string,
	from time.Time,
	to time.Time,
) ([]ProductionBucket, error) {
	if s.pool == nil {
		return nil, nil
	}

	query := `
		SELECT machine_id, bucket, good, rejected, run_ms
		FROM production_minute
		WHERE machine_id = $1 AND bucket >= $2 AND bucket <= $3
		ORDER BY bucket ASC
	`
	rows, err := s.pool.Query(ctx, query, machineID, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to query production buckets: %w", err)
	}
	defer rows.Close()

	var buckets []ProductionBucket
	for rows.Next() {
		var b ProductionBucket
		if err := rows.Scan(&b.MachineID, &b.Bucket, &b.Good, &b.Rejected, &b.RunMs); err != nil {
			return nil, err
		}
		buckets = append(buckets, b)
	}
	return buckets, nil
}
