CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    avatar_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS presence (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    status_am VARCHAR(50),
    status_pm VARCHAR(50),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, date)
);

CREATE TABLE IF NOT EXISTS teams (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS team_members (
    id SERIAL PRIMARY KEY,
    team_id INTEGER NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    productivity INTEGER DEFAULT 100, -- Percentage 0-100
    role VARCHAR(50) NOT NULL DEFAULT 'Member', -- 'Owner' or 'Member'
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(team_id, user_id)
);

-- Seed default team if not exists (This is a bit hacky in schema.sql but useful for dev)
INSERT INTO teams (name) VALUES ('Engineering') ON CONFLICT (name) DO NOTHING;

-- RBAC Migrations
ALTER TABLE users ADD COLUMN IF NOT EXISTS role VARCHAR(50) NOT NULL DEFAULT 'User';
ALTER TABLE teams ADD COLUMN IF NOT EXISTS owner_id INTEGER REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE team_members ADD COLUMN IF NOT EXISTS role VARCHAR(50) NOT NULL DEFAULT 'Member';

-- Migrate existing owners from teams table to team_members
-- 1. Insert owners into team_members if they are not already there (with Owner role)
INSERT INTO team_members (team_id, user_id, role)
SELECT id, owner_id, 'Owner'
FROM teams
WHERE owner_id IS NOT NULL
ON CONFLICT (team_id, user_id) DO UPDATE SET role = 'Owner';

-- Seed Admin User (Default)
INSERT INTO users (name, email, role) VALUES ('Admin User', 'admin@example.com', 'Admin') 
ON CONFLICT (email) DO UPDATE SET role = 'Admin';
