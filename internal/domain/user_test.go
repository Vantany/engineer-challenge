package domain_test

import (
	"testing"
	"time"

	"auth-service/internal/domain"
)

func TestUser_AuthMethod_VerifyPassword(t *testing.T) {
	policy := domain.NewDefaultPasswordPolicy()
	now := time.Now()

	user, err := domain.NewUser("user-1", "test@example.com", "Str0ng!Pass1", policy, now)
	if err != nil {
		t.Fatalf("NewUser() failed: %v", err)
	}

	if !user.VerifyPassword("Str0ng!Pass1", policy) {
		t.Errorf("VerifyPassword() failed with correct password")
	}

	if user.VerifyPassword("WrongPass1!", policy) {
		t.Errorf("VerifyPassword() succeeded with incorrect password")
	}
}

func TestUser_UpdatePassword(t *testing.T) {
	policy := domain.NewDefaultPasswordPolicy()
	now := time.Now()

	user, err := domain.NewUser("user-1", "test@example.com", "Str0ng!Pass1", policy, now)
	if err != nil {
		t.Fatalf("NewUser() failed: %v", err)
	}

	err = user.UpdatePassword("NewStr0ng!Pass2", policy, now.Add(time.Hour))
	if err != nil {
		t.Errorf("UpdatePassword() failed: %v", err)
	}

	if !user.VerifyPassword("NewStr0ng!Pass2", policy) {
		t.Errorf("VerifyPassword() failed with new password")
	}

	if user.VerifyPassword("Str0ng!Pass1", policy) {
		t.Errorf("VerifyPassword() succeeded with old password")
	}
}
