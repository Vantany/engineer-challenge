package domain

import "time"

type SessionID string

type Session struct {
	ID               SessionID `db:"id"`
	UserID           UserID    `db:"user_id"`
	RefreshTokenHash string    `db:"refresh_token_hash"`
	ExpiresAt        time.Time `db:"expires_at"`
	Revoked          bool      `db:"revoked"`
	CreatedAt        time.Time `db:"created_at"`
	UpdatedAt        time.Time `db:"updated_at"`
}

func (s *Session) IsExpired(now time.Time) bool {
	return now.After(s.ExpiresAt)
}

func (s *Session) Revoke(now time.Time) {
	s.Revoked = true
	s.UpdatedAt = now
}

