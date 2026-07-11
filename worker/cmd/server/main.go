package main

import (
	"context"
	"fmt"
	"log"
	"time"
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

	err = judge.Compile(context.Background(), cfg, allSubmissions[0])
	if err != nil {
		fmt.Printf("Error compiling!: %v\n", err)
	} else {
		fmt.Printf("Compiled submission successfully.\n") 
	}

	for {
		time.Sleep(1 * time.Second)
	}
}