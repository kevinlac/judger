package main

import (
	"fmt"
	"log"
	"time"
	"worker/internal/config"
	"worker/internal/db"

	"github.com/google/uuid"
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
    if err := store.InsertProblem("example-problem", 1500, 256, "standard"); err != nil {
		log.Fatalf("insert problem: %v", err)
	}
	submissionID := uuid.New()
	if err := store.InsertSubmission(submissionID, "example-problem", "C++"); err != nil {
		log.Fatalf("insert submission: %v", err)
	}

	fmt.Printf("Added stuff to db successfully.\n")

	for {
		time.Sleep(1 * time.Second)
	}
}