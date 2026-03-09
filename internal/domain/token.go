package domain

import "time"

type ResetToken struct {
	Token     string
	UserID    UserID
	ExpiresAt time.Time
	Used      bool
}

func (t *ResetToken) IsExpired(now time.Time) bool {
	return now.After(t.ExpiresAt)
}

func (t *ResetToken) MarkUsed() {
	t.Used = true
}

const DefaultResetTokenTTL = 15 * time.Minute


