package queries

import (
	"auth-service/internal/usecase"
	"context"
	"time"

	"auth-service/internal/domain"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type ValidateTokenQuery struct {
	AccessToken string
}

type ValidateTokenHandler struct {
	Tokens usecase.TokenService
	Users  usecase.UserReadRepository
}

type ValidateTokenResult struct {
	Valid     bool
	User      *domain.User
	ExpiresAt time.Time
}

func (h *ValidateTokenHandler) Handle(ctx context.Context, q ValidateTokenQuery) (*ValidateTokenResult, error) {
	ctx, span := otel.Tracer("usecase").Start(ctx, "ValidateTokenHandler.Handle", trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()

	if q.AccessToken == "" {
		return &ValidateTokenResult{Valid: false}, nil
	}

	claims, err := h.Tokens.ValidateAccessToken(q.AccessToken)
	if err != nil {
		return &ValidateTokenResult{Valid: false}, nil
	}

	user, err := h.Users.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return &ValidateTokenResult{Valid: false}, nil
	}

	return &ValidateTokenResult{
		Valid:     true,
		User:      user,
		ExpiresAt: claims.ExpiresAt,
	}, nil
}

