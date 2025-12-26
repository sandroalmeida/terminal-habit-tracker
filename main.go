package main

import (
	"fmt"
	"habit-tracker/internal/config"
	"habit-tracker/internal/db"
	"log"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	database, err := db.Connect(cfg)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer database.Close()

	fmt.Println("Connection verification successful!")
}
