package commands

import (
	"auth-service/internal/usecase"
	"context"
	"time"

	"auth-service/internal/domain"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type LogoutCommand struct {
	RefreshToken string
}

type LogoutHandler struct {
	Sessions     usecase.SessionRepository
	TokenService usecase.TokenService
	Now          func() time.Time
}

func (h *LogoutHandler) Handle(ctx context.Context, cmd LogoutCommand) error {
	ctx, span := otel.Tracer("usecase").Start(ctx, "LogoutHandler.Handle", trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()

	now := h.safeNow()

	if cmd.RefreshToken == "" {
		return domain.ErrInvalidCredentials
	}

	hash, err := h.TokenService.HashRefreshToken(cmd.RefreshToken)
	if err != nil {
		return err
	}

	session, err := h.Sessions.GetByRefreshTokenHash(ctx, hash)
	if err != nil {
		return err
	}
	if session == nil {
		return domain.ErrSessionNotFound
	}

	if session.Revoked {
		return domain.ErrSessionRevoked
	}

	session.Revoke(now)
	if err := h.Sessions.Update(ctx, session); err != nil {
		return err
	}

	return nil
}

func (h *LogoutHandler) safeNow() time.Time {
	if h.Now != nil {
		return h.Now()
	}
	return time.Now().UTC()
}
