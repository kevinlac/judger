package judge

import (
	"context"
	"fmt"
	"worker/internal/config"
	"worker/internal/db"

	"github.com/google/uuid"
)

func JudgeSubmission(ctx context.Context, store *db.Store, cfg config.Config, submissionID uuid.UUID) error {
    sub, err := store.GetSubmission(ctx, submissionID)
    if err != nil {
        return fmt.Errorf("get submission: %w", err)
    }

    prob, err := store.GetProblem(ctx, sub.ProblemID)
    if err != nil {
        _ = store.SetSubmissionVerdict(ctx, submissionID, db.VerdictIE)
        return fmt.Errorf("get problem: %w", err)
    }

    if err := store.UpdateSubmissionStatus(ctx, submissionID, db.StatusRunning); err != nil {
        return fmt.Errorf("set running: %w", err)
        // DB write itself failed
    }

    compileErr, infraErr := compile(ctx, cfg, *sub)
    if infraErr != nil {
        _ = store.SetSubmissionVerdict(ctx, submissionID, db.VerdictIE)
        return fmt.Errorf("compile infra error: %w", infraErr)
    }
    if compileErr != nil {
        return store.SetSubmissionVerdict(ctx, submissionID, db.VerdictCE)
    }

    verdict, err := run(ctx, cfg, *sub, *prob)
    if err != nil {
        _ = store.SetSubmissionVerdict(ctx, submissionID, db.VerdictIE)
        return fmt.Errorf("run error: %w", err)
    }

    return store.SetSubmissionVerdict(ctx, submissionID, verdict)
}