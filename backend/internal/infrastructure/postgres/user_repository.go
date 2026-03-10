package postgres

import (
	"context"
	"database/sql"

	"auth-service/internal/domain"

	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	ctx, span := otel.Tracer("postgres").Start(ctx, "UserRepository.Create", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	var tx *sqlx.Tx
	var err error
	commitNeeded := false

	if existingTx, ok := ctx.Value(txKey{}).(*sqlx.Tx); ok {
		tx = existingTx
	} else {
		tx, err = r.db.BeginTxx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()
		commitNeeded = true
	}

	const qUser = `
		INSERT INTO users (id, email, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
	`
	_, err = tx.ExecContext(ctx, qUser, user.ID, user.Email, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return err
	}

	const qAuth = `
		INSERT INTO auth_methods (id, user_id, provider, credential, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	for _, am := range user.AuthMethods {
		_, err = tx.ExecContext(ctx, qAuth, am.ID, am.UserID, am.Provider, am.Credential, am.CreatedAt, am.UpdatedAt)
		if err != nil {
			return err
		}
	}

	if commitNeeded {
		return tx.Commit()
	}
	return nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email domain.Email) (*domain.User, error) {
	ctx, span := otel.Tracer("postgres").Start(ctx, "UserRepository.GetByEmail", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	const qUser = `
		SELECT id, email, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	var u domain.User
	err := queryRowxContext(ctx, r.db, qUser, email).Scan(&u.ID, &u.Email, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	u.AuthMethods, err = r.getAuthMethods(ctx, u.ID)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id domain.UserID) (*domain.User, error) {
	const qUser = `
		SELECT id, email, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	var u domain.User
	err := queryRowxContext(ctx, r.db, qUser, id).Scan(&u.ID, &u.Email, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	u.AuthMethods, err = r.getAuthMethods(ctx, u.ID)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	ctx, span := otel.Tracer("postgres").Start(ctx, "UserRepository.Update", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	var tx *sqlx.Tx
	var err error
	commitNeeded := false

	if existingTx, ok := ctx.Value(txKey{}).(*sqlx.Tx); ok {
		tx = existingTx
	} else {
		tx, err = r.db.BeginTxx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()
		commitNeeded = true
	}

	const qUser = `
		UPDATE users
		SET email = $1, updated_at = $2
		WHERE id = $3
	`
	_, err = tx.ExecContext(ctx, qUser, user.Email, user.UpdatedAt, user.ID)
	if err != nil {
		return err
	}

	const qDel = `DELETE FROM auth_methods WHERE user_id = $1`
	_, err = tx.ExecContext(ctx, qDel, user.ID)
	if err != nil {
		return err
	}

	const qAuth = `
		INSERT INTO auth_methods (id, user_id, provider, credential, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	for _, am := range user.AuthMethods {
		_, err = tx.ExecContext(ctx, qAuth, am.ID, am.UserID, am.Provider, am.Credential, am.CreatedAt, am.UpdatedAt)
		if err != nil {
			return err
		}
	}

	if commitNeeded {
		return tx.Commit()
	}
	return nil
}

func (r *UserRepository) getAuthMethods(ctx context.Context, userID domain.UserID) ([]domain.AuthMethod, error) {
	const qAuth = `
		SELECT id, user_id, provider, credential, created_at, updated_at
		FROM auth_methods
		WHERE user_id = $1
	`
	rows, err := queryContext(ctx, r.db, qAuth, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ams []domain.AuthMethod
	for rows.Next() {
		var am domain.AuthMethod
		if err := rows.Scan(&am.ID, &am.UserID, &am.Provider, &am.Credential, &am.CreatedAt, &am.UpdatedAt); err != nil {
			return nil, err
		}
		ams = append(ams, am)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ams, nil
}
