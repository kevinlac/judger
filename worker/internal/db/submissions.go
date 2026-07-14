package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Lang string

const (
    LangC      Lang = "C"
    LangCPP    Lang = "C++"
    LangPython Lang = "Python"
    LangJava   Lang = "Java"
)

type SubStatus string

// chk_sub_status constraint
const (
    StatusQueued  SubStatus = "queued"
    StatusRunning SubStatus = "running"
    StatusJudged  SubStatus = "judged"
)

type Verdict string

// chk_sub_verdict constraint
const (
    VerdictAC  Verdict = "AC"
    VerdictWA  Verdict = "WA"
    VerdictTLE Verdict = "TLE"
    VerdictMLE Verdict = "MLE"
    VerdictRTE Verdict = "RTE"
    VerdictIE  Verdict = "IE"
	VerdictCE  Verdict = "CE"
)

type Submission struct {
	ID               uuid.UUID
	ProblemID        string
	Lang             Lang
	ProcessingStatus SubStatus
	Verdict          *Verdict // verdict is not determined till judged
	SubmittedAt      time.Time
	JudgedAt         sql.NullTime
}

// creates a new submission row with status queued
func (s *Store) InsertSubmission(id uuid.UUID, problemID string, lang Lang) error {
	query := `
    INSERT INTO submissions (id, problem_id, lang, processing_status)
    VALUES ($1, $2, $3, 'queued');`
	_, err := s.db.Exec(query, id, problemID, lang)
	return err
}

// inserts a submission with status queued if no
// row with that id exists yet, and is a no-op otherwise.
func (s *Store) InsertSubmissionIfAbsent(ctx context.Context, id uuid.UUID, problemID string, lang Lang) (bool, error) {
    query := `
    INSERT INTO submissions (id, problem_id, lang, processing_status)
    VALUES ($1, $2, $3, 'queued')
    ON CONFLICT (id) DO NOTHING;`
    res, err := s.db.ExecContext(ctx, query, id, problemID, lang)
    if err != nil {
        return false, err
    }
    n, err := res.RowsAffected()
    if err != nil {
        return false, err
    }
    return n > 0, nil
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

// ListSubmissionsByProblem returns all submissions for a given problem,
// most recent first.
func (s *Store) ListSubmissionsByProblem(ctx context.Context, problemID string) ([]Submission, error) {
	query := `
    SELECT id, problem_id, lang, processing_status, verdict, submitted_at, judged_at
    FROM submissions
    WHERE problem_id = $1
    ORDER BY submitted_at DESC;`
 
	rows, err := s.db.QueryContext(ctx, query, problemID)
	if err != nil {
		return nil, fmt.Errorf("list submissions for problem %s: %w", problemID, err)
	}
	defer rows.Close()
 
	var subs []Submission
	for rows.Next() {
		var sub Submission
		if err := rows.Scan(
			&sub.ID,
			&sub.ProblemID,
			&sub.Lang,
			&sub.ProcessingStatus,
			&sub.Verdict,
			&sub.SubmittedAt,
			&sub.JudgedAt,
		); err != nil {
			return nil, fmt.Errorf("scan submission: %w", err)
		}
		subs = append(subs, sub)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list submissions for problem %s: %w", problemID, err)
	}
	return subs, nil
}


// update a submission's processing_status (e.g. queued -> running)
func (s *Store) UpdateSubmissionStatus(ctx context.Context, id uuid.UUID, status SubStatus) error {
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
func (s *Store) SetSubmissionVerdict(ctx context.Context, id uuid.UUID, verdict Verdict) error {
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
