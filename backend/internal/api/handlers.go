package api

import (
	"backend/internal/store"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (s *Server) handleUsers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		users, err := s.store.GetUsers(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Convert to DTOs
		var response []UserResponse
		for _, u := range users {
			response = append(response, UserResponse{
				ID:        int(u.ID),
				Name:      u.Name,
				Email:     u.Email,
				AvatarURL: u.AvatarUrl,
				CreatedAt: u.CreatedAt,
			})
		}
		json.NewEncoder(w).Encode(response)
	case http.MethodPost:
		var req CreateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if req.Name == "" || req.Email == "" {
			http.Error(w, "Name and Email are required", http.StatusBadRequest)
			return
		}

		params := store.CreateUserParams{
			Name:      req.Name,
			Email:     req.Email,
			AvatarUrl: pgtype.Text{String: req.AvatarURL, Valid: req.AvatarURL != ""},
		}

		created, err := s.store.CreateUser(r.Context(), params)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(UserResponse{
			ID:        int(created.ID),
			Name:      req.Name,
			Email:     req.Email,
			AvatarURL: req.AvatarURL,
			CreatedAt: created.CreatedAt,
		})
	case http.MethodDelete:
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
		if err := s.store.DeleteUser(r.Context(), int32(id)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handlePresence(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		start := r.URL.Query().Get("start")
		end := r.URL.Query().Get("end")
		if start == "" || end == "" {
			http.Error(w, "Missing start or end date", http.StatusBadRequest)
			return
		}

		startDate, err := time.Parse("2006-01-02", start)
		if err != nil {
			http.Error(w, "Invalid start date", http.StatusBadRequest)
			return
		}
		endDate, err := time.Parse("2006-01-02", end)
		if err != nil {
			http.Error(w, "Invalid end date", http.StatusBadRequest)
			return
		}

		entries, err := s.store.GetPresence(r.Context(), store.GetPresenceParams{
			Column1: startDate,
			Column2: endDate,
		})
		if err != nil {
			log.Printf("GetPresence error: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		var response []PresenceResponse
		for _, p := range entries {
			response = append(response, PresenceResponse{
				ID:        int(p.ID),
				UserID:    int(p.UserID),
				Date:      p.Date,
				StatusAM:  p.StatusAm,
				StatusPM:  p.StatusPm,
				CreatedAt: p.CreatedAt,
			})
		}
		json.NewEncoder(w).Encode(response)
	case http.MethodPost:
		var req UpsertPresenceRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		date, err := time.Parse("2006-01-02", req.Date)
		if err != nil {
			http.Error(w, "Invalid date format", http.StatusBadRequest)
			return
		}

		params := store.UpsertPresenceParams{
			UserID:   int32(req.UserID),
			Date:     date,
			StatusAm: pgtype.Text{String: req.StatusAM, Valid: req.StatusAM != ""},
			StatusPm: pgtype.Text{String: req.StatusPM, Valid: req.StatusPM != ""},
		}

		p, err := s.store.UpsertPresence(r.Context(), params)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(PresenceResponse{
			ID:        int(p.ID),
			UserID:    req.UserID,
			Date:      req.Date,
			StatusAM:  req.StatusAM,
			StatusPM:  req.StatusPM,
			CreatedAt: p.CreatedAt,
		})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleTeams(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		teams, err := s.store.GetTeams(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		var response []TeamResponse
		for _, t := range teams {
			response = append(response, TeamResponse{
				ID:        int(t.ID),
				Name:      t.Name,
				CreatedAt: t.CreatedAt,
			})
		}
		json.NewEncoder(w).Encode(response)
	case http.MethodPost:
		var req CreateTeamRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if req.Name == "" {
			http.Error(w, "Team name is required", http.StatusBadRequest)
			return
		}
		t, err := s.store.CreateTeam(r.Context(), req.Name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(TeamResponse{
			ID:        int(t.ID),
			Name:      req.Name,
			CreatedAt: t.CreatedAt,
		})
	case http.MethodDelete:
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
		if err := s.store.DeleteTeam(r.Context(), int32(id)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleTeamMembers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		teamIDStr := r.URL.Query().Get("team_id")
		userIDStr := r.URL.Query().Get("user_id")

		if teamIDStr != "" {
			var teamID int
			if _, err := fmt.Sscanf(teamIDStr, "%d", &teamID); err != nil {
				http.Error(w, "Invalid team_id", http.StatusBadRequest)
				return
			}
			members, err := s.store.GetTeamMembers(r.Context(), int32(teamID))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			var response []TeamMemberResponse
			for _, tm := range members {
				response = append(response, TeamMemberResponse{
					ID:           int(tm.ID),
					TeamID:       int(tm.TeamID),
					UserID:       int(tm.UserID),
					Productivity: int(tm.Productivity.Int32),
					CreatedAt:    tm.CreatedAt,
					UserName:     tm.UserName,
					TeamName:     "",
				})
			}
			json.NewEncoder(w).Encode(response)
		} else if userIDStr != "" {
			var userID int
			if _, err := fmt.Sscanf(userIDStr, "%d", &userID); err != nil {
				http.Error(w, "Invalid user_id", http.StatusBadRequest)
				return
			}
			members, err := s.store.GetUserTeams(r.Context(), int32(userID))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			var response []TeamMemberResponse
			for _, tm := range members {
				response = append(response, TeamMemberResponse{
					ID:           int(tm.ID),
					TeamID:       int(tm.TeamID),
					UserID:       int(tm.UserID),
					Productivity: int(tm.Productivity.Int32),
					CreatedAt:    tm.CreatedAt,
					UserName:     "",
					TeamName:     tm.TeamName,
				})
			}
			json.NewEncoder(w).Encode(response)
		} else {
			http.Error(w, "Missing team_id or user_id", http.StatusBadRequest)
			return
		}
	case http.MethodPost:
		var req AddTeamMemberRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if req.Productivity < 0 || req.Productivity > 100 {
			http.Error(w, "Productivity must be between 0 and 100", http.StatusBadRequest)
			return
		}

		params := store.AddUserToTeamParams{
			TeamID:       int32(req.TeamID),
			UserID:       int32(req.UserID),
			Productivity: pgtype.Int4{Int32: int32(req.Productivity), Valid: true},
		}

		tm, err := s.store.AddUserToTeam(r.Context(), params)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(TeamMemberResponse{
			ID:           int(tm.ID),
			TeamID:       req.TeamID,
			UserID:       req.UserID,
			Productivity: req.Productivity,
			CreatedAt:    tm.CreatedAt,
		})
	case http.MethodDelete:
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

		params := store.RemoveUserFromTeamParams{
			TeamID: int32(teamID),
			UserID: int32(userID),
		}

		if err := s.store.RemoveUserFromTeam(r.Context(), params); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
