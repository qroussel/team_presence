package testutil

import (
	"fmt"
	"time"
)

// Note: We can't import store package here due to import cycles,
// so we define builder types that return generic maps/structs
// and let the calling code convert to store types.

// UserBuilder provides fluent API for creating test users
type UserBuilder struct {
	name      string
	email     string
	avatarURL string
}

// NewUserBuilder creates a new user builder with sensible defaults
func NewUserBuilder() *UserBuilder {
	return &UserBuilder{
		name:      "Test User " + RandomString(5),
		email:     fmt.Sprintf("test_%s@example.com", RandomString(5)),
		avatarURL: "http://example.com/avatar.jpg",
	}
}

// WithName sets the user name
func (b *UserBuilder) WithName(name string) *UserBuilder {
	b.name = name
	return b
}

// WithEmail sets the user email
func (b *UserBuilder) WithEmail(email string) *UserBuilder {
	b.email = email
	return b
}

// WithAvatarURL sets the avatar URL
func (b *UserBuilder) WithAvatarURL(url string) *UserBuilder {
	b.avatarURL = url
	return b
}

// Build returns the user data as a map (to avoid import cycles)
func (b *UserBuilder) Build() map[string]interface{} {
	return map[string]interface{}{
		"name":       b.name,
		"email":      b.email,
		"avatar_url": b.avatarURL,
	}
}

// Name returns the configured name
func (b *UserBuilder) Name() string {
	return b.name
}

// Email returns the configured email
func (b *UserBuilder) Email() string {
	return b.email
}

// AvatarURL returns the configured avatar URL
func (b *UserBuilder) AvatarURL() string {
	return b.avatarURL
}

// TeamBuilder provides fluent API for creating test teams
type TeamBuilder struct {
	name string
}

// NewTeamBuilder creates a new team builder with sensible defaults
func NewTeamBuilder() *TeamBuilder {
	return &TeamBuilder{
		name: "Test Team " + RandomString(8),
	}
}

// WithName sets the team name
func (b *TeamBuilder) WithName(name string) *TeamBuilder {
	b.name = name
	return b
}

// Name returns the configured team name
func (b *TeamBuilder) Name() string {
	return b.name
}

// Build returns the team data as a map
func (b *TeamBuilder) Build() map[string]interface{} {
	return map[string]interface{}{
		"name": b.name,
	}
}

// PresenceBuilder provides fluent API for creating test presence records
type PresenceBuilder struct {
	userID   int
	date     string
	statusAM string
	statusPM string
}

// NewPresenceBuilder creates a new presence builder with sensible defaults
func NewPresenceBuilder() *PresenceBuilder {
	return &PresenceBuilder{
		userID:   0, // Must be set by caller
		date:     time.Now().Format("2006-01-02"),
		statusAM: "office",
		statusPM: "office",
	}
}

// WithUserID sets the user ID
func (b *PresenceBuilder) WithUserID(userID int) *PresenceBuilder {
	b.userID = userID
	return b
}

// WithDate sets the date
func (b *PresenceBuilder) WithDate(date string) *PresenceBuilder {
	b.date = date
	return b
}

// WithStatusAM sets the AM status
func (b *PresenceBuilder) WithStatusAM(status string) *PresenceBuilder {
	b.statusAM = status
	return b
}

// WithStatusPM sets the PM status
func (b *PresenceBuilder) WithStatusPM(status string) *PresenceBuilder {
	b.statusPM = status
	return b
}

// Build returns the presence data as a map
func (b *PresenceBuilder) Build() map[string]interface{} {
	return map[string]interface{}{
		"user_id":   b.userID,
		"date":      b.date,
		"status_am": b.statusAM,
		"status_pm": b.statusPM,
	}
}

// UserID returns the configured user ID
func (b *PresenceBuilder) UserID() int {
	return b.userID
}

// Date returns the configured date
func (b *PresenceBuilder) Date() string {
	return b.date
}

// StatusAM returns the configured AM status
func (b *PresenceBuilder) StatusAM() string {
	return b.statusAM
}

// StatusPM returns the configured PM status
func (b *PresenceBuilder) StatusPM() string {
	return b.statusPM
}

// TeamMemberBuilder provides fluent API for creating test team members
type TeamMemberBuilder struct {
	teamID       int
	userID       int
	productivity int
}

// NewTeamMemberBuilder creates a new team member builder with sensible defaults
func NewTeamMemberBuilder() *TeamMemberBuilder {
	return &TeamMemberBuilder{
		teamID:       0, // Must be set by caller
		userID:       0, // Must be set by caller
		productivity: 100,
	}
}

// WithTeamID sets the team ID
func (b *TeamMemberBuilder) WithTeamID(teamID int) *TeamMemberBuilder {
	b.teamID = teamID
	return b
}

// WithUserID sets the user ID
func (b *TeamMemberBuilder) WithUserID(userID int) *TeamMemberBuilder {
	b.userID = userID
	return b
}

// WithProductivity sets the productivity percentage
func (b *TeamMemberBuilder) WithProductivity(productivity int) *TeamMemberBuilder {
	b.productivity = productivity
	return b
}

// Build returns the team member data as a map
func (b *TeamMemberBuilder) Build() map[string]interface{} {
	return map[string]interface{}{
		"team_id":      b.teamID,
		"user_id":      b.userID,
		"productivity": b.productivity,
	}
}

// TeamID returns the configured team ID
func (b *TeamMemberBuilder) TeamID() int {
	return b.teamID
}

// UserID returns the configured user ID
func (b *TeamMemberBuilder) UserID() int {
	return b.userID
}

// Productivity returns the configured productivity
func (b *TeamMemberBuilder) Productivity() int {
	return b.productivity
}
