package main

import (
	"log"

	"github.com/onecare/backend/internal/config"
	"github.com/onecare/backend/internal/database"
	"github.com/onecare/backend/internal/router"
)

func main() {
	// Load config from .env
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Connect to MySQL
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Printf("connected to MySQL at %s:%s", cfg.DBHost, cfg.DBPort)

	// Boot HTTP server
	r := router.Setup(db, cfg)

	log.Printf("OneCare API listening on :%s (env: %s)", cfg.AppPort, cfg.AppEnv)
	if err := r.Run(":" + cfg.AppPort); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
