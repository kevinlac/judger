package db

import (
	"database/sql"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

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

func (s *Store) InsertProblem(id string, timeLimit int, memoryLimit int, problemType string) error {
	query := `
	INSERT INTO problems (id, time_limit_ms, memory_limit_mb, problem_type)
	VALUES ($1, $2, $3, $4);`
	_, err := s.db.Exec(query, id, timeLimit, memoryLimit, problemType)
	return err
}

func (s *Store) InsertSubmission(id uuid.UUID, problemID string, lang string) error {
	query := `
	INSERT INTO submissions (id, problem_id, lang, processing_status)
	VALUES ($1, $2, $3, 'queued');`
	_, err := s.db.Exec(query, id, problemID, lang)
	return err
}