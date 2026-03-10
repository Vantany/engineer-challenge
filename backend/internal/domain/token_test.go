package domain_test

import (
	"testing"
	"time"

	"auth-service/internal/domain"
)

func TestResetToken_ExpirationAndUse(t *testing.T) {
	now := time.Now()
	token := &domain.ResetToken{
		Token:     "some-token-hash",
		UserID:    "user-1",
		ExpiresAt: now.Add(15 * time.Minute),
		Used:      false,
	}

	if token.IsExpired(now) {
		t.Errorf("IsExpired() returned true for a valid token")
	}

	if !token.IsExpired(now.Add(20 * time.Minute)) {
		t.Errorf("IsExpired() returned false for an expired token")
	}

	if token.Used {
		t.Errorf("token.Used should be false initially")
	}

	token.MarkUsed()

	if !token.Used {
		t.Errorf("token.Used should be true after MarkUsed()")
	}
}
