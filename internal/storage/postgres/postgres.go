// Package postgres is implementation of storage interface.
package postgres

import (
	"database/sql"

	"github.com/jmoiron/sqlx"

	"github.com/Decentr-net/theseus/internal/storage"
)

//const uniqueViolationErrorCode = "23505"

type pg struct {
	db *sqlx.DB
}

// New creates new instance of pg.
func New(db *sql.DB) storage.Storage {
	return pg{
		db: sqlx.NewDb(db, "postgres"),
	}
}
