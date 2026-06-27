package repository

import (
	"context"
	"database/sql"
	"fmt"
)

// txKey is an unexported type so nothing outside this package can forge a
// transaction into a context accidentally.
type txKey struct{}

// InjectTx stores an active *sql.Tx inside ctx.
func InjectTx(ctx context.Context, tx *sql.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// ExtractTx retrieves the *sql.Tx stored by InjectTx.
// Returns nil, false if no transaction is present.
func ExtractTx(ctx context.Context) (*sql.Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(*sql.Tx)
	return tx, ok
}

// QueryExecutor defines a common interface for executing queries
// against *sql.DB or *sql.Tx.
type QueryExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// GetExecutor extracts a transaction from the context if it exists,
// otherwise falling back to the standard db connection pool.
func GetExecutor(ctx context.Context, db *sql.DB) QueryExecutor {
	if tx, ok := ExtractTx(ctx); ok {
		return tx
	}
	return db
}

// -----------------------------------------------------------------
// TxManager implementation
// -----------------------------------------------------------------

type txManager struct {
	db *sql.DB
}

// NewTransactionManager creates a new TransactionManager.
func NewTransactionManager(db *sql.DB) TransactionManager {
	return &txManager{db: db}
}

// WithTx begins a transaction, injects it into ctx, calls fn, and
// commits or rolls back depending on whether fn returned an error.
func (m *txManager) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	// Inject the tx so repositories can pick it up transparently.
	txCtx := InjectTx(ctx, tx)

	if err := fn(txCtx); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}
