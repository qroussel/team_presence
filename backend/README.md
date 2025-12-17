# Team Presence Backend

This is the Go-based backend service for the Team Presence application. It provides a RESTful API for managing users, teams, and presence data, backed by a PostgreSQL database.

## Prerequisites

- **Go**: Version 1.21 or later.
- **PostgreSQL**: A running PostgreSQL instance.

## Configuration

The application is configured using environment variables.

| Variable | Description | Default |
| :--- | :--- | :--- |
| `PORT` | The port the server listens on | `8080` |
| `DATABASE_URL` | PostgreSQL connection string | `postgres://presence_user:presence_password@localhost:5432/presence_db?sslmode=disable` |

## Running Locally

1.  **Start Database**: Ensure your PostgreSQL database is running and accessible via `DATABASE_URL`.
2.  **Install Dependencies**:
    ```bash
    go mod download
    ```
3.  **Run Server**:
    ```bash
    go run main.go
    ```
    The server will start on port 8080 (or your configured `PORT`).

## API Endpoints

-   `GET /api/health`: Health check.
-   `GET /api/users`: List users.
-   `POST /api/users`: Create a user.
-   `DELETE /api/users?id=<id>`: Delete a user.
-   `GET /api/presence?start=<date>&end=<date>`: Get presence data.
-   `POST /api/presence`: Upsert presence data.
-   `GET /api/teams`: List teams.
-   `POST /api/teams`: Create a team.
-   `GET /api/team_members?team_id=<id> | user_id=<id>`: Get members of a team or teams of a user.
-   `POST /api/team_members`: Add a user to a team.
-   `DELETE /api/team_members?team_id=<id>&user_id=<id>`: Remove a user from a team.
