package api

import "time"

// UserRequest is the payload for creating a user
type CreateUserRequest struct {
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

// UserResponse is the response for user operations
type UserResponse struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	AvatarURL string    `json:"avatar_url"`
	CreatedAt time.Time `json:"created_at"`
}

// PresenceRequest is the payload for upserting presence
type UpsertPresenceRequest struct {
	UserID   int    `json:"user_id"`
	Date     string `json:"date"` // YYYY-MM-DD
	StatusAM string `json:"status_am"`
	StatusPM string `json:"status_pm"`
}

// PresenceResponse is the response for presence operations
type PresenceResponse struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Date      string    `json:"date"`
	StatusAM  string    `json:"status_am"`
	StatusPM  string    `json:"status_pm"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateTeamRequest is the payload for creating a team
type CreateTeamRequest struct {
	Name string `json:"name"`
}

// TeamResponse is the response for team operations
type TeamResponse struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// AddTeamMemberRequest is the payload for adding a user to a team
type AddTeamMemberRequest struct {
	TeamID       int `json:"team_id"`
	UserID       int `json:"user_id"`
	Productivity int `json:"productivity"`
}

// TeamMemberResponse is the response for team member operations
type TeamMemberResponse struct {
	ID           int       `json:"id"`
	TeamID       int       `json:"team_id"`
	UserID       int       `json:"user_id"`
	Productivity int       `json:"productivity"`
	CreatedAt    time.Time `json:"created_at"`
	UserName     string    `json:"user_name,omitempty"`
	TeamName     string    `json:"team_name,omitempty"`
}
