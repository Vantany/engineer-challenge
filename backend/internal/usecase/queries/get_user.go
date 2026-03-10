package queries

import (
	"auth-service/internal/usecase"
	"context"

	"auth-service/internal/domain"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type GetUserQuery struct {
	ID domain.UserID
}

type GetUserHandler struct {
	Users usecase.UserReadRepository
}

func (h *GetUserHandler) Handle(ctx context.Context, q GetUserQuery) (*domain.User, error) {
	ctx, span := otel.Tracer("usecase").Start(ctx, "GetUserHandler.Handle", trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()

	if q.ID == "" {
		return nil, nil
	}
	return h.Users.GetByID(ctx, q.ID)
}

