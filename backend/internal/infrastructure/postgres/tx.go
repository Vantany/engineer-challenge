package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type txKey struct{}

type TxManager struct {
	db *sqlx.DB
}

func NewTxManager(db *sqlx.DB) *TxManager {
	return &TxManager{db: db}
}

func (tm *TxManager) RunInTx(ctx context.Context, fn func(ctx context.Context) error) error {
	ctx, span := otel.Tracer("postgres").Start(ctx, "TxManager.RunInTx", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	tx, err := tm.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	txCtx := context.WithValue(ctx, txKey{}, tx)

	if err := fn(txCtx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil && !errors.Is(rbErr, sql.ErrTxDone) {
		}
		return err
	}

	return tx.Commit()
}

func execContext(ctx context.Context, db *sqlx.DB, query string, args ...any) (sql.Result, error) {
	if tx, ok := ctx.Value(txKey{}).(*sqlx.Tx); ok {
		return tx.ExecContext(ctx, query, args...)
	}
	return db.ExecContext(ctx, query, args...)
}

func queryRowxContext(ctx context.Context, db *sqlx.DB, query string, args ...any) *sqlx.Row {
	if tx, ok := ctx.Value(txKey{}).(*sqlx.Tx); ok {
		return tx.QueryRowxContext(ctx, query, args...)
	}
	return db.QueryRowxContext(ctx, query, args...)
}

func queryContext(ctx context.Context, db *sqlx.DB, query string, args ...any) (*sql.Rows, error) {
	if tx, ok := ctx.Value(txKey{}).(*sqlx.Tx); ok {
		return tx.QueryContext(ctx, query, args...)
	}
	return db.QueryContext(ctx, query, args...)
}

func namedExecContext(ctx context.Context, db *sqlx.DB, query string, arg interface{}) (sql.Result, error) {
	if tx, ok := ctx.Value(txKey{}).(*sqlx.Tx); ok {
		return tx.NamedExecContext(ctx, query, arg)
	}
	return db.NamedExecContext(ctx, query, arg)
}
