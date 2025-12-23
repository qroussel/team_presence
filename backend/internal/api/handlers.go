package api

import (
	"backend/internal/store"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

const (
	PathHealth      = "/api/health"
	PathUsers       = "/api/users"
	PathTeams       = "/api/teams"
	PathPresence    = "/api/presence"
	PathTeamMembers = "/api/team_members"

	MsgMethodNotAllowed    = "Method not allowed"
	MsgMissingID           = "Missing id"
	MsgInvalidID           = "Invalid id"
	MsgInvalidTeamID       = "Invalid team_id"
	MsgInvalidUserID       = "Invalid user_id"
	MsgMissingTeamOrUserID = "Missing team_id or user_id"
	DateFormat             = "2006-01-02"
)

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (s *Server) handleUsers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleGetUsers(w, r)
	case http.MethodPost:
		s.handleCreateUser(w, r)
	case http.MethodDelete:
		s.handleDeleteUser(w, r)
	default:
		http.Error(w, MsgMethodNotAllowed, http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleGetUsers(w http.ResponseWriter, r *http.Request) {
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
			Role:      u.Role,
			CreatedAt: u.CreatedAt,
		})
	}
	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	authUser := GetUserFromContext(r.Context())
	if authUser == nil || authUser.Role != "Admin" {
		http.Error(w, "Forbidden: Admins only", http.StatusForbidden)
		return
	}

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
}

func (s *Server) handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	authUser := GetUserFromContext(r.Context())
	if authUser == nil || authUser.Role != "Admin" {
		http.Error(w, "Forbidden: Admins only", http.StatusForbidden)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, MsgMissingID, http.StatusBadRequest)
		return
	}
	var id int
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		http.Error(w, MsgInvalidID, http.StatusBadRequest)
		return
	}
	if err := s.store.DeleteUser(r.Context(), int32(id)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handlePresence(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleGetPresence(w, r)
	case http.MethodPost:
		s.handleUpsertPresence(w, r)
	default:
		http.Error(w, MsgMethodNotAllowed, http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleGetPresence(w http.ResponseWriter, r *http.Request) {
	start := r.URL.Query().Get("start")
	end := r.URL.Query().Get("end")
	if start == "" || end == "" {
		http.Error(w, "Missing start or end date", http.StatusBadRequest)
		return
	}

	startDate, err := time.Parse(DateFormat, start)
	if err != nil {
		http.Error(w, "Invalid start date", http.StatusBadRequest)
		return
	}
	endDate, err := time.Parse(DateFormat, end)
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
}

// checkPresenceAuthorization determines if the authenticated user can modify presence for targetUserID.
// Rules: Admin, Self, or Owner of a team the target user is in.
func (s *Server) checkPresenceAuthorization(ctx context.Context, authUser *AuthUser, targetUserID int32) bool {
	if authUser.Role == "Admin" || authUser.ID == targetUserID {
		return true
	}

	// Check if authUser owns any team the target user is in
	userTeams, err := s.store.GetUserTeams(ctx, targetUserID)
	if err != nil {
		return false
	}

	for _, t := range userTeams {
		members, err := s.store.GetTeamMembers(ctx, t.TeamID)
		if err == nil {
			for _, m := range members {
				if m.UserID == authUser.ID && m.Role == "Owner" {
					return true
				}
			}
		}
	}
	return false
}

func (s *Server) handleUpsertPresence(w http.ResponseWriter, r *http.Request) {
	authUser := GetUserFromContext(r.Context())
	if authUser == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req UpsertPresenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !s.checkPresenceAuthorization(r.Context(), authUser, int32(req.UserID)) {
		http.Error(w, "Forbidden: You cannot modify presence for this user", http.StatusForbidden)
		return
	}

	date, err := time.Parse(DateFormat, req.Date)
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
}

func (s *Server) handleTeams(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleGetTeams(w, r)
	case http.MethodPost:
		s.handleCreateTeam(w, r)
	case http.MethodDelete:
		s.handleDeleteTeam(w, r)
	case http.MethodPut:
		http.Error(w, "Use /api/team_members for ownership changes", http.StatusBadRequest)
	default:
		http.Error(w, MsgMethodNotAllowed, http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleGetTeams(w http.ResponseWriter, r *http.Request) {
	authUser := GetUserFromContext(r.Context())
	// Determine if user is authorized to see all teams or just their own
	// If unauthenticated, they see nothing (or unauthorized)
	if authUser == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var response []TeamResponse

	if authUser.Role == "Admin" {
		teams, err := s.store.GetTeams(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for _, t := range teams {
			response = append(response, TeamResponse{
				ID:        int(t.ID),
				Name:      t.Name,
				CreatedAt: t.CreatedAt,
				Role:      "Admin", // Admins implicitly own everything
			})
		}
	} else {
		teams, err := s.store.GetTeamsForUser(r.Context(), authUser.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for _, t := range teams {
			response = append(response, TeamResponse{
				ID:        int(t.ID),
				Name:      t.Name,
				CreatedAt: t.CreatedAt,
				Role:      t.Role,
			})
		}
	}

	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleCreateTeam(w http.ResponseWriter, r *http.Request) {
	authUser := GetUserFromContext(r.Context())
	if authUser == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		http.Error(w, "Team name is required", http.StatusBadRequest)
		return
	}

	params := store.CreateTeamParams{
		Name:    req.Name,
		OwnerID: pgtype.Int4{Int32: authUser.ID, Valid: true},
	}
	t, err := s.store.CreateTeam(r.Context(), params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	addParams := store.AddUserToTeamParams{
		TeamID:       t.ID,
		UserID:       authUser.ID,
		Productivity: pgtype.Int4{Int32: 100, Valid: true},
		Role:         "Owner",
	}
	_, err = s.store.AddUserToTeam(r.Context(), addParams)
	if err != nil {
		s.store.DeleteTeam(r.Context(), t.ID)
		http.Error(w, "Failed to assign owner: "+err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(TeamResponse{
		ID:        int(t.ID),
		Name:      req.Name,
		CreatedAt: t.CreatedAt,
	})
}

// checkAdminOrTeamOwner verifies if the user is an Admin or an Owner of the specific team.
func (s *Server) checkAdminOrTeamOwner(ctx context.Context, authUser *AuthUser, teamID int32) bool {
	if authUser.Role == "Admin" {
		return true
	}
	members, err := s.store.GetTeamMembers(ctx, teamID)
	if err != nil {
		return false
	}
	for _, m := range members {
		if m.UserID == authUser.ID && m.Role == "Owner" {
			return true
		}
	}
	return false
}

func (s *Server) handleDeleteTeam(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, MsgMissingID, http.StatusBadRequest)
		return
	}
	var id int
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		http.Error(w, MsgInvalidID, http.StatusBadRequest)
		return
	}

	authUser := GetUserFromContext(r.Context())
	if authUser == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if !s.checkAdminOrTeamOwner(r.Context(), authUser, int32(id)) {
		http.Error(w, "Forbidden: Only Admin or Team Owner can delete this team", http.StatusForbidden)
		return
	}

	if err := s.store.DeleteTeam(r.Context(), int32(id)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleTeamMembers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleGetTeamMembers(w, r)
	case http.MethodPost:
		s.handleAddTeamMember(w, r)
	case http.MethodPut:
		s.handleUpdateTeamMember(w, r)
	case http.MethodDelete:
		s.handleRemoveTeamMember(w, r)
	default:
		http.Error(w, MsgMethodNotAllowed, http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleGetTeamMembers(w http.ResponseWriter, r *http.Request) {
	teamIDStr := r.URL.Query().Get("team_id")
	userIDStr := r.URL.Query().Get("user_id")

	if teamIDStr != "" {
		s.renderTeamMembersByTeamID(w, r, teamIDStr)
	} else if userIDStr != "" {
		s.renderTeamMembersByUserID(w, r, userIDStr)
	} else {
		http.Error(w, MsgMissingTeamOrUserID, http.StatusBadRequest)
	}
}

func (s *Server) renderTeamMembersByTeamID(w http.ResponseWriter, r *http.Request, teamIDStr string) {
	var teamID int
	if _, err := fmt.Sscanf(teamIDStr, "%d", &teamID); err != nil {
		http.Error(w, MsgInvalidTeamID, http.StatusBadRequest)
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
			Role:         tm.Role,
			CreatedAt:    tm.CreatedAt,
			UserName:     tm.UserName,
		})
	}
	json.NewEncoder(w).Encode(response)
}

func (s *Server) renderTeamMembersByUserID(w http.ResponseWriter, r *http.Request, userIDStr string) {
	var userID int
	if _, err := fmt.Sscanf(userIDStr, "%d", &userID); err != nil {
		http.Error(w, MsgInvalidUserID, http.StatusBadRequest)
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
			Role:         tm.Role,
			CreatedAt:    tm.CreatedAt,
			TeamName:     tm.TeamName,
		})
	}
	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleAddTeamMember(w http.ResponseWriter, r *http.Request) {
	authUser := GetUserFromContext(r.Context())
	if authUser == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req AddTeamMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.Productivity < 0 || req.Productivity > 100 {
		http.Error(w, "Productivity must be between 0 and 100", http.StatusBadRequest)
		return
	}

	if !s.checkAdminOrTeamOwner(r.Context(), authUser, int32(req.TeamID)) {
		http.Error(w, "Forbidden: Only Admin or Team Owner can add members", http.StatusForbidden)
		return
	}

	params := store.AddUserToTeamParams{
		TeamID:       int32(req.TeamID),
		UserID:       int32(req.UserID),
		Productivity: pgtype.Int4{Int32: int32(req.Productivity), Valid: true},
		Role:         "Member",
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
		Role:         "Member",
		CreatedAt:    tm.CreatedAt,
	})
}

// parseTeamAndUserIDs Helper to extract query params
func parseTeamAndUserIDs(r *http.Request) (int32, int32, error) {
	teamIDStr := r.URL.Query().Get("team_id")
	userIDStr := r.URL.Query().Get("user_id")
	if teamIDStr == "" || userIDStr == "" {
		return 0, 0, errors.New(MsgMissingTeamOrUserID)
	}

	var teamID, userID int
	if _, err := fmt.Sscanf(teamIDStr, "%d", &teamID); err != nil {
		return 0, 0, errors.New(MsgInvalidTeamID)
	}
	if _, err := fmt.Sscanf(userIDStr, "%d", &userID); err != nil {
		return 0, 0, errors.New(MsgInvalidUserID)
	}
	return int32(teamID), int32(userID), nil
}

func (s *Server) handleUpdateTeamMember(w http.ResponseWriter, r *http.Request) {
	teamID, userID, err := parseTeamAndUserIDs(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	authUser := GetUserFromContext(r.Context())
	if authUser == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req UpdateTeamMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !s.checkAdminOrTeamOwner(r.Context(), authUser, teamID) {
		http.Error(w, "Forbidden: Only Admin or Team Owner can update members", http.StatusForbidden)
		return
	}

	if req.Role != "" {
		if req.Role != "Owner" && req.Role != "Member" {
			http.Error(w, "Invalid role. Must be 'Owner' or 'Member'", http.StatusBadRequest)
			return
		}
		params := store.UpdateTeamMemberRoleParams{
			TeamID: teamID,
			UserID: userID,
			Role:   req.Role,
		}
		if err := s.store.UpdateTeamMemberRole(r.Context(), params); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleRemoveTeamMember(w http.ResponseWriter, r *http.Request) {
	teamID, userID, err := parseTeamAndUserIDs(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	authUser := GetUserFromContext(r.Context())
	if authUser == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	isAuthorized := s.checkAdminOrTeamOwner(r.Context(), authUser, teamID)
	if authUser.ID == userID {
		isAuthorized = true
	}

	if !isAuthorized {
		http.Error(w, "Forbidden: Only Admin, Team Owner, or User themselves can remove member", http.StatusForbidden)
		return
	}

	params := store.RemoveUserFromTeamParams{
		TeamID: teamID,
		UserID: userID,
	}

	if err := s.store.RemoveUserFromTeam(r.Context(), params); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
