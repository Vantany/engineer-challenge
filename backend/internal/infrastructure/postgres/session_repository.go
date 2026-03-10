package postgres

import (
	"context"
	"database/sql"

	"auth-service/internal/domain"

	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type SessionRepository struct {
	db *sqlx.DB
}

func NewSessionRepository(db *sqlx.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(ctx context.Context, s *domain.Session) error {
	ctx, span := otel.Tracer("postgres").Start(ctx, "SessionRepository.Create", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	const query = `
		INSERT INTO sessions (id, user_id, refresh_token_hash, expires_at, revoked, created_at, updated_at)
		VALUES (:id, :user_id, :refresh_token_hash, :expires_at, :revoked, :created_at, :updated_at)
	`

	_, err := namedExecContext(ctx, r.db, query, map[string]any{
		"id":                 s.ID,
		"user_id":            s.UserID,
		"refresh_token_hash": s.RefreshTokenHash,
		"expires_at":         s.ExpiresAt,
		"revoked":           s.Revoked,
		"created_at":        s.CreatedAt,
		"updated_at":        s.UpdatedAt,
	})
	return err
}

func (r *SessionRepository) GetByID(ctx context.Context, id domain.SessionID) (*domain.Session, error) {
	ctx, span := otel.Tracer("postgres").Start(ctx, "SessionRepository.GetByID", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	const query = `
		SELECT id, user_id, refresh_token_hash, expires_at, revoked, created_at, updated_at
		FROM sessions
		WHERE id = $1
	`

	var s domain.Session
	if err := queryRowxContext(ctx, r.db, query, id).Scan(
		&s.ID,
		&s.UserID,
		&s.RefreshTokenHash,
		&s.ExpiresAt,
		&s.Revoked,
		&s.CreatedAt,
		&s.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &s, nil
}

func (r *SessionRepository) GetByRefreshTokenHash(ctx context.Context, hash string) (*domain.Session, error) {
	ctx, span := otel.Tracer("postgres").Start(ctx, "SessionRepository.GetByRefreshTokenHash", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	const query = `
		SELECT id, user_id, refresh_token_hash, expires_at, revoked, created_at, updated_at
		FROM sessions
		WHERE refresh_token_hash = $1
	`

	var s domain.Session
	if err := queryRowxContext(ctx, r.db, query, hash).Scan(
		&s.ID,
		&s.UserID,
		&s.RefreshTokenHash,
		&s.ExpiresAt,
		&s.Revoked,
		&s.CreatedAt,
		&s.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &s, nil
}

func (r *SessionRepository) Update(ctx context.Context, s *domain.Session) error {
	ctx, span := otel.Tracer("postgres").Start(ctx, "SessionRepository.Update", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	const query = `
		UPDATE sessions
		SET refresh_token_hash = $1,
		    expires_at = $2,
		    revoked = $3,
		    updated_at = $4
		WHERE id = $5
	`

	_, err := execContext(ctx, r.db, query,
		s.RefreshTokenHash,
		s.ExpiresAt,
		s.Revoked,
		s.UpdatedAt,
		s.ID,
	)
	return err
}

