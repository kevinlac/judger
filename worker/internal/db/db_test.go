package db

import (
	"context"
	"database/sql"
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