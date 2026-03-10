package commands

import (
	"auth-service/internal/usecase"
	"context"
	"strings"
	"time"

	"auth-service/internal/domain"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type LoginCommand struct {
	Email    string
	Password string
}

type LoginHandler struct {
	Users         usecase.UserRepository
	Sessions      usecase.SessionRepository
	TxManager     usecase.TransactionManager
	PasswordPolicy domain.PasswordPolicy
	TokenService  usecase.TokenService
	Now           func() time.Time
}

type LoginResult struct {
	User               *domain.User
	Session            *domain.Session
	AccessToken        string
	RefreshToken       string
	AccessTokenExpiry  time.Time
	RefreshTokenExpiry time.Time
}

func (h *LoginHandler) Handle(ctx context.Context, cmd LoginCommand) (*LoginResult, error) {
	ctx, span := otel.Tracer("usecase").Start(ctx, "LoginHandler.Handle", trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()

	now := h.safeNow()

	email := domain.Email(strings.ToLower(strings.TrimSpace(cmd.Email)))
	if email == "" || !strings.Contains(string(email), "@") {
		return nil, domain.ErrInvalidEmailFormat
	}

	user, err := h.Users.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, domain.ErrInvalidCredentials
	}

	if !user.VerifyPassword(cmd.Password, h.PasswordPolicy) {
		return nil, domain.ErrInvalidCredentials
	}

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
		if err := h.Sessions.Create(txCtx, session); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &LoginResult{
		User:               user,
		Session:            session,
		AccessToken:        accessToken,
		RefreshToken:       refreshToken,
		AccessTokenExpiry:  accessExp,
		RefreshTokenExpiry: refreshExp,
	}, nil
}

func (h *LoginHandler) safeNow() time.Time {
	if h.Now != nil {
		return h.Now()
	}
	return time.Now().UTC()
}

