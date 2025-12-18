package api

import (
	"backend/internal/store"
	"backend/internal/testutil"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestStoreWithTx creates a new store with transaction-based isolation
// The transaction will be rolled back automatically after the test completes
func newTestStoreWithTx(t *testing.T) (*store.Store, context.Context) {
	t.Helper()

	s, err := store.NewStore(context.Background(), testutil.GetTestDatabaseURL())
	require.NoError(t, err)

	// Begin a transaction for test isolation
	tx, err := s.BeginTx(context.Background())
	require.NoError(t, err)

	// Create a context with the transaction
	ctx := store.WithTx(context.Background(), tx)

	t.Cleanup(func() {
		tx.Rollback(context.Background())
		s.Close()
	})

	return s, ctx
}

// withTxMiddleware injects the transaction from ctx into the request context
func withTxMiddleware(next http.Handler, ctx context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Helper to execute request with context
func executeRequest(server *Server, req *http.Request, ctx context.Context) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	// Inject the transaction context into the request
	// Note: ctx here comes from newTestStoreWithTx which HAS the tx.
	server.ServeHTTP(rr, req.WithContext(ctx))
	return rr
}

func TestHealthEndpoint(t *testing.T) {
	s, ctx := newTestStoreWithTx(t)
	server := NewServer(s)

	req, _ := http.NewRequest("GET", "/api/health", nil)
	rr := executeRequest(server, req, ctx)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "OK", rr.Body.String())
}

func TestUsersEndpoint(t *testing.T) {
	s, ctx := newTestStoreWithTx(t)
	server := NewServer(s)

	name := "API User " + testutil.RandomString(5)
	email := fmt.Sprintf("api_user_%s@example.com", testutil.RandomString(5))

	// POST /api/users
	newUserReq := CreateUserRequest{
		Name:      name,
		Email:     email,
		AvatarURL: "http://example.com/api_avatar.jpg",
	}
	body, _ := json.Marshal(newUserReq)
	req, _ := http.NewRequest("POST", PathUsers, bytes.NewBuffer(body))
	rr := executeRequest(server, req, ctx)

	require.Equal(t, http.StatusOK, rr.Code)
	var createdUser UserResponse
	err := json.Unmarshal(rr.Body.Bytes(), &createdUser)
	require.NoError(t, err)
	assert.Equal(t, name, createdUser.Name)
	assert.NotZero(t, createdUser.ID)

	// No manual cleanup needed! Transaction rollback handles it.

	// GET /api/users
	req, _ = http.NewRequest("GET", PathUsers, nil)
	rr = executeRequest(server, req, ctx)

	require.Equal(t, http.StatusOK, rr.Code)
	var users []UserResponse
	err = json.Unmarshal(rr.Body.Bytes(), &users)
	require.NoError(t, err)

	found := false
	for _, u := range users {
		if u.ID == createdUser.ID {
			found = true
			break
		}
	}
	assert.True(t, found)

	// DELETE /api/users
	req, _ = http.NewRequest("DELETE", fmt.Sprintf("/api/users?id=%d", createdUser.ID), nil)
	rr = executeRequest(server, req, ctx)
	require.Equal(t, http.StatusOK, rr.Code)

	// Verify Delete
	req, _ = http.NewRequest("GET", PathUsers, nil)
	rr = executeRequest(server, req, ctx)
	var usersAfter []UserResponse
	err = json.Unmarshal(rr.Body.Bytes(), &usersAfter)
	require.NoError(t, err)
	foundAfter := false
	for _, u := range usersAfter {
		if u.ID == createdUser.ID {
			foundAfter = true
			break
		}
	}
	assert.False(t, foundAfter)
}

func TestPresenceEndpoint(t *testing.T) {
	s, ctx := newTestStoreWithTx(t)
	server := NewServer(s)

	// Setup User using API
	name := "P User"
	email := fmt.Sprintf("p_api_%s@example.com", testutil.RandomString(5))

	setupUserReq := CreateUserRequest{Name: name, Email: email}
	setupBody, _ := json.Marshal(setupUserReq)
	setupReq, _ := http.NewRequest("POST", "/api/users", bytes.NewBuffer(setupBody))
	setupRR := executeRequest(server, setupReq, ctx)
	var user UserResponse
	json.Unmarshal(setupRR.Body.Bytes(), &user)

	date := time.Now().Format("2006-01-02")

	// POST /api/presence
	pReq := UpsertPresenceRequest{
		UserID:   user.ID,
		Date:     date,
		StatusAM: "office",
		StatusPM: "remote",
	}
	body, _ := json.Marshal(pReq)
	req, _ := http.NewRequest("POST", "/api/presence", bytes.NewBuffer(body))
	rr := executeRequest(server, req, ctx)
	require.Equal(t, http.StatusOK, rr.Code)

	var createdP PresenceResponse
	err := json.Unmarshal(rr.Body.Bytes(), &createdP)
	require.NoError(t, err)
	assert.Equal(t, pReq.StatusAM, createdP.StatusAM)

	// GET /api/presence
	req, _ = http.NewRequest("GET", fmt.Sprintf("/api/presence?start=%s&end=%s", date, date), nil)
	rr = executeRequest(server, req, ctx)
	require.Equal(t, http.StatusOK, rr.Code)

	var entries []PresenceResponse
	err = json.Unmarshal(rr.Body.Bytes(), &entries)
	require.NoError(t, err)
	require.NotEmpty(t, entries)

	found := false
	for _, entry := range entries {
		if entry.UserID == user.ID {
			assert.Equal(t, "office", entry.StatusAM)
			found = true
			break
		}
	}
	assert.True(t, found, "Should find presence entry for test user")
}

func TestTeamsEndpoint(t *testing.T) {
	s, ctx := newTestStoreWithTx(t)
	server := NewServer(s)

	teamName := "API Team " + testutil.RandomString(8)

	// POST /api/teams - Create team
	newTeamReq := CreateTeamRequest{Name: teamName}
	body, _ := json.Marshal(newTeamReq)
	req, _ := http.NewRequest("POST", PathTeams, bytes.NewBuffer(body))
	rr := executeRequest(server, req, ctx)

	require.Equal(t, http.StatusOK, rr.Code)
	var createdTeam TeamResponse
	err := json.Unmarshal(rr.Body.Bytes(), &createdTeam)
	require.NoError(t, err)
	assert.Equal(t, teamName, createdTeam.Name)
	assert.NotZero(t, createdTeam.ID)

	// No manual cleanup

	// GET /api/teams - List teams
	req, _ = http.NewRequest("GET", PathTeams, nil)
	rr = executeRequest(server, req, ctx)

	require.Equal(t, http.StatusOK, rr.Code)
	var teams []TeamResponse
	err = json.Unmarshal(rr.Body.Bytes(), &teams)
	require.NoError(t, err)

	found := false
	for _, team := range teams {
		if team.ID == createdTeam.ID {
			found = true
			assert.Equal(t, teamName, team.Name)
			break
		}
	}
	assert.True(t, found, "Created team should be found in GET /api/teams")

	// DELETE /api/teams - Delete team
	req, _ = http.NewRequest("DELETE", fmt.Sprintf("/api/teams?id=%d", createdTeam.ID), nil)
	rr = executeRequest(server, req, ctx)
	require.Equal(t, http.StatusOK, rr.Code)

	// Verify Delete
	req, _ = http.NewRequest("GET", PathTeams, nil)
	rr = executeRequest(server, req, ctx)
	var teamsAfter []TeamResponse
	err = json.Unmarshal(rr.Body.Bytes(), &teamsAfter)
	require.NoError(t, err)

	foundAfter := false
	for _, team := range teamsAfter {
		if team.ID == createdTeam.ID {
			foundAfter = true
			break
		}
	}
	assert.False(t, foundAfter, "Deleted team should not be found")
}

func TestTeamMembersEndpoint(t *testing.T) {
	s, ctx := newTestStoreWithTx(t)
	server := NewServer(s)

	// Setup: Create User and Team via API
	userName := "Member " + testutil.RandomString(5)
	userEmail := fmt.Sprintf("member_api_%s@example.com", testutil.RandomString(5))

	setupUserReq := CreateUserRequest{Name: userName, Email: userEmail}
	body, _ := json.Marshal(setupUserReq)
	req, _ := http.NewRequest("POST", "/api/users", bytes.NewBuffer(body))
	rr := executeRequest(server, req, ctx)
	var user UserResponse
	json.Unmarshal(rr.Body.Bytes(), &user)

	teamName := "Team " + testutil.RandomString(8)
	setupTeamReq := CreateTeamRequest{Name: teamName}
	body, _ = json.Marshal(setupTeamReq)
	req, _ = http.NewRequest("POST", "/api/teams", bytes.NewBuffer(body))
	rr = executeRequest(server, req, ctx)
	var team TeamResponse
	json.Unmarshal(rr.Body.Bytes(), &team)

	// POST /api/team_members - Add user to team
	productivity := 85
	newMemberReq := AddTeamMemberRequest{
		TeamID:       team.ID,
		UserID:       user.ID,
		Productivity: productivity,
	}
	body, _ = json.Marshal(newMemberReq)
	req, _ = http.NewRequest("POST", "/api/team_members", bytes.NewBuffer(body))
	rr = executeRequest(server, req, ctx)

	require.Equal(t, http.StatusOK, rr.Code)
	var createdMember TeamMemberResponse
	err := json.Unmarshal(rr.Body.Bytes(), &createdMember)
	require.NoError(t, err)
	assert.Equal(t, team.ID, createdMember.TeamID)
	assert.Equal(t, user.ID, createdMember.UserID)
	assert.Equal(t, productivity, createdMember.Productivity)

	// GET /api/team_members?team_id=X - Get team members
	req, _ = http.NewRequest("GET", fmt.Sprintf("/api/team_members?team_id=%d", team.ID), nil)
	rr = executeRequest(server, req, ctx)

	require.Equal(t, http.StatusOK, rr.Code)
	var members []TeamMemberResponse
	err = json.Unmarshal(rr.Body.Bytes(), &members)
	require.NoError(t, err)
	require.Len(t, members, 1)
	assert.Equal(t, user.ID, members[0].UserID)
	assert.Equal(t, userName, members[0].UserName)
	assert.Equal(t, productivity, members[0].Productivity)

	// GET /api/team_members?user_id=X - Get user teams
	req, _ = http.NewRequest("GET", fmt.Sprintf("/api/team_members?user_id=%d", user.ID), nil)
	rr = executeRequest(server, req, ctx)

	require.Equal(t, http.StatusOK, rr.Code)
	var userTeams []TeamMemberResponse
	err = json.Unmarshal(rr.Body.Bytes(), &userTeams)
	require.NoError(t, err)
	require.Len(t, userTeams, 1)
	assert.Equal(t, team.ID, userTeams[0].TeamID)
	assert.Equal(t, teamName, userTeams[0].TeamName)
	assert.Equal(t, productivity, userTeams[0].Productivity)

	// DELETE /api/team_members - Remove user from team
	req, _ = http.NewRequest("DELETE", fmt.Sprintf("/api/team_members?team_id=%d&user_id=%d", team.ID, user.ID), nil)
	rr = executeRequest(server, req, ctx)
	require.Equal(t, http.StatusOK, rr.Code)

	// Verify Removal
	req, _ = http.NewRequest("GET", fmt.Sprintf("/api/team_members?team_id=%d", team.ID), nil)
	rr = executeRequest(server, req, ctx)
	var membersAfter []TeamMemberResponse
	err = json.Unmarshal(rr.Body.Bytes(), &membersAfter)
	require.NoError(t, err)
	assert.Empty(t, membersAfter, "Team should have no members after deletion")
}
