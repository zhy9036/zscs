# Zscaler Migration Platform

ChatGPT-style Zscaler migration application. Go backend + React/TypeScript frontend.

See `zscaler_system.md` for the full design spec.

## Quick start

### Prerequisites
- Go 1.24+
- Node.js 20+
- Docker (for Postgres)
- [golang-migrate](https://github.com/golang-migrate/migrate) CLI

### 1. Start Postgres
```bash
docker compose up -d postgres
```

### 2. Run migrations
```bash
migrate -path backend/migrations \
  -database "postgres://zscaler:zscaler_dev_password@localhost:5432/zscaler_migration?sslmode=disable" \
  up
```

### 3. Seed a user
```bash
cd backend
SEED_USERNAME=admin SEED_PASSWORD=changeme123 \
  DATABASE_URL="postgres://zscaler:zscaler_dev_password@localhost:5432/zscaler_migration?sslmode=disable" \
  go run ./cmd/seed
```

### 4. Run the backend
```bash
cd backend
go run ./cmd/server
```

### 5. Run the frontend
```bash
cd frontend
npm install
npm run dev
```

Open http://localhost:5175 and log in with the seeded user.

## Configuration

See `.env.example` for all environment variables.

## Development

For local iteration, enable dev-only user registration:
```bash
ENABLE_DEV_REGISTER=true
```
Then `POST /api/v1/auth/register` becomes available.
