package main

import (
	"fmt"
	"log"
	"time"
	"worker/internal/config"
	"worker/internal/db"
)

func main() {
	cfg := config.Load()
	
	rawDB, err := db.Connect(cfg.DatabaseUrl())
	if err != nil {
        log.Fatalf("db connect: %v", err)
    }
	defer rawDB.Close()

	fmt.Printf("Connected to db successfully.\n")

	for {
		time.Sleep(1 * time.Second)
	}
}