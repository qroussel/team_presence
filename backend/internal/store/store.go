package store

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TxContextKey is the key for transaction context values
type contextKey string

const TxContextKey contextKey = "tx"

type Store struct {
	q    *Queries
	pool *pgxpool.Pool
}

func NewStore(ctx context.Context, dataSourceName string) (*Store, error) {
	pool, err := pgxpool.New(ctx, dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("error creating pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("error pinging db: %w", err)
	}

	return &Store{
		q:    New(pool),
		pool: pool,
	}, nil
}

func (s *Store) InitSchema(schemaPath string) error {
	content, err := ioutil.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("error reading schema file: %w", err)
	}

	if _, err := s.pool.Exec(context.Background(), string(content)); err != nil {
		return fmt.Errorf("error executing schema: %w", err)
	}

	return nil
}

func (s *Store) Close() {
	s.pool.Close()
}

// BeginTx begins a transaction for test isolation
func (s *Store) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return s.pool.Begin(ctx)
}

// getQueries returns the appropriate queries object (tx or pool) based on context
func (s *Store) getQueries(ctx context.Context) *Queries {
	if tx, ok := ctx.Value(TxContextKey).(pgx.Tx); ok {
		return s.q.WithTx(tx)
	}
	return s.q
}

// WithTx returns a context with the transaction injected
func WithTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, TxContextKey, tx)
}

// Wrappers for generated methods to support implicit transaction context

func (s *Store) CreateUser(ctx context.Context, arg CreateUserParams) (CreateUserRow, error) {
	return s.getQueries(ctx).CreateUser(ctx, arg)
}

func (s *Store) GetUsers(ctx context.Context) ([]GetUsersRow, error) {
	return s.getQueries(ctx).GetUsers(ctx)
}

func (s *Store) DeleteUser(ctx context.Context, id int32) error {
	return s.getQueries(ctx).DeleteUser(ctx, id)
}

func (s *Store) UpsertPresence(ctx context.Context, arg UpsertPresenceParams) (UpsertPresenceRow, error) {
	return s.getQueries(ctx).UpsertPresence(ctx, arg)
}

func (s *Store) GetPresence(ctx context.Context, arg GetPresenceParams) ([]GetPresenceRow, error) {
	return s.getQueries(ctx).GetPresence(ctx, arg)
}

func (s *Store) CreateTeam(ctx context.Context, name string) (CreateTeamRow, error) {
	return s.getQueries(ctx).CreateTeam(ctx, name)
}

func (s *Store) GetTeams(ctx context.Context) ([]Team, error) {
	return s.getQueries(ctx).GetTeams(ctx)
}

func (s *Store) DeleteTeam(ctx context.Context, id int32) error {
	return s.getQueries(ctx).DeleteTeam(ctx, id)
}

func (s *Store) AddUserToTeam(ctx context.Context, arg AddUserToTeamParams) (AddUserToTeamRow, error) {
	return s.getQueries(ctx).AddUserToTeam(ctx, arg)
}

func (s *Store) GetTeamMembers(ctx context.Context, teamID int32) ([]GetTeamMembersRow, error) {
	return s.getQueries(ctx).GetTeamMembers(ctx, teamID)
}

func (s *Store) RemoveUserFromTeam(ctx context.Context, arg RemoveUserFromTeamParams) error {
	return s.getQueries(ctx).RemoveUserFromTeam(ctx, arg)
}

func (s *Store) GetUserTeams(ctx context.Context, userID int32) ([]GetUserTeamsRow, error) {
	return s.getQueries(ctx).GetUserTeams(ctx, userID)
}
