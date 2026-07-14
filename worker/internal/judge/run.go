package judge

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"worker/internal/config"
	"worker/internal/db"
	"worker/internal/fsdata"
	"worker/internal/sandbox"
)

// VerdictAC here means that the program ran properly - no TLE or runtime errors from the user program.
// Any other result is a fail.
func genericRun(ctx context.Context, containerName string, args []string, stdin []byte, timeLimit time.Duration) (db.Verdict, string, time.Duration, error) {
    execCtx, cancel := context.WithTimeout(ctx, timeLimit)
    defer cancel()

    start := time.Now()
    stdout, _, err := sandbox.Exec(execCtx, containerName, args, stdin)
    elapsed := time.Since(start)

    if errors.Is(execCtx.Err(), context.DeadlineExceeded) {
        return db.VerdictTLE, "", timeLimit, nil
    }

    if err != nil {
        exitCode := -1
        var exitErr *exec.ExitError
        if errors.As(err, &exitErr) {
            exitCode = exitErr.ExitCode()
        }
        switch exitCode {
        case 137:
            out, inspectErr := exec.Command("docker", "inspect",
                "--format", "{{.State.OOMKilled}}", containerName).Output()
            if inspectErr == nil && strings.TrimSpace(string(out)) == "true" {
                // MLE
                return db.VerdictMLE, "", elapsed, nil
            }
            return db.VerdictRTE, "", elapsed, nil
        case 139:
            // segfault
            return db.VerdictRTE, "", elapsed, nil
        default:
            // other runtime error
            return db.VerdictRTE, "", elapsed, nil
        }
    }

    return db.VerdictAC, stdout, elapsed, nil
}

// standard string comparison check
func standardJudge(ctx context.Context, containerName, problemID string, runArgs []string, timeLimitMs int) (db.Verdict, error) {
    timeLimit := time.Duration(timeLimitMs) * time.Millisecond
    testcaseDir := fsdata.TestcasesDir("/data", problemID) // the worker itself reads from the bind mount

    entries, err := os.ReadDir(testcaseDir)
    if err != nil {
        return db.VerdictIE, err
    }

    for _, entry := range entries {
        if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".in") {
            continue
        }
        inputPath := filepath.Join(testcaseDir, entry.Name())
        outputPath := filepath.Join(testcaseDir, strings.TrimSuffix(entry.Name(), ".in")+".out")

        iData, err := os.ReadFile(inputPath)
        if err != nil {
            return db.VerdictIE, err
        }
        oData, err := os.ReadFile(outputPath)
        if err != nil {
            return db.VerdictIE, err
        }

        verdict, stdout, _, err := genericRun(ctx, containerName, runArgs, iData, timeLimit)
        if err != nil {
            return db.VerdictIE, err
        }
        if verdict != db.VerdictAC {
            return verdict, nil
        }

        userTokens := strings.Fields(stdout)
        expectedTokens := strings.Fields(string(oData))
        if len(userTokens) != len(expectedTokens) {
            return db.VerdictWA, nil
        }
        for i := range userTokens {
            if userTokens[i] != expectedTokens[i] {
                return db.VerdictWA, nil
            }
        }
    }
    return db.VerdictAC, nil
}

// to judge a submission. Will decide stuff like problem type, submission lang, etc.
func run(ctx context.Context, cfg config.Config, sub db.Submission, prob db.Problem) (db.Verdict, error) {
    containerName := fmt.Sprintf("j-run-%s", sub.ID)

    runArgs, image, err := runCommandFor(db.Lang(sub.Lang))
    if err != nil {
        return db.VerdictIE, err
    }

    opts := sandbox.RunOpts{
        Image:    image,
        Network:  "none",
        MemoryMB: prob.MemoryLimitMB,
        CPUs:     0.5,
        Mounts: []sandbox.Mount{{
            HostPath: fsdata.SubmissionDir(cfg.JudgeDataDir, sub.ID),
            Target:   "/app",
            ReadOnly: true, // compiled binary only, run phase is always read-only
        }},
    }

    if err := sandbox.Spawn(ctx, containerName, opts); err != nil {
        return db.VerdictIE, err
    }
    defer sandbox.Kill(containerName) // always clean up, even on early return

    switch prob.ProblemType {
    case db.ProblemTypeStandard:
        return standardJudge(ctx, containerName, sub.ProblemID, runArgs, prob.TimeLimitMs)
    // case string(db.ProblemTypeCustom):
    //     return customJudge(ctx, containerName, sub.ProblemID, runArgs, prob.TimeLimitMs)
    }
    return db.VerdictIE, fmt.Errorf("unknown problem type")
}

func runCommandFor(lang db.Lang) (args []string, image string, err error) {
    switch lang {
    case db.LangC:
        return []string{"./main"}, "c-judger", nil
    case db.LangCPP:
        return []string{"./main"}, "cpp-judger", nil
    case db.LangJava:
        return []string{"java", "Main"}, "java-judger", nil
    case db.LangPython:
        return []string{"python3", "solution.py"}, "python-judger", nil
    }
    return nil, "", fmt.Errorf("unrecognized language")
}