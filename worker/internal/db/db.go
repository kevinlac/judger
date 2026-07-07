package db

import (
	"database/sql"
	"errors"

	_ "github.com/lib/pq"
)

var ErrNotFound = errors.New("No matching row was found when querying the DB.")

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func Connect(databaseUrl string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseUrl)
    if err != nil {
        return nil, err
    }
    if err := db.Ping(); err != nil {
        return nil, err
    }
    return db, nil
}

// turns a zero-row-affected update/delete result into ErrNotFound
func rowsAffectedOrNotFound(res sql.Result) error {
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}