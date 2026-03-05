# FitProof

A fitness accountability app where users create and join challenges, submit daily proof of their workouts, and get fined for missed days.

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
- Leaderboard based on pre-aggregated points and fines
- JWT auth with refresh tokens (HTTP-only cookies, per-device sessions)
- Google OAuth sign-in
- Email verification via 6-digit OTP
- Automated missed-submission cron (runs daily at 00:05 UTC)

## Tech Stack

| Layer | Tech |
|---|---|
| Backend | Go 1.24, chi v5, pgx v5, sqlc, golang-migrate |
| Database | PostgreSQL 16 |
| Auth | JWT (golang-jwt/jwt/v5), bcrypt, Google OAuth2 |
| Email | go-mail (SMTP) |
| Frontend | Next.js 16, React 19, Tailwind CSS v4, TypeScript |
| Deployment | Docker, docker-compose |

## Getting Started

### Prerequisites

- Go 1.24+
- Node.js 20+
- PostgreSQL 16
- [golang-migrate](https://github.com/golang-migrate/migrate) CLI
- [sqlc](https://sqlc.dev) (only needed if modifying SQL queries)

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

All backend configuration lives in `backend/.env`. Copy from `backend/.env.example`.

| Variable | Required | Description |
|---|---|---|
| `DATABASE_URL` | Yes | PostgreSQL connection string (`pgx5://...`) |
| `JWT_SECRET` | Yes | 32-byte hex secret for signing tokens |
| `GOOGLE_CLIENT_ID` | No | Google OAuth client ID (disables Google sign-in if blank) |
| `GOOGLE_CLIENT_SECRET` | No | Google OAuth client secret |
| `GOOGLE_CALLBACK_URL` | No | OAuth redirect URI (default: `http://localhost:3000/api/auth/google/callback`) |
| `SMTP_HOST` | No | SMTP server host (OTPs printed to console if blank) |
| `SMTP_USERNAME` | No | SMTP username / email address |
| `SMTP_PASSWORD` | No | SMTP password or app password |

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

# Migrations
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

## API Overview

| Method | Path | Auth | Description |
|---|---|---|---|
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
| GET | `/challenges/:id` | JWT | Get challenge detail |
| POST | `/challenges/:id/metrics` | JWT | Add metrics to challenge |
| POST | `/challenges/join` | JWT | Join via invite code |
| POST | `/challenges/:id/submissions` | JWT | Submit daily proof |
| GET | `/challenges/:id/leaderboard` | JWT | Get leaderboard |
