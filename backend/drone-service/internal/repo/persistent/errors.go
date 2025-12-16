package repo

import (
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func isPgForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23503"
	}
	return false
}

func isNoRows(err error) bool {
	if errors.Is(err, pgx.ErrNoRows) {
		return true
	}
	if errors.Is(err, sql.ErrNoRows) {
		return true
	}
	return false
}
