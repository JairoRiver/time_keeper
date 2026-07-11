# Time Keeper

A full-stack time tracking application. Log in, create time entries with tags, and manage your work sessions through a web interface.

## Tech Stack

**Backend**
- Go 1.23 · Echo v4 · PostgreSQL 16
- Email/password auth (argon2id) · JWT sessions · SQLC · golang-migrate · Swagger

**Frontend**
- Server-rendered templ · Tailwind CSS v4

**Deployment**
- Podman · Podman Compose

## Prerequisites

- Go 1.23+
- Node.js 18+
- Podman & Podman Compose (podman-compose)

## Quickstart (Podman Compose)

The easiest way to run the full stack:

```bash
podman compose -f deploy/docker-compose/quickstart.yaml up --build
```

The app serves both the API and the web UI on a single port.

| Service | URL                          |
|---------|------------------------------|
| Web UI  | http://localhost:8080        |
| Swagger | http://localhost:8080/api/v1/swagger/ (dev only) |

## Manual Setup

### 1. Database

```bash
task postgres     # start PostgreSQL container
task createdb     # create the database
task migrateup    # run migrations
```

### 2. Backend

Edit `config.yaml` if needed, then:

```bash
task server       # go run ./main.go serve
```

### 3. Frontend

```bash
cd web
npm install
npm run dev
```

## Configuration

`config.yaml` (root):

```yaml
database:
  db_source: postgresql://root:secret@localhost:5432/time_keeper?sslmode=disable
  db_name: time_keeper
server:
  address: 0.0.0.0:8080
  secure_cookies: ${SECURE_COOKIES:-false}   # set true behind HTTPS in prod
  enable_swagger: ${ENABLE_SWAGGER:-true}    # disable in prod
```

## API Reference

Base URL: `http://localhost:8080/api/v1`

### Public

| Method | Path | Description |
|--------|------|-------------|
| POST | `/user` | Create user — returns access token + user ID |
| POST | `/refresh` | Refresh access token (requires `refresh_token` cookie) |

### Protected (Bearer token required)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/entry-time` | Create time entry |
| GET | `/entry-time/:id` | Get time entry by ID |
| PUT | `/entry-time` | Update time entry |
| DELETE | `/entry-time/:id` | Delete time entry |
| GET | `/entries-time?page_number=1` | List user's time entries (paginated) |

**Authorization header:** `Authorization: Bearer <access_token>`

Token lifetimes: access token 24 h · refresh token 30 days (HttpOnly cookie).

## Database Schema

**users** — `id`, `email`, `role`, `email_validated`, `is_active`, `secret_token_key`, `created_at`, `updated_at`

**time_entries** — `id`, `user_id`, `tag`, `time_start`, `time_end`, `created_at`, `updated_at`

## Task Commands

| Command | Description |
|---------|-------------|
| `task postgres` | Start PostgreSQL container |
| `task createdb` | Create database |
| `task dropdb` | Drop database |
| `task migrateup` | Run all pending migrations |
| `task migratedown` | Roll back one migration |
| `task new_migration` | Create a new migration file |
| `task sqlc` | Regenerate SQLC code |
| `task test` | Run all tests |
| `task server` | Start backend server |
| `task api_doc` | Regenerate Swagger docs |
| `task compose` | Run full stack with Podman Compose |

## Testing

```bash
task test
# or
go test -v -cover ./...
```

Tests live in `internal/controller/`, `internal/api/handler/`, and `pkg/token/`.
