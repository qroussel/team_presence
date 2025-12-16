package main

import (
	"backend/internal/store"
	"encoding/json"
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
		dbURL = "postgres://presence_user:presence_password@localhost:5432/presence_db?sslmode=disable"
	}

	s, err := store.New(dbURL)
	if err != nil {
		log.Fatalf("Failed to create store: %v", err)
	}
	defer s.Close()

	if err := s.InitSchema("schema.sql"); err != nil {
		log.Printf("Warning: Failed to init schema (db might not be ready): %v", err)
	}

	http.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	http.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			users, err := s.GetUsers(r.Context())
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(users)
		} else if r.Method == http.MethodPost {
			var u store.User
			if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			u, err := s.CreateUser(r.Context(), u)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(u)
		}
	})

	http.HandleFunc("/api/presence", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			start := r.URL.Query().Get("start")
			end := r.URL.Query().Get("end")
			if start == "" || end == "" {
				http.Error(w, "Missing start or end date", http.StatusBadRequest)
				return
			}
			p, err := s.GetPresence(r.Context(), start, end)
			if err != nil {
				log.Printf("GetPresence error: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(p)
		} else if r.Method == http.MethodPost {
			var p store.Presence
			if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			p, err := s.UpsertPresence(r.Context(), p)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(p)
		}
	})

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
