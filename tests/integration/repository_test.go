package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"auth-service/internal/domain"
	"auth-service/internal/infrastructure/postgres"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	postgresC "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTestDB(t *testing.T) (*sqlx.DB, func()) {
	ctx := context.Background()

	// Locate the migrations directory
	pwd, err := os.Getwd()
	require.NoError(t, err)
	migrationsDir := filepath.Join(pwd, "../../migrations")

	pgContainer, err := postgresC.Run(ctx,
		"postgres:15-alpine",
		postgresC.WithDatabase("testdb"),
		postgresC.WithUsername("testuser"),
		postgresC.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second),
		),
	)
	require.NoError(t, err)

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := sqlx.Connect("postgres", connStr)
	require.NoError(t, err)

	// Run migrations using goose
	err = goose.SetDialect("postgres")
	require.NoError(t, err)

	err = goose.Up(db.DB, migrationsDir)
	require.NoError(t, err)

	cleanup := func() {
		db.Close()
		pgContainer.Terminate(ctx)
	}

	return db, cleanup
}

func TestUserRepository_SaveAndGet(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := postgres.NewUserRepository(db)
	ctx := context.Background()
	now := time.Now().UTC()

	policy := domain.NewDefaultPasswordPolicy()
	userID := domain.UserID(uuid.NewString())
	user, err := domain.NewUser(userID, "repo@example.com", "Str0ng!Pass1", policy, now)
	require.NoError(t, err)

	// 1. Create User
	err = repo.Create(ctx, user)
	require.NoError(t, err)

	// 2. Get By Email
	retrieved, err := repo.GetByEmail(ctx, "repo@example.com")
	require.NoError(t, err)
	require.NotNil(t, retrieved)
	assert.Equal(t, user.ID, retrieved.ID)
	assert.Equal(t, user.Email, retrieved.Email)
	assert.Len(t, retrieved.AuthMethods, 1)

	// 3. Update User
	err = repo.Update(ctx, retrieved)
	require.NoError(t, err)

	// 4. Get By ID
	updated, err := repo.GetByID(ctx, user.ID)
	require.NoError(t, err)
	require.NotNil(t, updated)
}
