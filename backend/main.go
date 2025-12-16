package main

import (
	"backend/internal/store"
	"encoding/json"
	"fmt"
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
		} else if r.Method == http.MethodDelete {
			idStr := r.URL.Query().Get("id")
			if idStr == "" {
				http.Error(w, "Missing id", http.StatusBadRequest)
				return
			}
			var id int
			if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
				http.Error(w, "Invalid id", http.StatusBadRequest)
				return
			}
			if err := s.DeleteUser(r.Context(), id); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
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

	http.HandleFunc("/api/teams", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			teams, err := s.GetTeams(r.Context())
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(teams)
		} else if r.Method == http.MethodPost {
			var t store.Team
			if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			t, err := s.CreateTeam(r.Context(), t.Name)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(t)
		}
	})

	http.HandleFunc("/api/team_members", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			teamIDStr := r.URL.Query().Get("team_id")
			userIDStr := r.URL.Query().Get("user_id")

			if teamIDStr != "" {
				var teamID int
				if _, err := fmt.Sscanf(teamIDStr, "%d", &teamID); err != nil {
					http.Error(w, "Invalid team_id", http.StatusBadRequest)
					return
				}
				members, err := s.GetTeamMembers(r.Context(), teamID)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				json.NewEncoder(w).Encode(members)
			} else if userIDStr != "" {
				var userID int
				if _, err := fmt.Sscanf(userIDStr, "%d", &userID); err != nil {
					http.Error(w, "Invalid user_id", http.StatusBadRequest)
					return
				}
				members, err := s.GetUserTeams(r.Context(), userID)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				json.NewEncoder(w).Encode(members)
			} else {
				http.Error(w, "Missing team_id or user_id", http.StatusBadRequest)
				return
			}
		} else if r.Method == http.MethodPost {
			var tm store.TeamMember
			if err := json.NewDecoder(r.Body).Decode(&tm); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			tm, err := s.AddUserToTeam(r.Context(), tm.TeamID, tm.UserID, tm.Productivity)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(tm)
		} else if r.Method == http.MethodDelete {
			teamIDStr := r.URL.Query().Get("team_id")
			userIDStr := r.URL.Query().Get("user_id")
			if teamIDStr == "" || userIDStr == "" {
				http.Error(w, "Missing team_id or user_id", http.StatusBadRequest)
				return
			}
			var teamID, userID int
			if _, err := fmt.Sscanf(teamIDStr, "%d", &teamID); err != nil {
				http.Error(w, "Invalid team_id", http.StatusBadRequest)
				return
			}
			if _, err := fmt.Sscanf(userIDStr, "%d", &userID); err != nil {
				http.Error(w, "Invalid user_id", http.StatusBadRequest)
				return
			}
			if err := s.RemoveUserFromTeam(r.Context(), teamID, userID); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		}
	})

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
