# FitProof

A fitness accountability app where users create and join challenges, submit daily proof of their workouts (with optional photo/video media), and get fined for missed days.

## Structure

```
fitproof/
├── backend/      Go API server (chi, pgx, sqlc, JWT auth)
├── frontend/     Next.js 16 web app (React 19, Tailwind CSS v4)
└── docker-compose.yml
```

## Features

- Create and join fitness challenges with custom metrics (reps, duration, distance, etc.)
- Daily submission tracking with pass/fail evaluation per metric
- Media uploads (photos/videos via Cloudflare R2, 7-day retention)
- Challenge feed and per-challenge fines summary
- Leaderboard based on pre-aggregated points and fines
- JWT auth with refresh tokens (HTTP-only cookies, per-device sessions)
- Google OAuth sign-in
- Email verification via 6-digit OTP
- Daily cron (00:05 UTC) for missed-submission fines, missing-media fines, and media cleanup

## Tech Stack

| Layer | Tech |
|---|---|
| Backend | Go 1.25, chi v5, pgx v5, sqlc, golang-migrate |
| Database | PostgreSQL 16 |
| Auth | JWT (golang-jwt/jwt/v5), bcrypt, Google OAuth2 |
| Email | go-mail (SMTP) |
| Storage | Cloudflare R2 (AWS SDK v2) |
| Frontend | Next.js 16, React 19, Tailwind CSS v4, TypeScript |
| Deployment | Render (backend), Vercel (frontend), Neon (Postgres); Docker / docker-compose for local dev |

## Getting Started

### Prerequisites

- Go 1.25+
- Node.js 20+
- PostgreSQL 16
- [golang-migrate](https://github.com/golang-migrate/migrate) CLI (only needed for manual migration commands; the backend auto-applies migrations on startup)
- [sqlc](https://sqlc.dev) (only needed if modifying SQL queries)

Cloudflare R2 is optional. If the `R2_*` env vars are blank, the server boots normally but media-upload routes will return an error. Local dev does not require R2.

### 1. Clone and configure

```bash
git clone https://github.com/KaivalyaNaik/fitproof.git
cd fitproof

cp backend/.env.example backend/.env
# Edit backend/.env with your values
```

### 2. Start the database

```bash
docker-compose up postgres -d
```

Or use an existing PostgreSQL instance and set `DATABASE_URL` in `backend/.env`.

### 3. Run the backend

```bash
cd backend
go run ./cmd/server
```

Migrations are applied automatically on startup. The server starts on port `8080`.

### 4. Run the frontend

```bash
cd frontend
npm install
npm run dev
```

The app is available at `http://localhost:3000`. API calls are proxied to `http://localhost:8080`.

## Environment Variables

All backend configuration lives in `backend/.env`. Copy from `backend/.env.example`. The authoritative source is `Load()` in [backend/internal/config/config.go](backend/internal/config/config.go).

### App

| Variable | Required | Default | Description |
|---|---|---|---|
| `APP_ENV` | No | `development` | Environment name; `production` enables secure cookies |
| `APP_PORT` | No | `8080` | HTTP listen port |

### Database

| Variable | Required | Default | Description |
|---|---|---|---|
| `DATABASE_URL` | Yes | — | PostgreSQL connection string. Use `pgx5://` scheme so golang-migrate's pgx/v5 driver can read it |
| `DB_MAX_CONNECTIONS` | No | `10` | pgx pool max connections |
| `DB_MIN_CONNECTIONS` | No | `2` | pgx pool min connections |
| `MIGRATIONS_PATH` | No | `./migrations` | Path to migration files relative to the backend working dir |

### JWT

| Variable | Required | Default | Description |
|---|---|---|---|
| `JWT_SECRET` | Yes | — | 32-byte hex secret for signing tokens |
| `JWT_ACCESS_TOKEN_TTL` | No | `15m` | Access-token lifetime (Go duration string) |
| `JWT_REFRESH_TOKEN_TTL` | No | `168h` | Refresh-token lifetime (Go duration string) |

### Google OAuth

Leave all three blank to disable Google sign-in entirely.

| Variable | Required | Default | Description |
|---|---|---|---|
| `GOOGLE_CLIENT_ID` | No | — | OAuth client ID |
| `GOOGLE_CLIENT_SECRET` | No | — | OAuth client secret |
| `GOOGLE_CALLBACK_URL` | No | `http://localhost:3000/api/auth/google/callback` | OAuth redirect URI |

### SMTP

Leave `SMTP_HOST` blank to disable email sending — OTP codes are logged to stdout instead (useful for local dev).

| Variable | Required | Default | Description |
|---|---|---|---|
| `SMTP_HOST` | No | — | SMTP server host |
| `SMTP_PORT` | No | `587` | SMTP server port |
| `SMTP_USERNAME` | No | — | SMTP username / email address |
| `SMTP_PASSWORD` | No | — | SMTP password or app password |
| `SMTP_FROM` | No | `FitProof <noreply@fitproof.app>` | From-address used on outgoing mail |

### R2 Storage

Leave all five blank to disable media uploads. The backend will boot normally, but upload routes will reject requests.

| Variable | Required | Default | Description |
|---|---|---|---|
| `R2_ACCOUNT_ID` | No | — | Cloudflare account ID |
| `R2_ACCESS_KEY_ID` | No | — | R2 API token access-key ID |
| `R2_SECRET_ACCESS_KEY` | No | — | R2 API token secret |
| `R2_BUCKET` | No | — | R2 bucket name |
| `R2_PUBLIC_URL` | No | — | Public bucket URL (e.g. `https://pub-xxx.r2.dev`, no trailing slash) |

### Other

| Variable | Required | Default | Description |
|---|---|---|---|
| `FRONTEND_URL` | No | `http://localhost:3000` | Used by backend to redirect after OAuth |
| `EMAIL_VERIFICATION_TTL` | No | `15m` | OTP lifetime (Go duration string) |

### Google OAuth setup

1. Go to [console.cloud.google.com](https://console.cloud.google.com) → APIs & Services → Credentials
2. Create an OAuth 2.0 Client ID (Web application)
3. Add `http://localhost:3000/api/auth/google/callback` as an authorised redirect URI
4. Copy the client ID and secret into `backend/.env`

### Email setup (Gmail)

1. Enable 2FA on your Google account
2. Go to [myaccount.google.com/apppasswords](https://myaccount.google.com/apppasswords) and create an app password
3. Set `SMTP_HOST=smtp.gmail.com`, `SMTP_PORT=587`, `SMTP_USERNAME=you@gmail.com`, `SMTP_PASSWORD=<app-password>`

If `SMTP_HOST` is left blank, the server logs OTP codes to stdout instead of sending emails — useful for local development.

## Development Commands

```bash
# Backend
cd backend
go run ./cmd/server        # Run server
go build ./cmd/server      # Build binary
go test ./...              # Run all tests
go vet ./...               # Static analysis
sqlc generate              # Regenerate DB code from SQL queries

# Migrations (manual — backend auto-applies on startup)
migrate -path ./migrations -database $DATABASE_URL up
migrate -path ./migrations -database $DATABASE_URL down 1

# Frontend
cd frontend
npm run dev                # Dev server (port 3000)
npm run build              # Production build
```

## Docker

Run the full stack (backend + frontend + PostgreSQL):

```bash
docker-compose up --build
```

The backend requires `JWT_SECRET` to be set as an environment variable or in a `.env` file at the project root for docker-compose to pick it up.

## Deployment

The app runs across three managed services. `main` is the live branch — both Render and Vercel deploy from it.

| Component | Host | URL |
|---|---|---|
| Backend (Go API) | Render | https://fitproof.onrender.com |
| Frontend (Next.js) | Vercel | https://fitproof-six.vercel.app |
| Database | Neon (us-east-1) | — |

### Media & retention

Submission photos and videos are stored in Cloudflare R2 and exposed via the public bucket URL configured in `R2_PUBLIC_URL`. The daily 00:05 UTC cron deletes `submission_media` rows older than 7 days; the corresponding R2 objects are expired by an R2 lifecycle rule (configured manually in the Cloudflare dashboard, not by the app). New contributors should expect old media references to disappear from the database on this schedule — that is intended.

## API Overview

| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/health` | — | Health check (DB ping) |
| GET | `/ping` | — | Liveness probe |
| POST | `/auth/register` | — | Register with email + password |
| POST | `/auth/login` | — | Login, returns JWT cookies |
| POST | `/auth/refresh` | — | Refresh access token |
| POST | `/auth/logout` | — | Revoke refresh token |
| GET | `/auth/google` | — | Redirect to Google OAuth |
| GET | `/auth/google/callback` | — | Google OAuth callback |
| POST | `/auth/verify/send` | JWT | Send email verification OTP |
| POST | `/auth/verify` | JWT | Verify email with OTP |
| GET | `/me` | JWT | Get current user |
| GET | `/me/stats` | JWT | Get user stats |
| GET | `/metrics` | JWT | List available metrics |
| POST | `/challenges` | JWT | Create a challenge |
| GET | `/challenges` | JWT | List joined challenges |
| POST | `/challenges/join` | JWT | Join via invite code |
| GET | `/challenges/:id` | JWT | Get challenge detail |
| POST | `/challenges/:id/metrics` | JWT | Add metrics to challenge |
| POST | `/challenges/:id/submissions` | JWT | Submit daily proof |
| GET | `/challenges/:id/submissions` | JWT | List submissions for challenge |
| POST | `/challenges/:id/submissions/:subID/media` | JWT | Upload media for a submission |
| GET | `/challenges/:id/feed` | JWT | Get challenge activity feed |
| GET | `/challenges/:id/leaderboard` | JWT | Get leaderboard |
| GET | `/challenges/:id/fines-summary` | JWT | Get fines summary for challenge |
| PATCH | `/challenges/:id/status` | JWT | Update challenge status |
| POST | `/challenges/:id/leave` | JWT | Leave a challenge |
