package audit

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Execer interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

type OutboxRepository struct {
	pool *pgxpool.Pool
}

func NewOutboxRepository(pool *pgxpool.Pool) *OutboxRepository {
	return &OutboxRepository{
		pool: pool,
	}
}

func (r *OutboxRepository) Create(ctx context.Context, event Event) error {
	if r == nil || r.pool == nil {
		return fmt.Errorf("audit outbox repository is not configured")
	}

	return Insert(ctx, r.pool, event)
}

func Insert(ctx context.Context, execer Execer, event Event) error {
	if execer == nil {
		return fmt.Errorf("audit outbox execer is nil")
	}

	event = event.WithDefaults()

	payloadJSON, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("marshal audit payload: %w", err)
	}

	const query = `
		INSERT INTO audit_outbox (
			id,
			event_type,
			source_system,
			actor_id,
			subject_id,
			resource_type,
			resource_id,
			action,
			result,
			correlation_id,
			payload_json,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11::jsonb, $12)
	`

	_, err = execer.Exec(
		ctx,
		query,
		event.ID,
		event.EventType,
		event.SourceSystem,
		nullIfEmpty(event.ActorID),
		nullIfEmpty(event.SubjectID),
		event.ResourceType,
		event.ResourceID,
		event.Action,
		event.Result,
		nullIfEmpty(event.CorrelationID),
		string(payloadJSON),
		event.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert audit outbox event: %w", err)
	}

	return nil
}

func nullIfEmpty(value string) any {
	if value == "" {
		return nil
	}

	return value
}
