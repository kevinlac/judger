package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

const (
	ProbC = "C"
	ProbCPP = "C++"
	ProbPython = "Python"
	ProbJava = "Java"
)

const (
	ProbTypeStandard = "standard"
	ProbTypeCustom = "custom"
)

type Problem struct {
	ID            string
	TimeLimitMs   int
	MemoryLimitMB int
	ProblemType   string
	CreatedAt     time.Time
}

func (s *Store) InsertProblem(id string, timeLimit int, memoryLimit int, problemType string) error {
	query := `
    INSERT INTO problems (id, time_limit_ms, memory_limit_mb, problem_type)
    VALUES ($1, $2, $3, $4);`
	_, err := s.db.Exec(query, id, timeLimit, memoryLimit, problemType)
	return err
}

// insert or update if exists already
func (s *Store) UpsertProblem(ctx context.Context, id string, timeLimit int, memoryLimit int, problemType string) error {
    query := `
    INSERT INTO problems (id, time_limit_ms, memory_limit_mb, problem_type)
    VALUES ($1, $2, $3, $4)
    ON CONFLICT (id) DO UPDATE SET
        time_limit_ms = EXCLUDED.time_limit_ms,
        memory_limit_mb = EXCLUDED.memory_limit_mb,
        problem_type = EXCLUDED.problem_type;`
    _, err := s.db.ExecContext(ctx, query, id, timeLimit, memoryLimit, problemType)
    return err
}

// fetches a single problem by id
func (s *Store) GetProblem(ctx context.Context, id string) (*Problem, error) {
	query := `
    SELECT id, time_limit_ms, memory_limit_mb, problem_type, created_at
    FROM problems
    WHERE id = $1;`

	var p Problem
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID,
		&p.TimeLimitMs,
		&p.MemoryLimitMB,
		&p.ProblemType,
		&p.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get problem %s: %w", id, err)
	}
	return &p, nil
}

func (s *Store) ListProblems(ctx context.Context) ([]Problem, error) {
	query := `
    SELECT id, time_limit_ms, memory_limit_mb, problem_type, created_at
    FROM problems
    ORDER BY created_at DESC;`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list problems: %w", err)
	}
	defer rows.Close()

	var problems []Problem
	for rows.Next() {
		var p Problem
		if err := rows.Scan(
			&p.ID,
			&p.TimeLimitMs,
			&p.MemoryLimitMB,
			&p.ProblemType,
			&p.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan problem: %w", err)
		}
		problems = append(problems, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list problems: %w", err)
	}
	return problems, nil
}

// removes a problem by id. Fails if submissions still
// reference it (fk_problem_id is ON DELETE RESTRICT).
func (s *Store) DeleteProblem(ctx context.Context, id string) error {
	query := `DELETE FROM problems WHERE id = $1;`
	res, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete problem %s: %w", id, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}