package postgres

import (
	"context"
	"database/sql"
	"time"

	"auth-service/internal/domain"

	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type ResetTokenRepository struct {
	db *sqlx.DB
}

func NewResetTokenRepository(db *sqlx.DB) *ResetTokenRepository {
	return &ResetTokenRepository{db: db}
}

func (r *ResetTokenRepository) Create(ctx context.Context, token *domain.ResetToken) error {
	ctx, span := otel.Tracer("postgres").Start(ctx, "ResetTokenRepository.Create", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	const query = `
		INSERT INTO reset_tokens (token_hash, user_id, expires_at, used_at, created_at)
		VALUES ($1, $2, $3, NULL, $4)
	`

	_, err := execContext(ctx, r.db, query,
		token.Token,
		token.UserID,
		token.ExpiresAt,
		time.Now().UTC(),
	)
	return err
}

func (r *ResetTokenRepository) CountRecentByEmail(ctx context.Context, email domain.Email, since time.Time) (int, error) {
	ctx, span := otel.Tracer("postgres").Start(ctx, "ResetTokenRepository.CountRecentByEmail", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	const query = `
		SELECT COUNT(rt.token_hash)
		FROM reset_tokens rt
		JOIN users u ON u.id = rt.user_id
		WHERE u.email = $1 AND rt.created_at >= $2
	`

	var count int
	if err := queryRowxContext(ctx, r.db, query, email, since).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *ResetTokenRepository) GetByTokenHash(ctx context.Context, hash string) (*domain.ResetToken, error) {
	ctx, span := otel.Tracer("postgres").Start(ctx, "ResetTokenRepository.GetByTokenHash", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	const query = `
		SELECT token_hash, user_id, expires_at, used_at, created_at
		FROM reset_tokens
		WHERE token_hash = $1
	`

	var (
		t      domain.ResetToken
		usedAt sql.NullTime
	)

	if err := queryRowxContext(ctx, r.db, query, hash).Scan(
		&t.Token,
		&t.UserID,
		&t.ExpiresAt,
		&usedAt,
		&time.Time{}, // created_at нам здесь не нужен
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if usedAt.Valid {
		t.Used = true
	}

	return &t, nil
}

func (r *ResetTokenRepository) MarkUsed(ctx context.Context, hash string, usedAt time.Time) error {
	ctx, span := otel.Tracer("postgres").Start(ctx, "ResetTokenRepository.MarkUsed", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	const query = `
		UPDATE reset_tokens
		SET used_at = $1
		WHERE token_hash = $2
	`

	_, err := execContext(ctx, r.db, query, usedAt, hash)
	return err
}

