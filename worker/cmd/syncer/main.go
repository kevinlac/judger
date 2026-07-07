package main

import (
	"context"
	"log"
	"os"
	"worker/internal/config"
	"worker/internal/db"
	"worker/internal/syncer"
)

// Usage: docker compose run --rm syncer

func main() {
	cfg := config.Load()
 
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "data"
	}
 
	rawDB, err := db.Connect(cfg.DatabaseUrl())
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer rawDB.Close()
 
	store := db.NewStore(rawDB)
 
	result, err := syncer.Sync(context.Background(), store, dataDir)
	if err != nil {
		log.Fatalf("sync: %v", err)
	}
 
	log.Printf("synced %d problem(s); %d submission(s) inserted, %d already present, %d skipped (no meta.json, assumed added via API)",
		result.ProblemsUpserted, result.SubmissionsInserted, result.SubmissionsSkipped, result.SubmissionsNoMeta)
 
	for _, syncErr := range result.Errors {
		log.Printf("sync warning: %v", syncErr)
	}
	if len(result.Errors) > 0 {
		os.Exit(1)
	}
}

