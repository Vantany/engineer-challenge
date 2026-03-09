package worker

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"auth-service/internal/domain"
	"auth-service/internal/usecase"
)

type OutboxWorker struct {
	repo         usecase.OutboxRepository
	emailService usecase.EmailService
	logger       *slog.Logger
	pollInterval time.Duration
}

func NewOutboxWorker(
	repo usecase.OutboxRepository,
	emailService usecase.EmailService,
	logger *slog.Logger,
	pollInterval time.Duration,
) *OutboxWorker {
	if pollInterval == 0 {
		pollInterval = 5 * time.Second
	}
	return &OutboxWorker{
		repo:         repo,
		emailService: emailService,
		logger:       logger,
		pollInterval: pollInterval,
	}
}

func (w *OutboxWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	w.logger.Info("Outbox worker started")

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("Outbox worker stopped")
			return
		case <-ticker.C:
			w.processPendingEvents(ctx)
		}
	}
}

func (w *OutboxWorker) processPendingEvents(ctx context.Context) {
	// В целях упрощения обрабатываем до 50 событий за один раз
	events, err := w.repo.GetPending(ctx, 50)
	if err != nil {
		w.logger.Error("Failed to get pending outbox events", slog.String("error", err.Error()))
		return
	}

	for _, event := range events {
		if err := w.processEvent(ctx, event); err != nil {
			w.logger.Error("Failed to process outbox event",
				slog.String("event_id", event.ID.String()),
				slog.String("error", err.Error()),
			)
			continue
		}
	}
}

func (w *OutboxWorker) processEvent(ctx context.Context, event *domain.OutboxEvent) error {
	switch event.Type {
	case domain.OutboxEventTypeResetToken:
		var payload domain.ResetTokenPayload
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			return err
		}

		// Логируем email
		if err := w.emailService.SendResetToken(payload.Email, payload.Token); err != nil {
			return err
		}

	default:
		w.logger.Warn("Unknown outbox event type", slog.String("type", string(event.Type)))
	}

	// Помечаем событие как обработанное
	now := time.Now().UTC()
	if err := w.repo.MarkProcessed(ctx, event.ID.String(), now); err != nil {
		return err
	}

	w.logger.Info("Successfully processed outbox event", slog.String("event_id", event.ID.String()))
	return nil
}
