# Task Management API

A multi-user Task Management REST API built with Go, Echo, PostgreSQL, and Redis.

## Architecture

```
Clean Architecture (3 Layers)
┌─────────────────────────────────────┐
│  Handler (HTTP / Echo)              │  ← Request validation, routing
├─────────────────────────────────────┤
│  Usecase (Business Logic)           │  ← Core rules, transactions, idempotency
├─────────────────────────────────────┤
│  Repository (Data Access)           │  ← PostgreSQL queries, Redis operations
└─────────────────────────────────────┘
```

**Key design decisions:**
- **Interface-driven**: All repository and usecase layers use interfaces, enabling unit testing with mocks
- **Idempotency**: Uses Redis `SETNX` for atomic lock acquisition + 24h TTL cached responses
- **Transactions**: Task assignment runs in a single database transaction (update + audit log + notification)
- **Structured logging**: Zap JSON logger with request_id correlation across all log entries
- **Team membership**: Tracked via `team_id` (INT) on the `users` table — no separate teams table required

## Tech Stack

| Component | Technology |
|-----------|-----------|
| Language | Go 1.26 |
| Web Framework | Echo v4 |
| Database | PostgreSQL 16 |
| Cache | Redis 7 |
| Authentication | JWT (HS256) |
| Logging | Uber Zap |
| Migrations | golang-migrate |
| Config | Viper (YAML + env vars) |

## Prerequisites

- Go 1.26+
- PostgreSQL 16+
- Redis 7+
- Docker & Docker Compose (optional, for containerized setup)

## Quick Start

### Option 1: Docker Compose (Recommended)

```bash
# Start all services (Server + PostgreSQL + Redis)
docker-compose up --build

# The server will be available at http://localhost:8080
```

### Option 2: Local Development

1. **Set up the database:**
   ```bash
   createdb task_database
   ```

2. **Create config file:**
   ```bash
   cp config.yaml.example config.yaml
   # Edit config.yaml with your database/Redis credentials
   ```

3. **Run the application:**
   ```bash
   go run ./cmd/server
   ```

The server will start on `http://localhost:8080`. Migrations run automatically on startup.

## Running Tests

```bash
# Run all tests with race detector
go test ./... -v -race

# Run only idempotency race condition tests
go test ./internal/usecase/ -v -race -run TestIdempotency
```

## Postman Collection

A Postman collection is included in the root folder as task_manager.postman_collection.json.

## API Endpoints

### Authentication

#### Register
```
POST /auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "username": "johndoe",
  "password": "securepassword",
  "team_id": 1
}
```

#### Login
```
POST /auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "securepassword"
}
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "token": "eyJhbGci...",
    "user": {
      "id": "uuid",
      "email": "user@example.com",
      "username": "johndoe",
      "team_id": 1
    }
  },
  "timestamp": "2026-06-27T12:00:00Z"
}
```

### Tasks (JWT Required)

All task endpoints require the `Authorization: Bearer <token>` header.

#### Create Task
```
POST /tasks
Content-Type: application/json
Authorization: Bearer <token>
Idempotency-Key: <uuid>  (optional but recommended)

{
  "title": "Implement feature X",
  "description": "Detailed description here"
}
```

#### List Tasks
```
GET /tasks?status=todo&title=feature&page=1&limit=10
Authorization: Bearer <token>
```

**Query Parameters:**
| Param | Type | Description |
|-------|------|-------------|
| `status` | string | Filter by status: `todo`, `in_progress`, `done` |
| `title` | string | Search by title (case-insensitive) |
| `page` | int | Page number (default: 1) |
| `limit` | int | Items per page (default: 10, max: 100) |

#### Get Task
```
GET /tasks/:id
Authorization: Bearer <token>
```

#### Update Task
```
PUT /tasks/:id
Content-Type: application/json
Authorization: Bearer <token>

{
  "title": "Updated title",
  "status": "in_progress"
}
```

#### Delete Task
```
DELETE /tasks/:id
Authorization: Bearer <token>
```

#### Assign Task
```
POST /tasks/:id/assign
Content-Type: application/json
Authorization: Bearer <token>

{
  "assignee_id": "uuid-of-team-member"
}
```

### Health Check
```
GET /health
```

## Error Response Format

All errors follow a consistent JSON structure:

```json
{
  "status": "error",
  "code": "VALIDATION_ERROR",
  "message": "title is required",
  "timestamp": "2026-06-27T12:00:00Z"
}
```

**Error codes:**
| Code | HTTP Status | Description |
|------|-------------|-------------|
| `VALIDATION_ERROR` | 422 | Invalid input |
| `UNAUTHORIZED` | 401 | Missing/invalid token |
| `FORBIDDEN` | 403 | Insufficient permissions |
| `NOT_FOUND` | 404 | Resource not found |
| `CONFLICT` | 409 | Duplicate resource |
| `INTERNAL_ERROR` | 500 | Server error (details hidden) |

## Database Schema

| Table | Description |
|-------|-------------|
| `users` | Registered users (email, username, password hash, team_id) |
| `tasks` | Tasks with status, creator, and optional assignee |
| `task_logs` | Audit trail for task changes (JSONB old/new values) |

## Project Structure

```
├── cmd/server/main.go       # Entry point, DI, routing
├── config/                  # Configuration & module setup loaders (postgres, redis, app, logger)
├── internal/
│   ├── entity/              # Entities (pure Go structs)
│   ├── dto/                 # Request/response DTOs
│   ├── handler/             # HTTP handlers (Echo)
│   ├── middleware/           # JWT auth, logging, recovery, request ID
│   ├── repository/          # Data access layer (interfaces + PostgreSQL)
│   ├── usecase/             # Business logic layer
│   └── pkg/                 # Shared utilities (errors, responses, validation)
├── migrations/              # SQL migration files
├── Dockerfile               # Multi-stage production build
├── docker-compose.yml       # Full stack (Server + PostgreSQL + Redis)
└── config.yaml.example      # Configuration template
```
