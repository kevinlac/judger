package judge

import (
	"context"
	"errors"
	"fmt"
	"time"
	"worker/internal/config"
	"worker/internal/db"
	"worker/internal/fsdata"
	"worker/internal/sandbox"
)

var ErrCompileTimeout = errors.New("Could not compile within allocated time")

func Compile(ctx context.Context, cfg config.Config, sub db.Submission) error {
	fmt.Printf("Compiling submission for %s with ID %s\n", sub.ProblemID, sub.ID)
    compileCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()
	
	var submissionImage string
	var compileArgs []string
	// switch the image based on submission lang
	filename, err := fsdata.SourceFilename(sub.Lang)
	if err != nil {
		return err
	}
	switch sub.Lang {
	case db.ProbC:
		submissionImage = "c-judger"
		compileArgs = []string{"gcc", "-o", "main", filename}
	case db.ProbCPP:
		submissionImage = "cpp-judger"
		compileArgs = []string{"g++", "-o", "main", filename}
	case db.ProbJava:
		submissionImage = "java-judger"
		compileArgs = []string{"javac", filename}
	case db.ProbPython:
		return nil // no need to compile
	}

	if (submissionImage == "") {
		return fmt.Errorf("submission struct language is not recognized")
	}

	myMount := sandbox.Mount{
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

    _, _, err = sandbox.RunOnce(compileCtx, options)
	if err != nil {
		return err
	}
    if errors.Is(compileCtx.Err(), context.DeadlineExceeded) {
        return ErrCompileTimeout
    }
	return nil
}