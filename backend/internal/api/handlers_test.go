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

	"github.com/jackc/pgx/v5/pgtype"
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
// Helper to get context with Admin user
func getAdminContext(ctx context.Context, userID int32) context.Context {
	authUser := AuthUser{
		ID:   userID,
		Name: "Admin User",
		Role: "Admin",
	}
	return context.WithValue(ctx, UserContextKey, authUser)
}

func executeRequest(server *Server, req *http.Request, ctx context.Context) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
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

	adminCtx := getAdminContext(ctx, 1) // ID 1 doesn't need to exist for CreateUser, so this is fine.

	req, _ := http.NewRequest("POST", PathUsers, bytes.NewBuffer(body))
	rr := executeRequest(server, req, adminCtx)

	require.Equal(t, http.StatusOK, rr.Code)
	var createdUser UserResponse
	err := json.Unmarshal(rr.Body.Bytes(), &createdUser)
	require.NoError(t, err)
	assert.Equal(t, name, createdUser.Name)
	assert.NotZero(t, createdUser.ID)

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
	rr = executeRequest(server, req, adminCtx)
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
	// Need admin for creating user
	adminCtx := getAdminContext(ctx, 1)
	setupRR := executeRequest(server, setupReq, adminCtx)
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
	// Upsert Presence requires Admin or Self. Let's use Admin.
	// NOTE: upsert presence checks if user exists in various ways but mostly it just works if user ID is valid.
	// However, we are passing ID 1 as admin. If we stick to Admin check, we don't strictly require ID 1 to exist unless we check team logic.
	// But our newly Refactored `handleUpsertPresence` DOES check team logic if not Admin/Self.
	// Since we are Admin, it returns true immediately in checkPresenceAuthorization.
	rr := executeRequest(server, req, adminCtx)
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

	// Create real Admin User in DB
	adminUser, err := s.CreateUser(ctx, store.CreateUserParams{
		Name:      "Admin User Teams",
		Email:     "admin_teams_" + testutil.RandomString(5) + "@example.com",
		AvatarUrl: pgtype.Text{String: "http://example.com/avatar", Valid: true},
		Role:      "Admin",
	})
	require.NoError(t, err)

	teamName := "API Team " + testutil.RandomString(8)

	// POST /api/teams - Create team
	newTeamReq := CreateTeamRequest{Name: teamName}
	body, _ := json.Marshal(newTeamReq)
	req, _ := http.NewRequest("POST", PathTeams, bytes.NewBuffer(body))
	// Need Admin/User for CreateTeam. Use Real Admin User ID.
	adminCtx := getAdminContext(ctx, adminUser.ID)
	rr := executeRequest(server, req, adminCtx)

	require.Equal(t, http.StatusOK, rr.Code)
	var createdTeam TeamResponse
	err = json.Unmarshal(rr.Body.Bytes(), &createdTeam)
	require.NoError(t, err)
	assert.Equal(t, teamName, createdTeam.Name)
	assert.NotZero(t, createdTeam.ID)

	// GET /api/teams - List teams
	req, _ = http.NewRequest("GET", PathTeams, nil)
	rr = executeRequest(server, req, adminCtx)

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
	rr = executeRequest(server, req, adminCtx)
	require.Equal(t, http.StatusOK, rr.Code)

	// Verify Delete
	req, _ = http.NewRequest("GET", PathTeams, nil)
	rr = executeRequest(server, req, adminCtx)
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

func TestTeamsEndpointFiltering(t *testing.T) {
	s, ctx := newTestStoreWithTx(t)
	server := NewServer(s)

	// Admin
	uAdmin, _ := s.CreateUser(ctx, store.CreateUserParams{Name: "Admin", Email: "admin@filt.com", Role: "Admin"})
	// User A
	uA, _ := s.CreateUser(ctx, store.CreateUserParams{Name: "User A", Email: "ua@filt.com", Role: "User"})
	// User B
	uB, _ := s.CreateUser(ctx, store.CreateUserParams{Name: "User B", Email: "ub@filt.com", Role: "User"})

	// Team 1: Owned by A
	t1, _ := s.CreateTeam(ctx, store.CreateTeamParams{Name: "Team A", OwnerID: pgtype.Int4{Int32: uA.ID, Valid: true}})
	s.AddUserToTeam(ctx, store.AddUserToTeamParams{TeamID: t1.ID, UserID: uA.ID, Role: "Owner"})

	// Team 2: Owned by B
	t2, _ := s.CreateTeam(ctx, store.CreateTeamParams{Name: "Team B", OwnerID: pgtype.Int4{Int32: uB.ID, Valid: true}})
	s.AddUserToTeam(ctx, store.AddUserToTeamParams{TeamID: t2.ID, UserID: uB.ID, Role: "Owner"})

	// Helper to get teams
	getTeams := func(userID int32, role string) []TeamResponse {
		req, _ := http.NewRequest("GET", PathTeams, nil)
		// Mock auth context
		authUser := AuthUser{ID: userID, Role: role}
		reqCtx := context.WithValue(ctx, UserContextKey, authUser)
		rr := executeRequest(server, req, reqCtx)
		require.Equal(t, http.StatusOK, rr.Code)
		var teams []TeamResponse
		json.Unmarshal(rr.Body.Bytes(), &teams)
		return teams
	}

	// 1. Admin sees everything
	teamsAdmin := getTeams(uAdmin.ID, "Admin")
	assert.GreaterOrEqual(t, len(teamsAdmin), 2)
	// Check roles
	for _, tRes := range teamsAdmin {
		if tRes.ID == int(t1.ID) || tRes.ID == int(t2.ID) {
			assert.Equal(t, "Admin", tRes.Role)
		}
	}

	// 2. User A sees only Team A
	teamsA := getTeams(uA.ID, "User")
	assert.Len(t, teamsA, 1)
	assert.Equal(t, int(t1.ID), teamsA[0].ID)
	assert.Equal(t, "Owner", teamsA[0].Role)

	// 3. User B sees only Team B
	teamsB := getTeams(uB.ID, "User")
	assert.Len(t, teamsB, 1)
	assert.Equal(t, int(t2.ID), teamsB[0].ID)
	assert.Equal(t, "Owner", teamsB[0].Role)
}

func TestTeamMembersEndpoint(t *testing.T) {
	s, ctx := newTestStoreWithTx(t)
	server := NewServer(s)

	// Create real Admin User in DB
	adminUser, err := s.CreateUser(ctx, store.CreateUserParams{
		Name:      "Admin User Members",
		Email:     "admin_members_" + testutil.RandomString(5) + "@example.com",
		AvatarUrl: pgtype.Text{String: "http://example.com/avatar", Valid: true},
		Role:      "Admin",
	})
	require.NoError(t, err)
	adminCtx := getAdminContext(ctx, adminUser.ID)

	// Setup: Create User and Team via API
	userName := "Member " + testutil.RandomString(5)
	userEmail := fmt.Sprintf("member_api_%s@example.com", testutil.RandomString(5))

	setupUserReq := CreateUserRequest{Name: userName, Email: userEmail}
	body, _ := json.Marshal(setupUserReq)
	req, _ := http.NewRequest("POST", "/api/users", bytes.NewBuffer(body))
	rr := executeRequest(server, req, adminCtx)
	var user UserResponse
	json.Unmarshal(rr.Body.Bytes(), &user)

	teamName := "Team " + testutil.RandomString(8)
	setupTeamReq := CreateTeamRequest{Name: teamName}
	body, _ = json.Marshal(setupTeamReq)
	req, _ = http.NewRequest("POST", "/api/teams", bytes.NewBuffer(body))
	rr = executeRequest(server, req, adminCtx)
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
	rr = executeRequest(server, req, adminCtx)

	require.Equal(t, http.StatusOK, rr.Code)
	var createdMember TeamMemberResponse
	err = json.Unmarshal(rr.Body.Bytes(), &createdMember)
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
	require.Len(t, members, 2) // Should contain admin (Owner) and new member

	// GET /api/team_members?user_id=X - Get user teams
	req, _ = http.NewRequest("GET", fmt.Sprintf("/api/team_members?user_id=%d", user.ID), nil)
	rr = executeRequest(server, req, ctx)

	require.Equal(t, http.StatusOK, rr.Code)
	var userTeams []TeamMemberResponse
	err = json.Unmarshal(rr.Body.Bytes(), &userTeams)
	require.NoError(t, err)
	require.Len(t, userTeams, 1) // Only Team T (for user U)
	assert.Equal(t, team.ID, userTeams[0].TeamID)
	assert.Equal(t, teamName, userTeams[0].TeamName)
	assert.Equal(t, productivity, userTeams[0].Productivity)

	// DELETE /api/team_members - Remove user from team
	req, _ = http.NewRequest("DELETE", fmt.Sprintf("/api/team_members?team_id=%d&user_id=%d", team.ID, user.ID), nil)
	rr = executeRequest(server, req, adminCtx)
	require.Equal(t, http.StatusOK, rr.Code)

	// Verify Removal
	req, _ = http.NewRequest("GET", fmt.Sprintf("/api/team_members?team_id=%d", team.ID), nil)
	rr = executeRequest(server, req, ctx)
	var membersAfter []TeamMemberResponse
	err = json.Unmarshal(rr.Body.Bytes(), &membersAfter)
	require.NoError(t, err)

	// Since Admin is also a member (owner), the list is not empty, but user U should be gone.
	assert.Len(t, membersAfter, 1, "Team should have 1 member (Owner) after deletion")
	assert.Equal(t, adminUser.ID, int32(membersAfter[0].UserID))
}
