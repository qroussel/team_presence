package store

import (
	"backend/internal/testutil"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// VALIDATION TESTS
// ============================================================================

const (
	MsgValueTooLong         = "value too long"
	MsgForeignKeyConstraint = "foreign key constraint"
)

func TestCreateUserValidation(t *testing.T) {
	tests := []struct {
		name    string
		params  CreateUserParams
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid user with all fields",
			params: CreateUserParams{
				Name:      "John Doe",
				Email:     "john.doe@example.com",
				AvatarUrl: pgtype.Text{String: "http://example.com/avatar.jpg", Valid: true},
			},
			wantErr: false,
		},
		{
			name: "valid user with special characters in name",
			params: CreateUserParams{
				Name:  "José María O'Connor-Smith",
				Email: "jose@example.com",
			},
			wantErr: false,
		},
		{
			name: "valid user with unicode in name",
			params: CreateUserParams{
				Name:  "李明 (Li Ming)",
				Email: "liming@example.com",
			},
			wantErr: false,
		},
		{
			name: "empty name - accepted (NOT NULL allows empty string)",
			params: CreateUserParams{
				Name:  "",
				Email: "test@example.com",
			},
			wantErr: false,
		},
		{
			name: "empty email - accepted (NOT NULL allows empty string)",
			params: CreateUserParams{
				Name:  "Test User",
				Email: "",
			},
			wantErr: false,
		},
		{
			name: "very long name (500 chars) - should fail",
			params: CreateUserParams{
				Name:  testutil.RandomString(500),
				Email: "longname@example.com",
			},
			wantErr: true,
			errMsg:  MsgValueTooLong,
		},
		{
			name: "very long email (500 chars) - should fail",
			params: CreateUserParams{
				Name:  "Test",
				Email: testutil.RandomString(500) + "@example.com",
			},
			wantErr: true,
			errMsg:  "value too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, ctx := newTestStoreWithTx(t)

			// Make email unique for each test iteration
			if tt.params.Email != "" {
				tt.params.Email = fmt.Sprintf("%s_%s", tt.params.Email, testutil.RandomString(5))
			}

			_, err := s.CreateUser(ctx, tt.params)

			if tt.wantErr {
				require.Error(t, err, "Expected error for: %s", tt.name)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err, "Unexpected error for: %s", tt.name)
			}
		})
	}
}

func TestCreateUserUniqueEmail(t *testing.T) {
	s, ctx := newTestStoreWithTx(t)

	email := fmt.Sprintf("duplicate_%s@example.com", testutil.RandomString(5))

	// Create first user
	user1 := CreateUserParams{
		Name:  "User One",
		Email: email,
	}
	_, err := s.CreateUser(ctx, user1)
	require.NoError(t, err)

	// Try to create second user with same email
	user2 := CreateUserParams{
		Name:  "User Two",
		Email: email,
	}
	_, err = s.CreateUser(ctx, user2)
	require.Error(t, err, "Should fail due to unique constraint on email")
	assert.Contains(t, err.Error(), "duplicate key value")
}

func TestUserOperationsInvalidID(t *testing.T) {
	s, ctx := newTestStoreWithTx(t)

	tests := []struct {
		name        string
		userID      int32
		errExpected bool
	}{
		{
			name:        "non-existent user ID",
			userID:      999999,
			errExpected: false, // Delete of non-existent ID typically succeeds
		},
		{
			name:        "negative user ID",
			userID:      -1,
			errExpected: false,
		},
		{
			name:        "zero user ID",
			userID:      0,
			errExpected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name+" - DeleteUser", func(t *testing.T) {
			err := s.DeleteUser(ctx, tt.userID)
			if tt.errExpected {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})

		t.Run(tt.name+" - GetUserTeams", func(t *testing.T) {
			teams, err := s.GetUserTeams(ctx, tt.userID)
			require.NoError(t, err) // Should return empty list, not error
			assert.Empty(t, teams)
		})
	}
}

func TestCreateTeamValidation(t *testing.T) {
	tests := []struct {
		name     string
		teamName string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid team name",
			teamName: "Engineering Team",
			wantErr:  false,
		},
		{
			name:     "valid with special characters",
			teamName: "R&D - AI/ML Department",
			wantErr:  false,
		},
		{
			name:     "valid with unicode",
			teamName: "開発チーム (Dev Team)",
			wantErr:  false,
		},
		{
			name:     "empty team name - accepted (NOT NULL allows empty)",
			teamName: "",
			wantErr:  false,
		},
		{
			name:     "very long team name - should fail",
			teamName: testutil.RandomString(500),
			wantErr:  true,
			errMsg:   MsgValueTooLong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, ctx := newTestStoreWithTx(t)

			teamName := tt.teamName
			if teamName != "" {
				// Make unique to avoid conflicts
				teamName = teamName + " " + testutil.RandomString(5)
			}

			_, err := s.CreateTeam(ctx, CreateTeamParams{Name: teamName})

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCreateTeamUniqueName(t *testing.T) {
	s, ctx := newTestStoreWithTx(t)

	teamName := "Duplicate Team " + testutil.RandomString(5)

	// Create first team
	_, err := s.CreateTeam(ctx, CreateTeamParams{Name: teamName})
	require.NoError(t, err)

	// Try to create second team with same name
	_, err = s.CreateTeam(ctx, CreateTeamParams{Name: teamName})
	require.Error(t, err, "Should fail due to unique constraint on team name")
	assert.Contains(t, err.Error(), "duplicate key value")
}

func TestTeamOperationsInvalidID(t *testing.T) {
	s, ctx := newTestStoreWithTx(t)

	tests := []struct {
		name   string
		teamID int32
	}{
		{"non-existent team ID", 999999},
		{"negative team ID", -1},
		{"zero team ID", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name+" - DeleteTeam", func(t *testing.T) {
			err := s.DeleteTeam(ctx, tt.teamID)
			require.NoError(t, err) // Postgres allows delete of non-existent rows
		})

		t.Run(tt.name+" - GetTeamMembers", func(t *testing.T) {
			members, err := s.GetTeamMembers(ctx, tt.teamID)
			require.NoError(t, err)
			assert.Empty(t, members)
		})
	}
}

func TestAddUserToTeamValidationSuccess(t *testing.T) {
	tests := []struct {
		name         string
		productivity int32
	}{
		{name: "valid productivity 100", productivity: 100},
		{name: "valid productivity 0", productivity: 0},
		{name: "valid productivity 50", productivity: 50},
		{name: "negative productivity", productivity: -10}, // Assuming store allows it or check constraint isn't active/tested here
		{name: "productivity over 100", productivity: 150},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, ctx := newTestStoreWithTx(t)
			validUser, err := s.CreateUser(ctx, CreateUserParams{
				Name:  "Valid User",
				Email: fmt.Sprintf("valid_%s@example.com", testutil.RandomString(5)),
			})
			require.NoError(t, err)

			validTeam, err := s.CreateTeam(ctx, CreateTeamParams{Name: "Valid Team " + testutil.RandomString(5)})
			require.NoError(t, err)

			params := AddUserToTeamParams{
				TeamID:       validTeam.ID,
				UserID:       validUser.ID,
				Productivity: pgtype.Int4{Int32: tt.productivity, Valid: true},
			}
			_, err = s.AddUserToTeam(ctx, params)
			require.NoError(t, err)
		})
	}
}

func TestAddUserToTeamValidationFailure(t *testing.T) {
	tests := []struct {
		name   string
		teamID int32
		userID int32
		errMsg string
	}{
		{
			name:   "non-existent team",
			teamID: 999999,
			userID: 0, // Will use valid user ID
			errMsg: MsgForeignKeyConstraint,
		},
		{
			name:   "non-existent user",
			teamID: 0, // Will use valid team ID
			userID: 999999,
			errMsg: MsgForeignKeyConstraint,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, ctx := newTestStoreWithTx(t)

			var userID, teamID int32
			if tt.userID == 0 {
				user, err := s.CreateUser(ctx, CreateUserParams{
					Name:  "Valid User",
					Email: fmt.Sprintf("valid_%s@example.com", testutil.RandomString(5)),
				})
				require.NoError(t, err)
				userID = user.ID
			} else {
				userID = tt.userID
			}

			if tt.teamID == 0 {
				team, err := s.CreateTeam(ctx, CreateTeamParams{Name: "Valid Team " + testutil.RandomString(5)})
				require.NoError(t, err)
				teamID = team.ID
			} else {
				teamID = tt.teamID
			}

			params := AddUserToTeamParams{
				TeamID:       teamID,
				UserID:       userID,
				Productivity: pgtype.Int4{Int32: 100, Valid: true},
			}

			_, err := s.AddUserToTeam(ctx, params)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errMsg)
		})
	}
}

func TestUpsertPresenceValidationSuccess(t *testing.T) {
	tests := []struct {
		name    string
		params  UpsertPresenceParams
		dateStr string
	}{
		{
			name: "valid presence",
			params: UpsertPresenceParams{
				StatusAm: pgtype.Text{String: "office", Valid: true},
				StatusPm: pgtype.Text{String: "remote", Valid: true},
			},
			dateStr: "2024-01-15",
		},
		{
			name: "valid with empty status",
			params: UpsertPresenceParams{
				StatusAm: pgtype.Text{String: "", Valid: true},
				StatusPm: pgtype.Text{String: "", Valid: true},
			},
			dateStr: "2024-01-16",
		},
		{
			name:    "invalid date format fallback",
			params:  UpsertPresenceParams{}, // Defaults
			dateStr: "2024-01-17",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, ctx := newTestStoreWithTx(t)
			user, err := s.CreateUser(ctx, CreateUserParams{
				Name:  "Presence User",
				Email: fmt.Sprintf("presence_%s@example.com", testutil.RandomString(5)),
			})
			require.NoError(t, err)

			p := tt.params
			p.UserID = user.ID
			if tt.dateStr != "" {
				d, err := time.Parse("2006-01-02", tt.dateStr)
				if err == nil {
					p.Date = d
				}
			}
			_, err = s.UpsertPresence(ctx, p)
			require.NoError(t, err)
		})
	}
}

func TestUpsertPresenceValidationFailure(t *testing.T) {
	tests := []struct {
		name   string
		params UpsertPresenceParams
		errMsg string
	}{
		{
			name: "non-existent user",
			params: UpsertPresenceParams{
				UserID:   999999,
				StatusAm: pgtype.Text{String: "office", Valid: true},
				StatusPm: pgtype.Text{String: "remote", Valid: true},
			},
			errMsg: MsgForeignKeyConstraint,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, ctx := newTestStoreWithTx(t)
			p := tt.params
			// We need a date, even for failure usually
			p.Date = time.Now()

			_, err := s.UpsertPresence(ctx, p)
			require.Error(t, err)
			if tt.errMsg != "" {
				assert.Contains(t, err.Error(), tt.errMsg)
			}
		})
	}
}
