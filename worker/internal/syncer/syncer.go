package syncer

// brings the DB in line with what's on disk under data/.
//
// Usage: docker compose run --rm syncer
//
// Problems are authored by problem-setters directly on the filesystem
// (meta.json + testcases + problem.md), so for problems this package does
// an upsert: the filesystem is the source of truth, and the DB row is
// just a queryable index built from it.

// A submission directory only carries a
// meta.json when it's dev/test fixture data seeded straight onto disk
// (scripts/seed.py) before any DB row exists for it. A
// submission directory with no meta.json is assumed to already have a row
// (added via the API) and is left alone; one with a meta.json gets
// inserted if it isn't already present.

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/google/uuid"

	"worker/internal/db"
	"worker/internal/fsdata"
)

// mirrors data/problems/<id>/meta.json
type ProblemMeta struct {
	ID            string `json:"id"`
	TimeLimitMs   int    `json:"time_limit_ms"`
	MemoryLimitMB int    `json:"memory_limit_mb"`
	ProblemType   string `json:"problem_type"`
}

// mirrors data/submissions/<uuid>/meta.json
type SubmissionMeta struct {
	ProblemID string `json:"problem_id"`
	Lang      string `json:"lang"`
}

// Result summarizes what a Sync call did. Per-item failures are collected
// in Errors rather than aborting the whole run.
type Result struct {
	ProblemsUpserted    int
	SubmissionsInserted int
	SubmissionsSkipped  int // already present in the DB
	SubmissionsNoMeta   int // submissions that don't have meta.json present
	Errors              []error
}

func Sync(ctx context.Context, store *db.Store, dataDir string) (Result, error) {
	var result Result

	if err := syncProblems(ctx, store, dataDir, &result); err != nil {
		return result, fmt.Errorf("sync problems: %w", err)
	}
	if err := syncSubmissions(ctx, store, dataDir, &result); err != nil {
		return result, fmt.Errorf("sync submissions: %w", err)
	}
	return result, nil
}

func syncProblems(ctx context.Context, store *db.Store, dataDir string, result *Result) error {
	problemsDir := fsdata.ProblemsDir(dataDir)

	entries, err := os.ReadDir(problemsDir)
	if os.IsNotExist(err) {
		return nil // nothing to sync
	}
	if err != nil {
		return fmt.Errorf("read %s: %w", problemsDir, err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dirName := entry.Name()

		meta, err := readProblemMeta(dataDir, dirName)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("problem %s: %w", dirName, err))
			continue
		}
		if meta.ID != dirName {
			result.Errors = append(result.Errors,
				fmt.Errorf("problem %s: meta.json id %q does not match directory name", dirName, meta.ID))
			continue
		}

		if err := store.UpsertProblem(ctx, meta.ID, meta.TimeLimitMs, meta.MemoryLimitMB, db.ProblemType(meta.ProblemType)); err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("problem %s: %w", dirName, err))
			continue
		}
		result.ProblemsUpserted++
	}
	return nil
}

func syncSubmissions(ctx context.Context, store *db.Store, dataDir string, result *Result) error {
	submissionsDir := fsdata.SubmissionsDir(dataDir)

	entries, err := os.ReadDir(submissionsDir)
	if os.IsNotExist(err) {
		return nil // nothing to sync
	}
	if err != nil {
		return fmt.Errorf("read %s: %w", submissionsDir, err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dirName := entry.Name()

		id, err := uuid.Parse(dirName)
		if err != nil {
			result.Errors = append(result.Errors,
				fmt.Errorf("submission %s: directory name is not a valid UUID: %w", dirName, err))
			continue
		}

		// No meta.json means this submission was added the normal way,
		// through the API straight into Postgres — nothing to sync.
		if _, err := os.Stat(fsdata.SubmissionMetaPath(dataDir, id)); os.IsNotExist(err) {
			result.SubmissionsNoMeta++
			continue
		}

		meta, err := readSubmissionMeta(dataDir, id)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("submission %s: %w", dirName, err))
			continue
		}

		if _, err := fsdata.SubmissionSourcePath(dataDir, id, db.Lang(meta.Lang)); err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("submission %s: %w", dirName, err))
			continue
		}

		inserted, err := store.InsertSubmissionIfAbsent(ctx, id, meta.ProblemID, db.Lang(meta.Lang))
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("submission %s: %w", dirName, err))
			continue
		}
		if inserted {
			result.SubmissionsInserted++
		} else {
			result.SubmissionsSkipped++
		}
	}
	return nil
}

func readProblemMeta(dataDir, problemID string) (ProblemMeta, error) {
	var meta ProblemMeta
	raw, err := os.ReadFile(fsdata.ProblemMetaPath(dataDir, problemID))
	if err != nil {
		return meta, fmt.Errorf("read meta.json: %w", err)
	}
	if err := json.Unmarshal(raw, &meta); err != nil {
		return meta, fmt.Errorf("parse meta.json: %w", err)
	}
	return meta, nil
}

func readSubmissionMeta(dataDir string, id uuid.UUID) (SubmissionMeta, error) {
	var meta SubmissionMeta
	raw, err := os.ReadFile(fsdata.SubmissionMetaPath(dataDir, id))
	if err != nil {
		return meta, fmt.Errorf("read meta.json: %w", err)
	}
	if err := json.Unmarshal(raw, &meta); err != nil {
		return meta, fmt.Errorf("parse meta.json: %w", err)
	}
	return meta, nil
}