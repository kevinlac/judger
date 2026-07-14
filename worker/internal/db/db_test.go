package db

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var store *Store

func (s *Store) resetForTest(t *testing.T) {
    t.Helper()
    _, err := s.db.Exec("TRUNCATE problems, submissions RESTART IDENTITY CASCADE")
    if err != nil {
        t.Fatalf("truncate: %v", err)
    }
}

func TestMain(m *testing.M) {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:15-alpine",
		postgres.WithDatabase("test_db"),
		postgres.WithUsername("test_user"),
		postgres.WithPassword("test_pass"),
		testcontainers.WithWaitStrategy(
			// Postgres logs this message twice on first init: once for the
			// temporary bootstrap instance, once for the real server.
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30 * time.Second)),
	)
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalf("failed to get connection string: %s", err)
	}

	rawDB, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("failed to open db: %s", err)
	}
	defer rawDB.Close()
	
	// init test db
	schema, err := os.ReadFile("../../../db-init/init.sql")
	if err != nil {
		log.Fatalf("failed to read schema file: %s", err)
	}

	_, err = rawDB.Exec(string(schema))
	if err != nil {
		log.Fatalf("failed to initialize database schema: %s", err)
	}

	store = NewStore(rawDB)
	code := m.Run()

	rawDB.Close()
	_ = pgContainer.Terminate(ctx)

	os.Exit(code)
}

func TestInsertProblemSubmission(t *testing.T) {
	store.resetForTest(t)
    if err := store.InsertProblem("example-problem", 1500, 256, "standard"); err != nil {
		t.Errorf("insert problem: %v", err)
	}
	submissionID := uuid.New()
	if err := store.InsertSubmission(submissionID, "example-problem", "C++"); err != nil {
		t.Errorf("insert submission: %v", err)
	}
}

func TestGetProblem(t *testing.T) {
	store.resetForTest(t)
	ctx := context.Background()
 
	if err := store.InsertProblem("example-problem", 1500, 256, "standard"); err != nil {
		t.Fatalf("insert problem: %v", err)
	}
 
	p, err := store.GetProblem(ctx, "example-problem")
	if err != nil {
		t.Fatalf("get problem: %v", err)
	}
	if p.ID != "example-problem" {
		t.Errorf("ID = %q, want %q", p.ID, "example-problem")
	}
	if p.TimeLimitMs != 1500 {
		t.Errorf("TimeLimitMs = %d, want 1500", p.TimeLimitMs)
	}
	if p.MemoryLimitMB != 256 {
		t.Errorf("MemoryLimitMB = %d, want 256", p.MemoryLimitMB)
	}
	if p.ProblemType != "standard" {
		t.Errorf("ProblemType = %q, want %q", p.ProblemType, "standard")
	}
	if p.CreatedAt.IsZero() {
		t.Errorf("CreatedAt should not be zero")
	}
}

func TestGetProblem_NotFound(t *testing.T) {
	store.resetForTest(t)
	ctx := context.Background()
 
	_, err := store.GetProblem(ctx, "does-not-exist")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}


func TestListProblems(t *testing.T) {
	store.resetForTest(t)
	ctx := context.Background()
 
	if err := store.InsertProblem("problem-a", 1000, 128, "standard"); err != nil {
		t.Fatalf("insert problem-a: %v", err)
	}
	if err := store.InsertProblem("problem-b", 2000, 512, "custom"); err != nil {
		t.Fatalf("insert problem-b: %v", err)
	}
 
	problems, err := store.ListProblems(ctx)
	if err != nil {
		t.Fatalf("list problems: %v", err)
	}
	if len(problems) != 2 {
		t.Fatalf("got %d problems, want 2", len(problems))
	}
 
	seen := map[string]bool{}
	for _, p := range problems {
		seen[p.ID] = true
	}
	if !seen["problem-a"] || !seen["problem-b"] {
		t.Errorf("ListProblems missing expected rows: %+v", problems)
	}
}

func TestDeleteProblem(t *testing.T) {
	store.resetForTest(t)
	ctx := context.Background()
 
	if err := store.InsertProblem("example-problem", 1500, 256, "standard"); err != nil {
		t.Fatalf("insert problem: %v", err)
	}
 
	if err := store.DeleteProblem(ctx, "example-problem"); err != nil {
		t.Fatalf("delete problem: %v", err)
	}
 
	_, err := store.GetProblem(ctx, "example-problem")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("get after delete: err = %v, want ErrNotFound", err)
	}
}

func TestDeleteReferencedProblem(t *testing.T) {
	store.resetForTest(t)
	ctx := context.Background()
 
	if err := store.InsertProblem("example-problem", 1500, 256, "standard"); err != nil {
		t.Fatalf("insert problem: %v", err)
	}
	if err := store.InsertSubmission(uuid.New(), "example-problem", "C++"); err != nil {
		t.Fatalf("insert submission: %v", err)
	}
 
	// fk_problem_id is ON DELETE RESTRICT, so this must fail while a submission still references the problem
	err := store.DeleteProblem(ctx, "example-problem")
	if err == nil {
		t.Fatalf("expected delete to fail due to FK restriction, got nil error")
	}
	if errors.Is(err, ErrNotFound) {
		t.Errorf("expected FK violation, got ErrNotFound")
	}
}

func TestUpdateSubmissionStatus(t *testing.T) {
	store.resetForTest(t)
	ctx := context.Background()
 
	if err := store.InsertProblem("example-problem", 1500, 256, "standard"); err != nil {
		t.Fatalf("insert problem: %v", err)
	}
	submissionID := uuid.New()
	if err := store.InsertSubmission(submissionID, "example-problem", "C++"); err != nil {
		t.Fatalf("insert submission: %v", err)
	}
 
	if err := store.UpdateSubmissionStatus(ctx, submissionID, StatusRunning); err != nil {
		t.Fatalf("update submission status: %v", err)
	}
 
	sub, err := store.GetSubmission(ctx, submissionID)
	if err != nil {
		t.Fatalf("get submission: %v", err)
	}
	if sub.ProcessingStatus != StatusRunning {
		t.Errorf("ProcessingStatus = %q, want %q", sub.ProcessingStatus, StatusRunning)
	}
}

func TestSetSubmissionVerdict(t *testing.T) {
	store.resetForTest(t)
	ctx := context.Background()
 
	if err := store.InsertProblem("example-problem", 1500, 256, "standard"); err != nil {
		t.Fatalf("insert problem: %v", err)
	}
	submissionID := uuid.New()
	if err := store.InsertSubmission(submissionID, "example-problem", "C++"); err != nil {
		t.Fatalf("insert submission: %v", err)
	}
 
	if err := store.SetSubmissionVerdict(ctx, submissionID, VerdictAC); err != nil {
		t.Fatalf("set submission verdict: %v", err)
	}
 
	sub, err := store.GetSubmission(ctx, submissionID)
	if err != nil {
		t.Fatalf("get submission: %v", err)
	}
	if sub.ProcessingStatus != StatusJudged {
		t.Errorf("ProcessingStatus = %q, want %q", sub.ProcessingStatus, StatusJudged)
	}
	if sub.Verdict == nil {
        t.Fatalf("Verdict should be set once judged")
    }
    if *sub.Verdict != VerdictAC {
        t.Errorf("Verdict = %q, want %q", *sub.Verdict, VerdictAC)
    }
	if !sub.JudgedAt.Valid {
		t.Errorf("JudgedAt should be set once judged")
	}
}
