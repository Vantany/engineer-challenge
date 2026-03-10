package queries

import (
	"auth-service/internal/usecase"
	"context"

	"auth-service/internal/domain"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type GetSessionQuery struct {
	ID domain.SessionID
}

type GetSessionHandler struct {
	Sessions usecase.SessionReadRepository
}

func (h *GetSessionHandler) Handle(ctx context.Context, q GetSessionQuery) (*domain.Session, error) {
	ctx, span := otel.Tracer("usecase").Start(ctx, "GetSessionHandler.Handle", trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()

	if q.ID == "" {
		return nil, nil
	}
	return h.Sessions.GetByID(ctx, q.ID)
}

