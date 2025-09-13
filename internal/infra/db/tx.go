package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// context key for sql.Tx
type ctxKey string

const txKey ctxKey = "sqlTx"

func withTx(ctx context.Context, tx *sql.Tx) context.Context {
	return context.WithValue(ctx, txKey, tx)
}

func getTx(ctx context.Context) *sql.Tx {
	if v := ctx.Value(txKey); v != nil {
		if tx, ok := v.(*sql.Tx); ok {
			return tx
		}
	}
	return nil
}

// TxManager implements domain.Tx using database/sql transactions.
type TxManager struct {
	DB *sql.DB
}

// Do begins a transaction with default options.
func (m *TxManager) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	return m.DoWithOpts(ctx, nil, fn)
}

// DoWithOpts allows specifying *sql.TxOptions (isolation, read-only, etc.).
func (m *TxManager) DoWithOpts(ctx context.Context, opts *sql.TxOptions, fn func(ctx context.Context) error) (err error) {
	if m == nil || m.DB == nil {
		return fmt.Errorf("tx manager not initialized")
	}

	tx, err := m.DB.BeginTx(ctx, opts)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	// panic セーフティ：必ず Rollback を試みてから再panic
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	ctxTx := withTx(ctx, tx)

	if err = fn(ctxTx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			err = errors.Join(err, fmt.Errorf("rollback after fn error: %w", rbErr))
		}
		return fmt.Errorf("tx fn: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}
