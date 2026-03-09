package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type OutboxEventStatus string

const (
	OutboxEventStatusPending   OutboxEventStatus = "pending"
	OutboxEventStatusProcessed OutboxEventStatus = "processed"
)

const (
	OutboxEventTypeResetToken OutboxEventType = "ResetToken"
)

type OutboxEventType string

type OutboxEvent struct {
	ID          uuid.UUID         `db:"id"`
	Type        OutboxEventType   `db:"type"`
	Payload     json.RawMessage   `db:"payload"`
	Status      OutboxEventStatus `db:"status"`
	CreatedAt   time.Time         `db:"created_at"`
	ProcessedAt *time.Time        `db:"processed_at"`
}

type ResetTokenPayload struct {
	Email string `json:"email"`
	Token string `json:"token"`
}

func NewResetTokenOutboxEvent(email, token string, now time.Time) (*OutboxEvent, error) {
	payload := ResetTokenPayload{
		Email: email,
		Token: token,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return &OutboxEvent{
		ID:        uuid.New(),
		Type:      OutboxEventTypeResetToken,
		Payload:   payloadBytes,
		Status:    OutboxEventStatusPending,
		CreatedAt: now,
	}, nil
}
