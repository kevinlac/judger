package main

import (
	"context"
	"fmt"
	"log"
	"worker/internal/config"
	"worker/internal/db"
	"worker/internal/judge"
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
	
	for _, sub := range allSubmissions {
		judge.JudgeSubmission(context.Background(), store, cfg, sub.ID)
	}
}