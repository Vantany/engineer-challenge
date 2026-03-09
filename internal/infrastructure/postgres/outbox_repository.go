package postgres

import (
	"context"
	"time"

	"auth-service/internal/domain"

	"github.com/jmoiron/sqlx"
)

type OutboxRepository struct {
	db *sqlx.DB
}

func NewOutboxRepository(db *sqlx.DB) *OutboxRepository {
	return &OutboxRepository{db: db}
}

func (r *OutboxRepository) Create(ctx context.Context, event *domain.OutboxEvent) error {
	const query = `
		INSERT INTO outbox_events (id, type, payload, status, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := execContext(ctx, r.db, query,
		event.ID,
		event.Type,
		event.Payload,
		event.Status,
		event.CreatedAt,
	)
	return err
}

func (r *OutboxRepository) GetPending(ctx context.Context, limit int) ([]*domain.OutboxEvent, error) {
	const query = `
		SELECT id, type, payload, status, created_at, processed_at
		FROM outbox_events
		WHERE status = 'pending'
		ORDER BY created_at ASC
		LIMIT $1
	`

	var events []*domain.OutboxEvent

	var rows *sqlx.Rows
	var err error
	if tx, ok := ctx.Value(txKey{}).(*sqlx.Tx); ok {
		rows, err = tx.QueryxContext(ctx, query, limit)
	} else {
		rows, err = r.db.QueryxContext(ctx, query, limit)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var e domain.OutboxEvent
		if err := rows.StructScan(&e); err != nil {
			return nil, err
		}
		events = append(events, &e)
	}
	return events, rows.Err()
}

func (r *OutboxRepository) MarkProcessed(ctx context.Context, id string, processedAt time.Time) error {
	const query = `
		UPDATE outbox_events
		SET status = 'processed', processed_at = $1
		WHERE id = $2
	`
	_, err := execContext(ctx, r.db, query, processedAt, id)
	return err
}
