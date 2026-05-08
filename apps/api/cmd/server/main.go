package main

import (
	"log"

	"github.com/aash/mtracker/apps/api/internal/config"
	"github.com/aash/mtracker/apps/api/internal/database"
	"github.com/aash/mtracker/apps/api/internal/router"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}

	r := router.Setup(db, cfg)

	log.Printf("Mtracker API listening on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
