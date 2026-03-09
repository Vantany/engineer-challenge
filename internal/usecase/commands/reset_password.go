package commands

import (
	"auth-service/internal/usecase"
	"context"
	"time"

	"auth-service/internal/domain"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type ResetPasswordCommand struct {
	Token       string
	NewPassword string
}

type ResetPasswordHandler struct {
	Users          usecase.UserRepository
	ResetTokens    usecase.ResetTokenRepository
	Sessions       usecase.SessionRepository
	TxManager      usecase.TransactionManager
	PasswordPolicy domain.PasswordPolicy
	TokenService   usecase.TokenService
	Now            func() time.Time
}

type ResetPasswordResult struct {
	User               *domain.User
	Session            *domain.Session
	AccessToken        string
	RefreshToken       string
	AccessTokenExpiry  time.Time
	RefreshTokenExpiry time.Time
}

func (h *ResetPasswordHandler) Handle(ctx context.Context, cmd ResetPasswordCommand) (*ResetPasswordResult, error) {
	ctx, span := otel.Tracer("usecase").Start(ctx, "ResetPasswordHandler.Handle", trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()

	now := h.safeNow()

	if cmd.Token == "" {
		return nil, domain.ErrResetTokenExpired
	}

	// Для хранения reset-токена в БД мы используем хэш поэтому здесь также хэшируем входной токен
	tokenHash, err := h.TokenService.HashRefreshToken(cmd.Token)
	if err != nil {
		return nil, err
	}

	resetToken, err := h.ResetTokens.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, err
	}
	if resetToken == nil {
		return nil, domain.ErrResetTokenExpired
	}
	if resetToken.Used {
		return nil, domain.ErrResetTokenUsed
	}
	if resetToken.IsExpired(now) {
		return nil, domain.ErrResetTokenExpired
	}

	user, err := h.Users.GetByID(ctx, resetToken.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, domain.ErrInvalidCredentials
	}

	if err := user.UpdatePassword(cmd.NewPassword, h.PasswordPolicy, now); err != nil {
		return nil, err
	}

	// После смены пароля открываем новую сессию.
	accessToken, accessExp, err := h.TokenService.GenerateAccessToken(user, now)
	if err != nil {
		return nil, err
	}

	refreshToken, refreshExp, err := h.TokenService.GenerateRefreshToken(now)
	if err != nil {
		return nil, err
	}

	refreshHash, err := h.TokenService.HashRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	session := &domain.Session{
		ID:               domain.SessionID(uuid.NewString()),
		UserID:           user.ID,
		RefreshTokenHash: refreshHash,
		ExpiresAt:        refreshExp,
		Revoked:          false,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	err = h.TxManager.RunInTx(ctx, func(txCtx context.Context) error {
		if err := h.Users.Update(txCtx, user); err != nil {
			return err
		}

		if err := h.ResetTokens.MarkUsed(txCtx, tokenHash, now); err != nil {
			return err
		}

		if err := h.Sessions.Create(txCtx, session); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &ResetPasswordResult{
		User:               user,
		Session:            session,
		AccessToken:        accessToken,
		RefreshToken:       refreshToken,
		AccessTokenExpiry:  accessExp,
		RefreshTokenExpiry: refreshExp,
	}, nil
}

func (h *ResetPasswordHandler) safeNow() time.Time {
	if h.Now != nil {
		return h.Now()
	}
	return time.Now().UTC()
}
