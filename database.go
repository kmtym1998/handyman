package handyman

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/jackc/pgconn"
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

// PGErrCode returns the error code of Postgres errors.
// use pgerrcode package to get the error code of a Postgres error.
func PGErrCode(err error) (string, bool) {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code, true
	}

	return "", false
}
