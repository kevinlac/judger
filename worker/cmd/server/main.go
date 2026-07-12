package main

import (
	"context"
	"fmt"
	"log"
	"worker/internal/config"
	"worker/internal/db"
	"worker/internal/fsdata"
	"worker/internal/judge"
	"worker/internal/sandbox"
)

func main() {
	cfg := config.Load()
	
	rawDB, err := db.Connect(cfg.DatabaseUrl())
	if err != nil {
        log.Fatalf("db connect: %v", err)
    }
	defer rawDB.Close()

	fmt.Printf("Connected to db successfully.\n")

	store := db.NewStore(rawDB)

	// should only return our one submission from example data
	allSubmissions, err := store.ListSubmissionsByProblem(context.Background(), "sum-two-numbers")
	fmt.Printf("No. of submissions for sum-two-numbers: %d\n", len(allSubmissions))

	err = judge.Compile(context.Background(), cfg, allSubmissions[0])
	if err != nil {
		fmt.Printf("Error compiling!: %v\n", err)
	} else {
		fmt.Printf("Compiled submission successfully.\n") 
	}

	// test of spawn, exec, kill
	ctx := context.Background()
	opts := sandbox.RunOpts{
		Image:   "cpp-judger",
		Network: "none",
		MemoryMB: 256,
		CPUs: 0.25,
		Mounts: []sandbox.Mount{{
			HostPath: fsdata.SubmissionDir(cfg.JudgeDataDir, allSubmissions[0].ID),
			Target:   "/app",
			ReadOnly: true,
		}},
	}
	if err := sandbox.Spawn(ctx, "test-sandbox", opts); err != nil {
		log.Fatal(err)
	}
	defer sandbox.Kill("test-sandbox")

	out, errOut, err := sandbox.Exec(ctx, "test-sandbox", []string{"./main"}, []byte("10 3\n"))
	fmt.Printf("stdout: %v\n stderr: %v\n err: %v\n", out, errOut, err)
}