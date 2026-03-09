package commands

import (
	"auth-service/internal/usecase"
	"context"
	"time"

	"auth-service/internal/domain"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type RefreshTokenCommand struct {
	RefreshToken string
}

type RefreshTokenHandler struct {
	Users         usecase.UserRepository
	Sessions      usecase.SessionRepository
	TokenService  usecase.TokenService
	Now           func() time.Time
}

type RefreshTokenResult struct {
	User               *domain.User
	Session            *domain.Session
	AccessToken        string
	RefreshToken       string
	AccessTokenExpiry  time.Time
	RefreshTokenExpiry time.Time
}

func (h *RefreshTokenHandler) Handle(ctx context.Context, cmd RefreshTokenCommand) (*RefreshTokenResult, error) {
	ctx, span := otel.Tracer("usecase").Start(ctx, "RefreshTokenHandler.Handle", trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()

	now := h.safeNow()

	if cmd.RefreshToken == "" {
		return nil, domain.ErrInvalidCredentials
	}

	refreshHash, err := h.TokenService.HashRefreshToken(cmd.RefreshToken)
	if err != nil {
		return nil, err
	}

	session, err := h.Sessions.GetByRefreshTokenHash(ctx, refreshHash)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, domain.ErrSessionNotFound
	}
	if session.Revoked {
		return nil, domain.ErrSessionRevoked
	}
	if session.IsExpired(now) {
		return nil, domain.ErrSessionNotFound
	}

	user, err := h.Users.GetByID(ctx, session.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, domain.ErrInvalidCredentials
	}

	accessToken, accessExp, err := h.TokenService.GenerateAccessToken(user, now)
	if err != nil {
		return nil, err
	}

	newRefreshToken, refreshExp, err := h.TokenService.GenerateRefreshToken(now)
	if err != nil {
		return nil, err
	}

	newRefreshHash, err := h.TokenService.HashRefreshToken(newRefreshToken)
	if err != nil {
		return nil, err
	}

	session.RefreshTokenHash = newRefreshHash
	session.ExpiresAt = refreshExp
	session.UpdatedAt = now

	if err := h.Sessions.Update(ctx, session); err != nil {
		return nil, err
	}

	return &RefreshTokenResult{
		User:               user,
		Session:            session,
		AccessToken:        accessToken,
		RefreshToken:       newRefreshToken,
		AccessTokenExpiry:  accessExp,
		RefreshTokenExpiry: refreshExp,
	}, nil
}

func (h *RefreshTokenHandler) safeNow() time.Time {
	if h.Now != nil {
		return h.Now()
	}
	return time.Now().UTC()
}

