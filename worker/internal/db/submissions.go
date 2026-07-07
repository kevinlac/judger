package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// chk_sub_status constraint
const (
	StatusQueued  = "queued"
	StatusRunning = "running"
	StatusJudged  = "judged"
)

// chk_sub_verdict constraint
const (
	VerdictAC  = "AC"
	VerdictWA  = "WA"
	VerdictTLE = "TLE"
	VerdictMLE = "MLE"
	VerdictRTE = "RTE"
	VerdictIE  = "IE"
)

type Submission struct {
	ID               uuid.UUID
	ProblemID        string
	Lang             string
	ProcessingStatus string
	Verdict          sql.NullString
	SubmittedAt      time.Time
	JudgedAt         sql.NullTime
}

// creates a new submission row with status queued
func (s *Store) InsertSubmission(id uuid.UUID, problemID string, lang string) error {
	query := `
    INSERT INTO submissions (id, problem_id, lang, processing_status)
    VALUES ($1, $2, $3, 'queued');`
	_, err := s.db.Exec(query, id, problemID, lang)
	return err
}

// fetches a single submission by id
func (s *Store) GetSubmission(ctx context.Context, id uuid.UUID) (*Submission, error) {
	query := `
    SELECT id, problem_id, lang, processing_status, verdict, submitted_at, judged_at
    FROM submissions
    WHERE id = $1;`

	var sub Submission
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&sub.ID,
		&sub.ProblemID,
		&sub.Lang,
		&sub.ProcessingStatus,
		&sub.Verdict,
		&sub.SubmittedAt,
		&sub.JudgedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get submission %s: %w", id, err)
	}
	return &sub, nil
}

// update a submission's processing_status (e.g. queued -> running)
func (s *Store) UpdateSubmissionStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `
    UPDATE submissions
    SET processing_status = $2
    WHERE id = $1;`
	res, err := s.db.ExecContext(ctx, query, id, status)
	if err != nil {
		return fmt.Errorf("update submission %s status: %w", id, err)
	}
	return rowsAffectedOrNotFound(res)
}

// adds final verdict for a submission, marking it judged and stamping judged_at to now
func (s *Store) SetSubmissionVerdict(ctx context.Context, id uuid.UUID, verdict string) error {
	query := `
    UPDATE submissions
    SET verdict = $2,
        processing_status = 'judged',
        judged_at = NOW()
    WHERE id = $1;`
	res, err := s.db.ExecContext(ctx, query, id, verdict)
	if err != nil {
		return fmt.Errorf("set verdict for submission %s: %w", id, err)
	}
	return rowsAffectedOrNotFound(res)
}
