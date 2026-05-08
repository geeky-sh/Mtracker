# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What is Mtracker

Mtracker is a personal activity tracker. Users define activities (e.g. "Running", "Reading") and log them against specific dates. The app curates insights from the log history — a calendar view shows which days each activity was performed, and a pie chart breaks down relative effort across activities over a configurable time window.

## Commands

### Go API (run from `apps/api/`)

```bash
go run ./cmd/server                          # run server locally
go test ./internal/...                       # run all tests
go test ./internal/<pkg>/... -run <TestName> # run a single test
go test ./internal/... -coverprofile=coverage.out && go tool cover -func=coverage.out  # coverage
```

### Docker (run from repo root)

```bash
docker compose up -d                  # start postgres + api
docker compose up -d --build api      # rebuild and restart api only
```

Postgres is exposed on port **5433** (external), API on **8080**.

### Mobile (run from `apps/mobile/`)

```bash
npm install
npx expo start        # then press 'a' for Android emulator, 'i' for iOS simulator
```

## Architecture

### Monorepo layout

`apps/api` (Go REST API) + `apps/mobile` (Expo/React Native) + `k8s/` (production manifests). The root `.env` file is shared by docker-compose and loaded by the API via `godotenv`.

### Go API

Entry point: `cmd/server/main.go` → loads config → connects DB (with AutoMigrate) → calls `router.Setup()`. All logic lives under `internal/` in six packages: `config`, `database`, `models`, `middleware`, `handlers`, `router`.

**Auth flow**: `POST /api/v1/auth/login` accepts a Google OAuth access token, calls the Google userinfo endpoint to verify it, upserts the user row, and returns a signed HS256 JWT. All subsequent requests carry `Authorization: Bearer <token>`; `middleware.Auth` validates the token and writes `user_id` into the gin context.

**Activity colors**: assigned from a fixed 12-color palette cycled by `existingCount % 12` at creation time.

**Log constraint**: one log per activity per day, enforced by a unique DB index (`idx_activity_logs_activity_date`) and a 30-day look-back window validated in the Create handler.

### Mobile

- `services/api.ts` — axios client; JWT persisted via `expo-secure-store` and attached as a Bearer header on every request
- `services/auth.ts` — Expo OAuth session flow; on success calls the backend login endpoint and stores the JWT
- Navigation: bottom tab navigator (`AppNavigator.tsx`) with Home, New Activity, Track, and Profile tabs
- `HomeScreen` owns the calendar + pie chart view; it builds a `logsByDate` map (keyed `YYYY-MM-DD`) during data fetch, used for calendar dot rendering and log deletion

### Testing (Go)

- SQL mocking: `github.com/DATA-DOG/go-sqlmock v1.5.2` — import path is `github.com/DATA-DOG/go-sqlmock` (v1, not v2)
- Mock DB helper in `handlers/testhelper_test.go`: uses `SkipDefaultTransaction: true` and `MatchExpectationsInOrder(false)`
- GORM postgres issues **`ExecContext`** for DELETE → use `mock.ExpectExec(...)`, not `ExpectQuery`
- GORM postgres issues **`QueryContext`** (RETURNING clause) for INSERT → use `mock.ExpectQuery(...).WillReturnRows(...)`
- Package-level vars exist specifically for test injection: `openDialector` (`database` package), `jwtSigner` and `googleUserInfoURL` (`handlers` package) — override and restore via `t.Cleanup`
