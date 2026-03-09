package commands

import (
	"auth-service/internal/logger"
	"auth-service/internal/usecase"
	"context"
	"strings"
	"time"

	"auth-service/internal/domain"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type RegisterCommand struct {
	Email    string
	Password string
}

type RegisterHandler struct {
	Users        usecase.UserRepository
	Sessions     usecase.SessionRepository
	TxManager    usecase.TransactionManager
	PasswordPolicy domain.PasswordPolicy
	TokenService usecase.TokenService
	Now          func() time.Time
}

type RegisterResult struct {
	User              *domain.User
	Session           *domain.Session
	AccessToken       string
	RefreshToken      string
	AccessTokenExpiry time.Time
	RefreshTokenExpiry time.Time
}

func (h *RegisterHandler) Handle(ctx context.Context, cmd RegisterCommand) (*RegisterResult, error) {
	ctx, span := otel.Tracer("usecase").Start(ctx, "RegisterHandler.Handle", trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()

	log := logger.FromContext(ctx)

	now := h.safeNow()

	email := domain.Email(strings.ToLower(strings.TrimSpace(cmd.Email)))
	if email == "" || !strings.Contains(string(email), "@") {
		return nil, domain.ErrInvalidEmailFormat
	}

	existing, err := h.Users.GetByEmail(ctx, email)
	if err != nil {
		log.Error("failed to get user by email", "error", err)
		return nil, err
	}
	if existing != nil {
		return nil, domain.ErrEmailAlreadyExists
	}

	userID := domain.UserID(uuid.NewString())
	user, err := domain.NewUser(userID, email, cmd.Password, h.PasswordPolicy, now)
	if err != nil {
		return nil, err
	}

	accessToken, accessExp, err := h.TokenService.GenerateAccessToken(user, now)
	if err != nil {
		log.Error("failed to generate access token", "error", err)
		return nil, err
	}

	refreshToken, refreshExp, err := h.TokenService.GenerateRefreshToken(now)
	if err != nil {
		log.Error("failed to generate refresh token", "error", err)
		return nil, err
	}

	refreshHash, err := h.TokenService.HashRefreshToken(refreshToken)
	if err != nil {
		log.Error("failed to hash refresh token", "error", err)
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
		if err := h.Users.Create(txCtx, user); err != nil {
			return err
		}

		if err := h.Sessions.Create(txCtx, session); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		log.Error("failed to save user and session to db", "error", err)
		return nil, err
	}

	log.Info("user successfully registered", "user_id", user.ID)

	return &RegisterResult{
		User:               user,
		Session:            session,
		AccessToken:        accessToken,
		RefreshToken:       refreshToken,
		AccessTokenExpiry:  accessExp,
		RefreshTokenExpiry: refreshExp,
	}, nil
}

func (h *RegisterHandler) safeNow() time.Time {
	if h.Now != nil {
		return h.Now()
	}
	return time.Now().UTC()
}

