package commands

import (
	"context"
	"time"

	"auth-service/internal/domain"
	"auth-service/internal/usecase"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type RequestPasswordResetCommand struct {
	Email string
}

type RequestPasswordResetHandler struct {
	Users        usecase.UserRepository
	ResetTokens  usecase.ResetTokenRepository
	Outbox       usecase.OutboxRepository
	TxManager    usecase.TransactionManager
	TokenService usecase.TokenService
	Now          func() time.Time
}

func (h *RequestPasswordResetHandler) Handle(ctx context.Context, cmd RequestPasswordResetCommand) error {
	ctx, span := otel.Tracer("usecase").Start(ctx, "RequestPasswordResetHandler.Handle", trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()

	now := h.safeNow()

	email := domain.Email(cmd.Email)
	user, err := h.Users.GetByEmail(ctx, email)
	if err != nil {
		return err
	}
	if user == nil {
		// Не раскрываем существует ли пользователь
		return nil
	}

	since := now.Add(-1 * time.Hour)
	count, err := h.ResetTokens.CountRecentByEmail(ctx, email, since)
	if err != nil {
		return err
	}
	if count >= 3 {
		return domain.ErrTooManyResetRequests
	}

	rawToken := uuid.NewString()
	tokenHash, err := h.TokenService.HashRefreshToken(rawToken)
	if err != nil {
		return err
	}
	resetToken := user.CreateResetToken(tokenHash, now)

	outboxEvent, err := domain.NewResetTokenOutboxEvent(string(user.Email), rawToken, now)
	if err != nil {
		return err
	}

	// Сохраняем токен и событие в рамках одной транзакции
	err = h.TxManager.RunInTx(ctx, func(txCtx context.Context) error {
		if err := h.ResetTokens.Create(txCtx, resetToken); err != nil {
			return err
		}
		if err := h.Outbox.Create(txCtx, outboxEvent); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (h *RequestPasswordResetHandler) safeNow() time.Time {
	if h.Now != nil {
		return h.Now()
	}
	return time.Now().UTC()
}
