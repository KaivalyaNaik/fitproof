# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Status

This is a **Go backend in early implementation phase**. The architecture and database design are fully documented in `docs/Fitness Challenge App.md`. Directory structure exists but `go.mod` and all code files are yet to be created.

## Tech Stack

- **Language:** Go
- **HTTP Router:** chi (`github.com/go-chi/chi/v5`)
- **Database:** PostgreSQL
- **Query layer:** sqlc + pgx (`github.com/jackc/pgx/v5`)
- **Migrations:** golang-migrate
- **Auth:** JWT access tokens + refresh tokens (HTTP-only cookies)
- **Password hashing:** bcrypt or argon2
- **Deployment:** Docker

## Commands

Once `go.mod` is initialized:

```bash
go run ./cmd/server          # Run the server
go build ./cmd/server        # Build the binary
go test ./...                # Run all tests
go test ./internal/services/... # Run tests for a specific package
go vet ./...                 # Static analysis
```

Migration commands (golang-migrate):
```bash
migrate -path ./migrations -database $DATABASE_URL up
migrate -path ./migrations -database $DATABASE_URL down 1
```

sqlc code generation (after `sqlc.yaml` is configured):
```bash
sqlc generate
```

## Architecture

Three-layer architecture with strict separation:

```
Handler → Service → Repository
```

- **`internal/handlers/`** — HTTP request/response handling, input validation, JSON serialization
- **`internal/services/`** — Business logic: scoring rules evaluation, leaderboard calculation, submission validation
- **`internal/repositories/`** — All database access via sqlc-generated type-safe queries
- **`internal/middleware/`** — JWT authentication, logging, error handling
- **`internal/models/`** — Shared data structures/entities
- **`cmd/server/`** — Entry point, dependency wiring, server startup
- **`pkg/`** — Shared utilities usable outside `internal/`

## Database Schema

Nine tables in PostgreSQL:

**Auth:** `users`, `refresh_tokens`
**Domain:** `metrics`, `challenges`, `user_challenges`, `challenge_metrics`
**Events:** `daily_submissions`, `submission_metric_values`
**Derived state:** `challenge_scores` (pre-aggregated leaderboard state)

Key constraints:
- `user_challenges`: UNIQUE(user_id, challenge_id)
- `daily_submissions`: UNIQUE(user_challenge_id, date)
- `refresh_tokens`: unique per user + device; stored as hashed values

Indexes needed on: `challenge_id`, `user_challenge_id`, `date`.

## Key Business Logic

**Submission flow** (`POST /submissions`) runs in a single DB transaction:
1. Validate JWT → validate user_challenge → check no duplicate submission for today
2. Insert `daily_submissions`
3. Load `challenge_metrics` rules
4. Evaluate each metric (min/max target) → insert `submission_metric_values` (with `passed` bool and `points_awarded`)
5. Update `challenge_scores` (total_points, last_submission_date)

**Missed submission cron** (runs daily at 00:05 UTC):
- Finds participants with no submission for the previous day
- Inserts a `daily_submissions` row with `submission_type = 'missed'`
- Applies `fine_amount` to `challenge_scores.total_fines`

**Leaderboard** (`GET /challenges/{id}/leaderboard`):
- Reads from `challenge_scores` (pre-aggregated, no expensive history scan)
- Ranks users with SQL window functions at read time

## Auth Design

- JWT access tokens: short-lived, validated in middleware
- Refresh tokens: cryptographically random, stored as hashed values in `refresh_tokens`, issued per device session (tracked by device_id, ip_address, user_agent), individually revocable
- Both tokens stored in HTTP-only cookies
