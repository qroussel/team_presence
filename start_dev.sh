#!/bin/bash

# Configuration
DB_NAME="presence-db"
DB_USER="presence_user"
DB_PASS="presence_password"
DB_PORT="5432"

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}Starting Team Presence Development Environment...${NC}"

# 1. Start PostgreSQL
echo -e "${GREEN}[1/3] Starting PostgreSQL...${NC}"
if [ ! "$(docker ps -q -f name=$DB_NAME)" ]; then
    if [ "$(docker ps -aq -f name=$DB_NAME)" ]; then
        echo "Starting existing container..."
        docker start $DB_NAME
    else
        echo "Creating new container..."
        docker run -d \
          --name $DB_NAME \
          -e POSTGRES_USER=$DB_USER \
          -e POSTGRES_PASSWORD=$DB_PASS \
          -e POSTGRES_DB=presence_db \
          -p $DB_PORT:$DB_PORT \
          postgres:15-alpine
    fi
else
    echo "PostgreSQL is already running."
fi

# 2. Start Backend
echo -e "${GREEN}[2/3] Starting Backend (Go)...${NC}"
cd backend
DATABASE_URL="postgres://$DB_USER:$DB_PASS@localhost:$DB_PORT/presence_db?sslmode=disable" go run main.go &
BACKEND_PID=$!
cd ..

# 3. Start Frontend
echo -e "${GREEN}[3/3] Starting Frontend (React)...${NC}"
cd frontend
# Check if node_modules exists
if [ ! -d "node_modules" ]; then
    echo "Installing frontend dependencies..."
    npm install
fi
npm run dev

# Cleanup on exit
trap "kill $BACKEND_PID" EXIT
