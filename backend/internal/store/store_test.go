package store

import (
	"backend/internal/testutil"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestStoreWithTx creates a new store with transaction-based isolation
// The transaction will be rolled back automatically after the test completes
func newTestStoreWithTx(t *testing.T) (*Store, context.Context) {
	t.Helper()

	s, err := NewStore(context.Background(), testutil.GetTestDatabaseURL())
	require.NoError(t, err, "Failed to create test store")

	// Begin a transaction for test isolation
	tx, err := s.BeginTx(context.Background())
	require.NoError(t, err, "Failed to begin transaction")

	// Create a context with the transaction
	ctx := context.WithValue(context.Background(), TxContextKey, tx)

	// Cleanup: rollback transaction and close store
	t.Cleanup(func() {
		tx.Rollback(context.Background())
		s.Close()
	})

	return s, ctx
}

func TestUserOperations(t *testing.T) {
	t.Parallel() // Enable parallel execution

	s, ctx := newTestStoreWithTx(t)

	// Create user using builder
	builder := testutil.NewUserBuilder()
	params := CreateUserParams{
		Name:      builder.Name(),
		Email:     builder.Email(),
		AvatarUrl: pgtype.Text{String: builder.AvatarURL(), Valid: builder.AvatarURL() != ""},
	}
	createdUser, err := s.CreateUser(ctx, params)
	require.NoError(t, err)
	assert.NotZero(t, createdUser.ID)
	// Check fields returned or check by querying
	// CreatedUserRow only has ID and CreatedAt. We can't check Name/Email from it directly.

	t.Run("GetUsers includes created user", func(t *testing.T) {
		users, err := s.GetUsers(ctx)
		require.NoError(t, err)

		found, ok := testutil.FindInSlice(users, func(u GetUsersRow) bool {
			return u.ID == createdUser.ID
		})
		require.True(t, ok, "Created user should be found in GetUsers")
		assert.Equal(t, params.Name, found.Name)
		assert.Equal(t, params.Email, found.Email)
		assert.Equal(t, params.AvatarUrl.String, found.AvatarUrl) // GetUsersRow returns string
	})

	t.Run("DeleteUser removes user", func(t *testing.T) {
		err := s.DeleteUser(ctx, createdUser.ID)
		require.NoError(t, err)

		usersAfter, err := s.GetUsers(ctx)
		require.NoError(t, err)

		foundAfter := testutil.ContainsInSlice(usersAfter, func(u GetUsersRow) bool {
			return u.ID == createdUser.ID
		})
		assert.False(t, foundAfter, "Deleted user should not be found")
	})
}

func TestTeamOperations(t *testing.T) {
	t.Parallel()

	s, ctx := newTestStoreWithTx(t)

	// Create team using builder
	builder := testutil.NewTeamBuilder()
	createdTeam, err := s.CreateTeam(ctx, builder.Name())
	require.NoError(t, err)
	assert.NotZero(t, createdTeam.ID)

	t.Run("GetTeams includes created team", func(t *testing.T) {
		teams, err := s.GetTeams(ctx)
		require.NoError(t, err)

		found, ok := testutil.FindInSlice(teams, func(team Team) bool {
			return team.ID == createdTeam.ID
		})
		require.True(t, ok, "Created team should be found in GetTeams")
		assert.Equal(t, builder.Name(), found.Name)
	})

	t.Run("DeleteTeam removes team", func(t *testing.T) {
		err := s.DeleteTeam(ctx, createdTeam.ID)
		require.NoError(t, err)

		teamsAfter, err := s.GetTeams(ctx)
		require.NoError(t, err)

		foundAfter := testutil.ContainsInSlice(teamsAfter, func(team Team) bool {
			return team.ID == createdTeam.ID
		})
		assert.False(t, foundAfter, "Deleted team should not be found")
	})
}

func TestTeamMemberOperations(t *testing.T) {
	t.Parallel()

	s, ctx := newTestStoreWithTx(t)

	// Setup: Create user and team using builders
	userBuilder := testutil.NewUserBuilder()
	userParams := CreateUserParams{Name: userBuilder.Name(), Email: userBuilder.Email()}
	user, err := s.CreateUser(ctx, userParams)
	require.NoError(t, err)

	teamBuilder := testutil.NewTeamBuilder()
	team, err := s.CreateTeam(ctx, teamBuilder.Name())
	require.NoError(t, err)

	productivity := int32(85)

	t.Run("AddUserToTeam creates team member", func(t *testing.T) {
		params := AddUserToTeamParams{
			TeamID:       team.ID,
			UserID:       user.ID,
			Productivity: pgtype.Int4{Int32: productivity, Valid: true},
		}

		member, err := s.AddUserToTeam(ctx, params)
		require.NoError(t, err)
		assert.NotZero(t, member.ID)
	})

	t.Run("GetTeamMembers returns team members", func(t *testing.T) {
		members, err := s.GetTeamMembers(ctx, team.ID)
		require.NoError(t, err)
		require.Len(t, members, 1)
		assert.Equal(t, user.ID, members[0].UserID)
		assert.Equal(t, userBuilder.Name(), members[0].UserName)
		assert.Equal(t, productivity, members[0].Productivity.Int32)
	})

	t.Run("GetUserTeams returns user's teams", func(t *testing.T) {
		teams, err := s.GetUserTeams(ctx, user.ID)
		require.NoError(t, err)
		require.Len(t, teams, 1)
		assert.Equal(t, team.ID, teams[0].TeamID)
		assert.Equal(t, teamBuilder.Name(), teams[0].TeamName)
	})

	t.Run("AddUserToTeam updates productivity", func(t *testing.T) {
		updatedProd := int32(90)
		params := AddUserToTeamParams{
			TeamID:       team.ID,
			UserID:       user.ID,
			Productivity: pgtype.Int4{Int32: updatedProd, Valid: true},
		}
		member, err := s.AddUserToTeam(ctx, params) // Re-adding triggers ON CONFLICT UPDATE
		require.NoError(t, err)

		// Verify
		members, _ := s.GetTeamMembers(ctx, team.ID)
		require.Equal(t, updatedProd, members[0].Productivity.Int32)
		_ = member
	})

	t.Run("RemoveUserFromTeam removes member", func(t *testing.T) {
		params := RemoveUserFromTeamParams{
			TeamID: team.ID,
			UserID: user.ID,
		}
		err := s.RemoveUserFromTeam(ctx, params)
		require.NoError(t, err)

		membersAfter, err := s.GetTeamMembers(ctx, team.ID)
		require.NoError(t, err)
		assert.Empty(t, membersAfter)

		userTeamsAfter, err := s.GetUserTeams(ctx, user.ID)
		require.NoError(t, err)
		assert.Empty(t, userTeamsAfter)
	})
}

func TestPresenceOperations(t *testing.T) {
	s, ctx := newTestStoreWithTx(t)

	// Setup: Create User
	uName := "PresenceUser " + testutil.RandomString(5)
	uEmail := fmt.Sprintf("presence_%s@example.com", testutil.RandomString(5))
	user, err := s.CreateUser(ctx, CreateUserParams{Name: uName, Email: uEmail})
	require.NoError(t, err)

	dateStr := time.Now().Format("2006-01-02")
	date, _ := time.Parse("2006-01-02", dateStr)
	statusAM := "office"
	statusPM := "remote"

	// Upsert Presence
	params := UpsertPresenceParams{
		UserID:   user.ID,
		Date:     date,
		StatusAm: pgtype.Text{String: statusAM, Valid: true},
		StatusPm: pgtype.Text{String: statusPM, Valid: true},
	}
	createdP, err := s.UpsertPresence(ctx, params)
	require.NoError(t, err)
	assert.NotZero(t, createdP.ID)
	// CreateRow only returned ID/CreatedAt.

	// Get Presence
	start := date
	end := date
	entries, err := s.GetPresence(ctx, GetPresenceParams{Column1: start, Column2: end})
	require.NoError(t, err)
	require.NotEmpty(t, entries)

	found := false
	for _, e := range entries {
		if e.UserID == user.ID && e.Date == dateStr { // Generated GetPresenceRow.Date is string
			found = true
			assert.Equal(t, statusAM, e.StatusAm)
			assert.Equal(t, statusPM, e.StatusPm)
			break
		}
	}
	assert.True(t, found, "Presence should be found")

	// Upsert Update
	newStatusAM := "off"
	params.StatusAm = pgtype.Text{String: newStatusAM, Valid: true}
	_, err = s.UpsertPresence(ctx, params)
	require.NoError(t, err)

	// Verify Update
	entriesUpdated, err := s.GetPresence(ctx, GetPresenceParams{Column1: start, Column2: end})
	require.NoError(t, err)
	for _, e := range entriesUpdated {
		if e.UserID == user.ID && e.Date == dateStr {
			assert.Equal(t, newStatusAM, e.StatusAm)
			break
		}
	}
}
