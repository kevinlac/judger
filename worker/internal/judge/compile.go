package judge

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"time"
	"worker/internal/config"
	"worker/internal/db"
	"worker/internal/fsdata"
	"worker/internal/sandbox"
)

var ErrCompileTimeout = errors.New("Could not compile within allocated time")

func compile(ctx context.Context, cfg config.Config, sub db.Submission) (compileErr error, infraErr error) {
	fmt.Printf("Compiling submission for %s with ID %s\n", sub.ProblemID, sub.ID)
    compileCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()
	
	var submissionImage string
	var compileArgs []string
	// switch the image based on submission lang
	filename, err := fsdata.SourceFilename(sub.Lang)
	if err != nil {
		return nil, err
	}
	switch sub.Lang {
	case db.LangC:
		submissionImage = "c-judger"
		compileArgs = []string{"gcc", "-o", "main", filename}
	case db.LangCPP:
		submissionImage = "cpp-judger"
		compileArgs = []string{"g++", "-o", "main", filename}
	case db.LangJava:
		submissionImage = "java-judger"
		compileArgs = []string{"javac", filename}
	case db.LangPython:
		return nil, nil // no need to compile
	}

	if (submissionImage == "") {
		return nil, fmt.Errorf("submission struct language is not recognized")
	}

	myMount := sandbox.Mount{
		// since we are running stuff on host machine's docker socket
		// we mount from our original path in .env
		HostPath: fsdata.SubmissionDir(cfg.JudgeDataDir, sub.ID),
		Target: "/app",
		ReadOnly: false,
	}

	options := sandbox.RunOpts{
		Image: submissionImage,
		Mounts: []sandbox.Mount{myMount},
		MemoryMB: 256,
		CPUs: 0.25,
		Network: "none",
		Args: compileArgs,
	}

	_, stderr, err := sandbox.RunOnce(compileCtx, options)
    if errors.Is(compileCtx.Err(), context.DeadlineExceeded) {
        return nil, ErrCompileTimeout // compiling taking too long is the user's fault
    }
    if err != nil {
        var exitErr *exec.ExitError
        if errors.As(err, &exitErr) {
            // compiler ran and rejected the code, stderr has the reason
            return fmt.Errorf("compile error: %s", stderr), nil
        }
        // err wasn't an ExitError at all, e.g. couldn't even start the container, mount failed, etc — that's IE
        return nil, fmt.Errorf("infra error running compiler: %w", err)
    }
    return nil, nil
}