package integration

import (
	"context"
	"testing"
	"time"

	"auth-service/internal/domain"
	"auth-service/internal/infrastructure/postgres"
	"auth-service/internal/usecase"
	"auth-service/internal/usecase/commands"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockTokenService is a simple mock to avoid setting up JWT keys
type mockTokenService struct{}

func (m *mockTokenService) GenerateAccessToken(user *domain.User, now time.Time) (string, time.Time, error) {
	return "access-token", now.Add(time.Hour), nil
}

func (m *mockTokenService) GenerateRefreshToken(now time.Time) (string, time.Time, error) {
	return "refresh-token", now.Add(time.Hour * 24), nil
}

func (m *mockTokenService) HashRefreshToken(token string) (string, error) {
	return "hashed-" + token, nil
}

func (m *mockTokenService) ValidateAccessToken(token string) (*usecase.TokenClaims, error) {
	return nil, nil
}

func TestCommand_RegisterUser_TxRollback(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	usersRepo := postgres.NewUserRepository(db)
	sessionsRepo := postgres.NewSessionRepository(db)
	txManager := postgres.NewTxManager(db)

	handler := &commands.RegisterHandler{
		Users:          usersRepo,
		Sessions:       sessionsRepo, // wait, we will intentionally cause an error here
		TxManager:      txManager,
		PasswordPolicy: domain.NewDefaultPasswordPolicy(),
		TokenService:   &mockTokenService{},
		Now:            func() time.Time { return time.Now().UTC() },
	}

	// We'll sabotage the sessions table to force an insert error.
	_, err := db.Exec("ALTER TABLE sessions RENAME COLUMN refresh_token_hash TO bad_column")
	require.NoError(t, err)

	ctx := context.Background()

	cmd := commands.RegisterCommand{
		Email:    "rollback@example.com",
		Password: "Valid!Password1",
	}

	// This should fail during the session creation part, which is AFTER user creation.
	_, err = handler.Handle(ctx, cmd)
	require.Error(t, err, "expected error due to sabotaged sessions table")

	// Verify Rollback: the user should NOT exist in the database.
	user, err := usersRepo.GetByEmail(ctx, "rollback@example.com")
	require.NoError(t, err)
	assert.Nil(t, user, "user should not exist because the transaction rolled back")
}
