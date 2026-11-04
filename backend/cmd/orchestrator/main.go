package main

import (
	"context"
	"log"

	"wraithforge/backend/pkg/config"
	"wraithforge/backend/pkg/database"
	"wraithforge/backend/pkg/server"
)

func main() {
	cfg := config.Load()

	ctx := context.Background()
	db, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	s := server.New(db)
	log.Printf("listening on %s", cfg.ListenAddr)
	if err := s.Run(cfg.ListenAddr); err != nil {
		log.Fatal(err)
	}
}
