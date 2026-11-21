package sqlite

import (
	"database/sql"

	"alp-drive/internal/infrastructure/sqlite/sqlcgen"
)

const listLimit = 10_000

type Store struct {
	db      *sql.DB
	queries *sqlcgen.Queries
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db, queries: sqlcgen.New(db)}
}
