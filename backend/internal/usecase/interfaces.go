package usecase

import (
	"context"
	"time"

	"auth-service/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	Update(ctx context.Context, user *domain.User) error
	GetByEmail(ctx context.Context, email domain.Email) (*domain.User, error)
	GetByID(ctx context.Context, id domain.UserID) (*domain.User, error)
}

type UserReadRepository interface {
	GetByID(ctx context.Context, id domain.UserID) (*domain.User, error)
}

type SessionRepository interface {
	Create(ctx context.Context, session *domain.Session) error
	GetByRefreshTokenHash(ctx context.Context, hash string) (*domain.Session, error)
	Update(ctx context.Context, session *domain.Session) error
}

type SessionReadRepository interface {
	GetByID(ctx context.Context, id domain.SessionID) (*domain.Session, error)
}

type ResetTokenRepository interface {
	Create(ctx context.Context, token *domain.ResetToken) error
	CountRecentByEmail(ctx context.Context, email domain.Email, since time.Time) (int, error)
	GetByTokenHash(ctx context.Context, hash string) (*domain.ResetToken, error)
	MarkUsed(ctx context.Context, hash string, usedAt time.Time) error
}

type TokenClaims struct {
	UserID    domain.UserID
	ExpiresAt time.Time
}

type TokenService interface {
	GenerateAccessToken(user *domain.User, now time.Time) (string, time.Time, error)
	GenerateRefreshToken(now time.Time) (string, time.Time, error)
	HashRefreshToken(token string) (string, error)
	ValidateAccessToken(token string) (*TokenClaims, error)
}

type EmailService interface {
	SendResetToken(email, token string) error
}

type OutboxRepository interface {
	Create(ctx context.Context, event *domain.OutboxEvent) error
	GetPending(ctx context.Context, limit int) ([]*domain.OutboxEvent, error)
	MarkProcessed(ctx context.Context, id string, processedAt time.Time) error
}

type TransactionManager interface {
	RunInTx(ctx context.Context, fn func(ctx context.Context) error) error
}
