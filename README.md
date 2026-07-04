# tiny mchwa 🐜

Todo list microservice for the homelab ecosystem. "Mchwa" is Swahili for ant.

## Tech Stack

| Layer | Choice |
|-------|--------|
| Backend | Go + Fiber v3 |
| Database | PostgreSQL |
| Auth | Keycloak (OAuth) — planned, not yet implemented |
| Testing | testify |

## Quick Start

```bash
# Set up env
cp .env.example .env
# Edit .env with your Postgres password

# Run
DATABASE_URL=postgres://postgres:***@localhost:5432/tiny_mchwa?sslmode=disable go run ./cmd/server/
```

Server starts on `http://localhost:3000`.

## Project Structure

```
cmd/server/main.go          ← Entry point
internal/
├── app/app.go               ← Wiring, routes, middleware
├── config/config.go         ← Env-based config + logger setup
├── database/db.go           ← sqlx connection pool
├── model/                   ← Domain types, request/response DTOs
├── repository/              ← SQL queries (data access)
├── service/                 ← Business logic, computed status
├── handler/                 ← HTTP handlers
└── middleware/               ← RequestID, Recover, Helmet, Logger
```

## API

Base URL: `/api/v1`

### Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/todolists` | Create a todolist |
| `GET` | `/api/v1/todolists` | List todolists (paginated, filterable) |
| `GET` | `/api/v1/todolists/:id` | Get a todolist |
| `PUT` | `/api/v1/todolists/:id` | Update a todolist |
| `DELETE` | `/api/v1/todolists/:id` | Delete a todolist (cascades to tasks) |
| `GET` | `/health` | Health check |

### Pagination

```
GET /api/v1/todolists?page=1&perPage=20
```

### Filtering

```
GET /api/v1/todolists?status=inprogress&sourceService=blog&title=grocery
```

### Response Format

```json
{
    "data": {},
    "error": null,
    "meta": { "page": 1, "perPage": 20, "total": 45, "totalPages": 3 }
}
```

## Status Logic

Todolist status is **computed** from tasks, not stored:

| Condition | Status |
|-----------|--------|
| No tasks | `pending` |
| All tasks `pending` | `pending` |
| Any task `inprogress` | `inprogress` |
| All tasks `done` | `done` |

## MVP Scope

**In scope:** CRUD todolists, pagination, filtering.

**Out of scope:** CRUD tasks, task status computation, frontend, cross-service integration, Keycloak auth.

## Config

| Env Var | Default | Description |
|---------|---------|-------------|
| `PORT` | `3000` | Server port |
| `DATABASE_URL` | — | Postgres connection string |
| `LOG_LEVEL` | `info` | `info` or `debug` |

## Development

```bash
make run      # Start server
make build    # Build binary to bin/server
make test     # Run tests
make vet      # Run go vet
make lint     # Alias for vet
make dev      # Hot reload with air (install: go install github.com/air-verse/air@latest)
make clean    # Remove build artifacts
```
