# Team Presence Backend

The backend service for the Team Presence application, providing a robust RESTful API for managing users, presence, and teams. Built with Go and PostgreSQL.

## Architecture

The application follows a clean, modular architecture separating concerns between the API layer and the data access layer.

### Layers

- **Main Entry (`main.go`)**: Responsible for initialization, dependency injection, and starting the HTTP server. It composes the `Store` and `Server`.
- **API Layer (`internal/api`)**: Handles HTTP requests, input validation, and maps DTOs (Data Transfer Objects) to domain models. It contains:
    - `server.go`: Router configuration and server definition.
    - `handlers.go`: HTTP handler logic.
    - `types.go`: Request/Response structs (DTOs) ensuring strict API contracts.
- **Data Access Layer (`internal/store`)**: Interacts directly with PostgreSQL. This layer is **generated** to ensure type safety and performance.

### Project Structure

```
backend/
├── internal/
│   ├── api/        # HTTP Handlers, Routing, DTOs
│   ├── store/      # Generated Go code from SQL (Models, Queries)
│   └── testutil/   # Testing helpers and fixtures
├── main.go         # Entry point
├── query.sql       # Raw SQL queries used by sqlc
├── schema.sql      # Database schema definition
└── sqlc.yaml       # Code generation configuration
```

## Architectural Choices & Key Technologies

### 1. Database Access: `sqlc` + `pgx`
Instead of a heavy ORM (like GORM), we use [sqlc](https://sqlc.dev/) with [pgx](https://github.com/jackc/pgx).
- **Why**: `sqlc` compiles raw SQL queries into type-safe Go code. This prevents runtime SQL errors and provides correct types for bindings.
- **Performance**: `pgx` is a high-performance driver optimized for PostgreSQL features.
- **Type Simplification**: We configured `sqlc` to map PostgreSQL types to **standard Go types** (e.g., `int32`, `string`, `time.Time`) rather than driver-specific wrappers (like `pgtype.Int4`). This keeps the business logic clean and readable.

### 2. Testing Strategy: Transactional Isolation
Our tests are designed for speed and reliability.
- **Integration over Mocking**: Since most logic is in data persistence, we test against a real PostgreSQL database.
- **Transaction Rollback**: Each test connects to the database, starts a transaction, performs operations, and **rolls back** the transaction at the end.
    - **Benefit**: This guarantees a clean state for every test case.
    - **Benefit**: It allows tests to run in **parallel** (`t.Parallel()`) without race conditions or data pollution.

### 3. API Design
- **RESTful**: Standard HTTP verbs and resources.
- **Explicit DTOs**: We use specific structs for requests and responses (e.g., `CreateUserRequest`, `UserResponse`) rather than reusing database models. This decouples the public API from the internal database schema.

## Getting Started

### Prerequisites
- **Go**: Version 1.21+
- **PostgreSQL**: A running instance.

### Configuration
Environment variables (optional, defaults provided):

| Variable | Description | Default |
| :--- | :--- | :--- |
| `PORT` | API Server port | `8080` |
| `DATABASE_URL` | PostgreSQL DSZ | `postgres://presence_user:presence_password@localhost:5432/presence_db?sslmode=disable` |

### Running the App
1. **Start Database**: Ensure your DB is reachable.
2. **Install Dependencies**:
   ```bash
   go mod download
   ```
3. **Run**:
   ```bash
   go run main.go
   ```

## Testing

To run the test suite:

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...
```

**Note**: Tests require a running database instance accessible via the `DATABASE_URL` used in `internal/testutil/helpers.go`.

## API Reference

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `GET` | `/api/health` | Health check |
| `GET` | `/api/users` | List all users |
| `POST` | `/api/users` | Create a new user |
| `DELETE` | `/api/users?id={id}` | Delete a user |
| `GET` | `/api/presence` | Get presence (params: `start`, `end`) |
| `POST` | `/api/presence` | Upsert presence for a user/date |
| `GET` | `/api/teams` | List all teams |
| `POST` | `/api/teams` | Create a new team |
| `GET` | `/api/team_members` | Get members (params: `team_id` or `user_id`) |
| `POST` | `/api/team_members` | Add user to team |
| `PUT` | `/api/team_members` | Update member role/role/productivity |
| `DELETE` | `/api/team_members` | Remove user from team |

## RBAC & Permissions

The application uses a 3-level Role-Based Access Control system: **Admin**, **Team Owner**, and **Member/User**.

| Endpoint | Method | Admin | Team Owner | User | Note |
| :--- | :--- | :---: | :---: | :---: | :--- |
| `/api/teams` | `POST` | ✅ | ✅ | ✅ | Creator becomes Owner |
| `/api/teams` | `DELETE` | ✅ | ✅ (Own Team) | ❌ | |
| `/api/team_members` | `POST` | ✅ | ✅ (Own Team) | ❌ | Add Members |
| `/api/team_members` | `PUT` | ✅ | ✅ (Own Team) | ❌ | Promote/Demote |
| `/api/team_members` | `DELETE`| ✅ | ✅ (Own Team) | ✅ (Self) | Remove/Leave |
| `/api/users` | `POST/DELETE` | ✅ | ❌ | ❌ | Admin only |
| `/api/presence` | `POST` | ✅ | ✅ (Team Members) | ✅ (Self) | |
