# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Status

FitProof is a deployed production app: a Go 1.25 backend on Render, a Next.js 16 frontend on Vercel, and a Neon-hosted PostgreSQL database. The `main` branch is the live branch — both Render and Vercel deploy from it. The original spec at `docs/Fitness Challenge App.md` is a planning doc, not the current source of truth; the code is.

## Tech Stack

**Backend (Go 1.25):**
- HTTP Router: chi (`github.com/go-chi/chi/v5`)
- Database: PostgreSQL (Neon in production)
- Query layer: sqlc + pgx (`github.com/jackc/pgx/v5`)
- Migrations: golang-migrate (auto-run on startup)
- Auth: JWT access tokens + refresh tokens (HTTP-only cookies); Google OAuth optional
- Password hashing: bcrypt
- Email: SMTP via `github.com/wneessen/go-mail` (used for OTP verification)
- Media storage: Cloudflare R2 via AWS SDK v2 (optional — disabled if R2 env vars are blank)

**Frontend:** Next.js 16, React 19, Tailwind CSS v4, TypeScript

**Deployment:** Render (backend) + Vercel (frontend) + Neon (Postgres). Docker / docker-compose is for the local dev stack only.

## Commands

```bash
# Backend (run from backend/)
cd backend
go run ./cmd/server          # Run the server (migrations auto-apply on startup)
go build ./cmd/server        # Build the binary
go test ./...                # Run all tests
go test ./internal/services/... # Run tests for a specific package
go vet ./...                 # Static analysis

# Frontend (run from frontend/)
cd frontend
npm run dev                  # Dev server on :3000 with /api proxy to :8080
npm run build                # Production build
```

Migrations auto-run on backend startup via `db.RunMigrations`. The manual `migrate` CLI is still useful for local resets:

```bash
migrate -path ./migrations -database $DATABASE_URL up
migrate -path ./migrations -database $DATABASE_URL down 1
```

sqlc code generation (after editing `.sql` files in `internal/repositories/queries/`):

```bash
sqlc generate
```

## Architecture

Three-layer architecture with strict separation:

```
Handler → Service → Repository
```

- **`internal/handlers/`** — HTTP request/response handling, input validation, JSON serialization
- **`internal/services/`** — Business logic: scoring rules evaluation, leaderboard calculation, submission validation, cron tasks
- **`internal/repositories/`** — Database access via sqlc-generated type-safe queries
- **`internal/repositories/db/`** — sqlc-generated code; **do not hand-edit** (regenerate via `sqlc generate`)
- **`internal/repositories/queries/`** — `.sql` query files that drive sqlc generation
- **`internal/middleware/`** — JWT authentication, request logging, panic recovery
- **`internal/models/`** — Shared data structures/entities
- **`internal/db/`** — pgx pool setup and migration runner
- **`internal/config/`** — env-var loader (`Load()` is the authoritative list of every var the backend reads)
- **`cmd/server/`** — Entry point, dependency wiring, route definitions, cron loop

Four shared packages under `pkg/` (usable outside `internal/`):

- **`pkg/storage/`** — Cloudflare R2 client for media uploads/deletes
- **`pkg/email/`** — SMTP sender (no-op when `SMTP_HOST` is blank)
- **`pkg/drive/`** — Google Drive helper
- **`pkg/respond/`** — HTTP JSON-response helper

## Database Schema

Eleven tables in PostgreSQL:

**Auth:** `users`, `refresh_tokens`, `email_verification_tokens`
**Domain:** `metrics`, `challenges`, `user_challenges`, `challenge_metrics`
**Events:** `daily_submissions`, `submission_metric_values`, `submission_media`
**Derived state:** `challenge_scores` (pre-aggregated leaderboard state)

Key constraints:
- `user_challenges`: UNIQUE(user_id, challenge_id)
- `daily_submissions`: UNIQUE(user_challenge_id, date)
- `refresh_tokens`: unique per user + device; stored as hashed values

Indexes needed on: `challenge_id`, `user_challenge_id`, `date`.

Migrations live under `backend/migrations/`:
1. `000001_initial_schema`
2. `000002_seed_metrics`
3. `000003_add_oauth_and_email_verification`
4. `000004_add_media_key`
5. `000005_submission_media`

## Key Business Logic

**Submission flow** (`POST /challenges/:id/submissions`) runs in a single DB transaction:
1. Validate JWT → validate user_challenge → check no duplicate submission for today
2. Insert `daily_submissions`
3. Load `challenge_metrics` rules
4. Evaluate each metric (min/max target) → insert `submission_metric_values` (with `passed` bool and `points_awarded`)
5. Update `challenge_scores` (total_points, last_submission_date)

**Daily cron** (runs at 00:05 UTC, see `runDailyCron` in `cmd/server/main.go`) performs three tasks in sequence:
1. `ProcessMissedSubmissions` — for each participant with no submission yesterday, insert a `daily_submissions` row with `submission_type = 'missed'` and apply `fine_amount` to `challenge_scores.total_fines`.
2. `ProcessMissingMedia` — for submissions yesterday that required media but didn't include it, apply the same fine.
3. `DeleteExpiredMedia` — purge DB rows in `submission_media` older than 7 days. The actual R2 objects are expired by an R2 lifecycle rule (configured manually in the Cloudflare dashboard).

**Leaderboard** (`GET /challenges/:id/leaderboard`):
- Reads from `challenge_scores` (pre-aggregated, no expensive history scan)
- Ranks users with SQL window functions at read time

## Auth Design

- **JWT access tokens:** short-lived (default 15m via `JWT_ACCESS_TOKEN_TTL`), validated in middleware
- **Refresh tokens:** cryptographically random, stored as hashed values in `refresh_tokens`, issued per device session (tracked by device_id, ip_address, user_agent), individually revocable; default TTL 168h via `JWT_REFRESH_TOKEN_TTL`
- Both tokens stored in HTTP-only cookies
- **Google OAuth:** browser-initiated `/auth/google` → `/auth/google/callback`. Disabled at runtime if `GOOGLE_CLIENT_ID` or `GOOGLE_CLIENT_SECRET` is blank.
- **Email verification:** 6-digit OTP, bcrypt-hashed in `email_verification_tokens`, TTL configurable via `EMAIL_VERIFICATION_TTL` (default 15m). When `SMTP_HOST` is blank, OTPs are logged to stdout for local dev.
