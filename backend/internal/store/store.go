package store

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"time"

	_ "github.com/lib/pq"
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

type Presence struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Date      string    `json:"date"` // YYYY-MM-DD
	StatusAM  string    `json:"status_am"`
	StatusPM  string    `json:"status_pm"`
	CreatedAt time.Time `json:"created_at"`
}

func New(dataSourceName string) (*Store, error) {
	db, err := sql.Open("postgres", dataSourceName)
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
