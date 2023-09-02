package handyman

import (
	"context"
	"database/sql"
	"fmt"
)

// Tx executes the given function in a transaction.
func Tx(ctx context.Context, db *sql.DB, opts *sql.TxOptions, f func(tx *sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, opts)
	if err != nil {
		return fmt.Errorf("begin error: %w", err)
	}

	if err := f(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("rollback error: %+v (outer error: %+v)", rbErr, err)
		}

		return fmt.Errorf("tx rollback: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit error: %w", err)
	}

	return nil
}
