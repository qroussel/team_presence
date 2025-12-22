-- name: CreateUser :one
INSERT INTO users (name, email, avatar_url, role) VALUES ($1, $2, $3, $4) RETURNING id, created_at;

-- name: GetUsers :many
SELECT id, name, email, COALESCE(avatar_url, '') as avatar_url, role, created_at FROM users;

-- name: GetUserByID :one
SELECT id, name, email, COALESCE(avatar_url, '') as avatar_url, role, created_at FROM users WHERE id = $1;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;

-- name: UpsertPresence :one
INSERT INTO presence (user_id, date, status_am, status_pm)
VALUES ($1, $2, $3, $4)
ON CONFLICT (user_id, date) DO UPDATE SET 
    status_am = EXCLUDED.status_am,
    status_pm = EXCLUDED.status_pm
RETURNING id, created_at;

-- name: GetPresence :many
SELECT id, user_id, to_char(date, 'YYYY-MM-DD') as date, COALESCE(status_am, '') as status_am, COALESCE(status_pm, '') as status_pm, created_at 
FROM presence 
WHERE date >= $1::date AND date <= $2::date;

-- name: CreateTeam :one
INSERT INTO teams (name, owner_id) VALUES ($1, $2) RETURNING id, owner_id, created_at;

-- name: GetTeams :many
SELECT id, name, owner_id, created_at FROM teams ORDER BY name;

-- name: GetTeamByID :one
SELECT id, name, owner_id, created_at FROM teams WHERE id = $1;

-- name: DeleteTeam :exec
DELETE FROM teams WHERE id = $1;

-- name: AddUserToTeam :one
INSERT INTO team_members (team_id, user_id, productivity, role)
VALUES ($1, $2, $3, $4)
ON CONFLICT (team_id, user_id) DO UPDATE SET productivity = EXCLUDED.productivity
RETURNING id, created_at;

-- name: GetTeamMembers :many
SELECT tm.id, tm.team_id, tm.user_id, tm.productivity, tm.role, tm.created_at, u.name as user_name
FROM team_members tm
JOIN users u ON tm.user_id = u.id
WHERE tm.team_id = $1;

-- name: RemoveUserFromTeam :exec
DELETE FROM team_members WHERE team_id = $1 AND user_id = $2;

-- name: GetUserTeams :many
SELECT tm.id, tm.team_id, tm.user_id, tm.productivity, tm.role, tm.created_at, t.name as team_name, t.owner_id
FROM team_members tm
JOIN teams t ON tm.team_id = t.id
WHERE tm.user_id = $1;

-- name: UpdateTeamMemberRole :exec
UPDATE team_members SET role = $3 WHERE team_id = $1 AND user_id = $2;

-- name: UpdateTeamOwner :exec
UPDATE teams SET owner_id = $2 WHERE id = $1;
