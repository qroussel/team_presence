# Team Presence Application

This is a small application for managing team presence, office days, and remote work, featuring a Go backend and a React frontend.

## Project Structure

- **`backend/`**: Contains the Go API service handling business logic and database interactions.
- **`frontend/`**: Contains the React application (Vite) for the user interface.

## Prerequisites

-   **Node.js**: v18 or later.
-   **Go**: v1.21 or later.
-   **Docker**: For running the PostgreSQL database.

## Quick Start

You can run the application using the provided development script or by managing the components manually.

### Option 1: Development Script

We provide a helper script to start the database, backend, and frontend with a single command.

```bash
./start_dev.sh
```

### Option 2: Manual Setup

1.  **Start PostgreSQL**:
    Run the database container:
    ```bash
    docker run --rm -d \
      --name presence-db \
      -e POSTGRES_USER=presence_user \
      -e POSTGRES_PASSWORD=presence_password \
      -e POSTGRES_DB=presence_db \
      -p 5432:5432 \
      postgres:15-alpine
    ```

2.  **Start Backend**:
    ```bash
    cd backend
    go run main.go
    ```
    The API will be available at `http://localhost:8080`.

3.  **Start Frontend**:
    Open a new terminal:
    ```bash
    cd frontend
    npm install # only first time
    npm run dev
    ```
    The UI will be available at `http://localhost:5173`.

## Manual Setup

For individual component setup and development, please refer to the specific READMEs:

- [Backend Documentation](./backend/README.md)
- [Frontend Documentation](./frontend/README.md)
