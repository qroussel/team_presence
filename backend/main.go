package main

import (
	"backend/internal/api"
	"backend/internal/store"
	"context"
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// Default to local development DB; in production, always set DATABASE_URL
		dbURL = "postgres://presence_user:presence_password@localhost:5432/presence_db?sslmode=disable"
	}

	// Create a context for the store connection
	ctx := context.Background()
	s, err := store.NewStore(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to create store: %v", err)
	}
	defer s.Close()

	if err := s.InitSchema("schema.sql"); err != nil {
		log.Printf("Warning: Failed to init schema (db might not be ready): %v", err)
	}

	if err := s.SeedAdmin(context.Background()); err != nil {
		log.Printf("Warning: Failed to seed admin: %v", err)
	}

	server := api.NewServer(s)

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, server); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
