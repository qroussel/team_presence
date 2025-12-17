package store

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Store struct {
	db *sql.DB
}

type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	AvatarURL string    `json:"avatar_url"`
	CreatedAt time.Time `json:"created_at"`
}

type Team struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type TeamMember struct {
	ID           int       `json:"id"`
	TeamID       int       `json:"team_id"`
	UserID       int       `json:"user_id"`
	Productivity int       `json:"productivity"`
	CreatedAt    time.Time `json:"created_at"`
	// Joined fields for convenience
	UserName string `json:"user_name,omitempty"`
	TeamName string `json:"team_name,omitempty"`
}

type Presence struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Date      string    `json:"date"` // YYYY-MM-DD
	StatusAM  string    `json:"status_am"`
	StatusPM  string    `json:"status_pm"`
	CreatedAt time.Time `json:"created_at"`
}

func New(dataSourceName string) (*Store, error) {
	db, err := sql.Open("pgx", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("error opening db: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error pinging db: %w", err)
	}

	return &Store{db: db}, nil
}

func (s *Store) InitSchema(schemaPath string) error {
	content, err := ioutil.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("error reading schema file: %w", err)
	}

	if _, err := s.db.Exec(string(content)); err != nil {
		return fmt.Errorf("error executing schema: %w", err)
	}

	return nil
}

func (s *Store) Close() {
	s.db.Close()
}

func (s *Store) CreateUser(ctx context.Context, u User) (User, error) {
	query := `INSERT INTO users (name, email, avatar_url) VALUES ($1, $2, $3) RETURNING id, created_at`
	err := s.db.QueryRowContext(ctx, query, u.Name, u.Email, u.AvatarURL).Scan(&u.ID, &u.CreatedAt)
	if err != nil {
		return User{}, err
	}
	return u, nil
}

func (s *Store) GetUsers(ctx context.Context) ([]User, error) {
	query := `SELECT id, name, email, COALESCE(avatar_url, ''), created_at FROM users`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.AvatarURL, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (s *Store) UpsertPresence(ctx context.Context, p Presence) (Presence, error) {
	query := `
		INSERT INTO presence (user_id, date, status_am, status_pm)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, date) DO UPDATE SET 
			status_am = EXCLUDED.status_am,
			status_pm = EXCLUDED.status_pm
		RETURNING id, created_at
	`
	err := s.db.QueryRowContext(ctx, query, p.UserID, p.Date, p.StatusAM, p.StatusPM).Scan(&p.ID, &p.CreatedAt)
	if err != nil {
		return Presence{}, err
	}
	return p, nil
}

func (s *Store) GetPresence(ctx context.Context, start, end string) ([]Presence, error) {
	query := `SELECT id, user_id, to_char(date, 'YYYY-MM-DD'), COALESCE(status_am, ''), COALESCE(status_pm, ''), created_at FROM presence WHERE date >= $1::date AND date <= $2::date`

	rows, err := s.db.QueryContext(ctx, query, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []Presence
	for rows.Next() {
		var p Presence
		if err := rows.Scan(&p.ID, &p.UserID, &p.Date, &p.StatusAM, &p.StatusPM, &p.CreatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, p)
	}
	return entries, nil
}

func (s *Store) CreateTeam(ctx context.Context, name string) (Team, error) {
	query := `INSERT INTO teams (name) VALUES ($1) RETURNING id, created_at`
	var t Team
	t.Name = name
	err := s.db.QueryRowContext(ctx, query, name).Scan(&t.ID, &t.CreatedAt)
	if err != nil {
		return Team{}, err
	}
	return t, nil
}

func (s *Store) GetTeams(ctx context.Context) ([]Team, error) {
	query := `SELECT id, name, created_at FROM teams ORDER BY name`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []Team
	for rows.Next() {
		var t Team
		if err := rows.Scan(&t.ID, &t.Name, &t.CreatedAt); err != nil {
			return nil, err
		}
		teams = append(teams, t)
	}
	return teams, nil
}

func (s *Store) AddUserToTeam(ctx context.Context, teamID, userID, productivity int) (TeamMember, error) {
	query := `
		INSERT INTO team_members (team_id, user_id, productivity)
		VALUES ($1, $2, $3)
		ON CONFLICT (team_id, user_id) DO UPDATE SET productivity = EXCLUDED.productivity
		RETURNING id, created_at
	`
	var tm TeamMember
	tm.TeamID = teamID
	tm.UserID = userID
	tm.Productivity = productivity

	err := s.db.QueryRowContext(ctx, query, teamID, userID, productivity).Scan(&tm.ID, &tm.CreatedAt)
	if err != nil {
		return TeamMember{}, err
	}
	return tm, nil
}

func (s *Store) GetTeamMembers(ctx context.Context, teamID int) ([]TeamMember, error) {
	query := `
		SELECT tm.id, tm.team_id, tm.user_id, tm.productivity, tm.created_at, u.name
		FROM team_members tm
		JOIN users u ON tm.user_id = u.id
		WHERE tm.team_id = $1
	`
	rows, err := s.db.QueryContext(ctx, query, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []TeamMember
	for rows.Next() {
		var tm TeamMember
		if err := rows.Scan(&tm.ID, &tm.TeamID, &tm.UserID, &tm.Productivity, &tm.CreatedAt, &tm.UserName); err != nil {
			return nil, err
		}
		members = append(members, tm)
	}
	return members, nil
}

func (s *Store) RemoveUserFromTeam(ctx context.Context, teamID, userID int) error {
	query := `DELETE FROM team_members WHERE team_id = $1 AND user_id = $2`
	_, err := s.db.ExecContext(ctx, query, teamID, userID)
	return err
}

func (s *Store) DeleteUser(ctx context.Context, userID int) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := s.db.ExecContext(ctx, query, userID)
	return err
}

func (s *Store) GetUserTeams(ctx context.Context, userID int) ([]TeamMember, error) {
	query := `
		SELECT tm.id, tm.team_id, tm.user_id, tm.productivity, tm.created_at, t.name
		FROM team_members tm
		JOIN teams t ON tm.team_id = t.id
		WHERE tm.user_id = $1
	`
	// Note: TeamMember struct uses "UserName", let's hijack it or add "TeamName"?
	// Current TeamMember struct:
	// type TeamMember struct { ... UserName string }
	// I should probably add TeamName to TeamMember struct or create a DTO.
	// For simplicity, let's just add TeamName to TeamMember struct definition first.
	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []TeamMember
	for rows.Next() {
		var tm TeamMember
		var teamName string
		if err := rows.Scan(&tm.ID, &tm.TeamID, &tm.UserID, &tm.Productivity, &tm.CreatedAt, &teamName); err != nil {
			return nil, err
		}
		tm.TeamName = teamName
		members = append(members, tm)
	}
	return members, nil
}
