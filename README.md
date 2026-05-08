# Mtracker

A personal activity tracker — log what you do, see when you did it.

```
apps/
  api/        Go (Gin) REST API
  mobile/     React Native (Expo) Android app
k8s/          Kubernetes manifests
scripts/      Helper scripts
docker-compose.yml
```

---

## Features

| Feature | Description |
|---------|-------------|
| **Google Sign-In** | Sign up / log in with your Google account |
| **New Activity** | Create an activity (name + optional description). Colour is auto-assigned. Duplicate suggestions appear as you type. |
| **Track Activity** | Record that you performed an activity — for today, yesterday, or 2 days ago. One log per activity per day. |
| **Analytics** | Home screen calendar highlights every date an activity was logged, in its colour. |

---

## Tech Stack

| Layer | Choice |
|-------|--------|
| Mobile | React Native + Expo (Android first, iOS-ready) |
| Backend | Go 1.22 + Gin |
| Database | PostgreSQL 16 |
| Auth | Google OAuth 2.0 + JWT |
| Local infra | Docker Compose |
| Production | Kubernetes |

---

## Local Setup (Quick Start)

### Prerequisites

| Tool | Minimum version |
|------|----------------|
| Go | 1.22 |
| Node.js | 20 LTS |
| Docker + Docker Compose | latest |
| Expo Go (Android) | Install from Play Store on your device / emulator |

### 1 — Google OAuth credentials

You need **two** OAuth clients from [Google Cloud Console](https://console.cloud.google.com/apis/credentials):

**a) Web Application client** (used by the backend for token verification)
- Go to *Create Credentials → OAuth Client ID → Web application*
- No redirect URIs needed
- Copy the **Client ID** → this becomes `GOOGLE_CLIENT_ID` in `.env`

**b) Android client** (used by the Expo app)
- Go to *Create Credentials → OAuth Client ID → Android*
- Package name: `com.aash.mtracker`
- SHA-1 fingerprint: use `keytool -list -v -keystore ~/.android/debug.keystore -alias androiddebugkey -storepass android -keypass android`
- Copy the **Client ID** → becomes `EXPO_PUBLIC_GOOGLE_ANDROID_CLIENT_ID` in `apps/mobile/.env`

> You can also set `EXPO_PUBLIC_GOOGLE_WEB_CLIENT_ID` to the same Web client ID as a fallback — Expo's OAuth flow will open a browser tab which works everywhere during development.

### 2 — Configure environment

```bash
# Run the automated setup (checks tools, installs deps, creates .env)
./scripts/setup-local.sh
```

Edit `.env` (backend):

```dotenv
JWT_SECRET=a_long_random_secret_here
GOOGLE_CLIENT_ID=<your_web_client_id>.apps.googleusercontent.com
```

Create `apps/mobile/.env`:

```dotenv
EXPO_PUBLIC_API_URL=http://10.0.2.2:8080          # Android emulator
# EXPO_PUBLIC_API_URL=http://192.168.x.x:8080     # Physical device — use your machine's LAN IP
EXPO_PUBLIC_GOOGLE_ANDROID_CLIENT_ID=<android_client_id>.apps.googleusercontent.com
EXPO_PUBLIC_GOOGLE_WEB_CLIENT_ID=<web_client_id>.apps.googleusercontent.com
```

> **Android emulator**: `10.0.2.2` is the loopback alias for the host machine.
> **Physical device**: use your machine's local IP. Find it with `ipconfig getifaddr en0` (macOS) or `hostname -I` (Linux).

### 3 — Start the backend

```bash
docker-compose up -d
```

This starts:
- `mtracker-postgres` on port `5432`
- `mtracker-api` on port `8080`

Migrations run automatically on first boot.

Verify: `curl http://localhost:8080/health` → `{"status":"ok"}`

### 4 — Start the mobile app

```bash
cd apps/mobile
npm install
npx expo start
```

- **Android emulator**: press `a` in the terminal
- **Physical device**: open Expo Go on your phone → scan the QR code

---

## API Reference

All protected endpoints require `Authorization: Bearer <jwt>`.

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `POST` | `/api/v1/auth/google` | No | Exchange Google access token for JWT |
| `GET` | `/api/v1/profile` | Yes | Get logged-in user profile |
| `GET` | `/api/v1/activities` | Yes | List all activities |
| `POST` | `/api/v1/activities` | Yes | Create activity |
| `GET` | `/api/v1/activities/search?q=` | Yes | Find similar activities |
| `GET` | `/api/v1/logs?activity_id=` | Yes | Get all logs for an activity |
| `POST` | `/api/v1/logs` | Yes | Log an activity for a date |
| `DELETE` | `/api/v1/logs/:id` | Yes | Remove a log entry |
| `GET` | `/health` | No | Health check |

### POST `/api/v1/auth/google`
```json
{ "access_token": "<google_oauth_access_token>" }
```
Response:
```json
{ "token": "<jwt>", "user": { "id": "...", "email": "...", "name": "...", "avatar_url": "..." } }
```

### POST `/api/v1/activities`
```json
{ "name": "Morning run", "description": "Optional note" }
```

### POST `/api/v1/logs`
```json
{ "activity_id": "<uuid>", "logged_date": "2026-04-22" }
```
`logged_date` must be today, yesterday, or 2 days ago (UTC). Duplicate logs for the same activity + date are rejected with `409 Conflict`.

---

## Database Schema

```
users
  id UUID PK, email, google_id, name, avatar_url, created_at, updated_at

activities
  id UUID PK, user_id FK→users, name, description, color, created_at, updated_at

activity_logs
  id UUID PK, activity_id FK→activities, user_id FK→users,
  logged_date DATE, created_at
  UNIQUE(activity_id, logged_date)
```

---

## Colour Palette

Activities are assigned one of 12 colours in rotation (by creation count, cycled):

`#FF6B6B` `#4ECDC4` `#45B7D1` `#96CEB4` `#F7DC6F` `#DDA0DD`
`#98D8C8` `#F0B27A` `#BB8FCE` `#85C1E9` `#82E0AA` `#F1948A`

---

## Backend Development (without Docker)

```bash
cd apps/api
go mod download
# Set env vars in your shell or via .env at repo root
go run ./cmd/server
```

---

## Production — Kubernetes

### One-time secret setup

```bash
cp k8s/postgres/secret.example.yaml k8s/postgres/secret.yaml
cp k8s/api/secret.example.yaml      k8s/api/secret.yaml
# Fill in base64 values, then apply
```

### Deploy

```bash
./scripts/k8s-deploy.sh ghcr.io/<your-username>/mtracker-api latest
```

### Access in-cluster

```bash
kubectl port-forward svc/mtracker-api 8080:80 -n mtracker
```

---

## Project Structure

```
Mtracker/
├── apps/
│   ├── api/
│   │   ├── cmd/server/main.go          Entry point
│   │   └── internal/
│   │       ├── config/                 Env-based config loader
│   │       ├── database/               GORM connection + AutoMigrate
│   │       ├── models/                 User, Activity, ActivityLog
│   │       ├── handlers/               auth.go  activities.go  logs.go
│   │       ├── middleware/             JWT auth middleware
│   │       └── router/                 Route registration
│   └── mobile/
│       ├── App.tsx                     Root — auth gate
│       └── src/
│           ├── navigation/             Bottom tab navigator
│           ├── screens/                Login, Home (analytics), New, Track
│           ├── components/             ActivityCard, ColorBadge
│           ├── services/               api.ts (Axios)  auth.ts (Expo OAuth)
│           ├── types/                  Shared TypeScript types
│           └── utils/                  colours helper, useDebounce
├── k8s/
│   ├── namespace.yaml
│   ├── postgres/                       PVC, Deployment, Service, Secret
│   └── api/                           Deployment, Service, ConfigMap, Secret
├── scripts/
│   ├── setup-local.sh
│   └── k8s-deploy.sh
└── docker-compose.yml
```

---

## Troubleshooting

| Problem | Fix |
|---------|-----|
| `connection refused` on Android emulator | Use `10.0.2.2` not `localhost` in `EXPO_PUBLIC_API_URL` |
| Google sign-in fails | Check client IDs and that the correct package name / SHA-1 is registered in Google Cloud |
| `required environment variable … is not set` | Fill in `.env` before `docker-compose up` |
| `409 Conflict` on log creation | That activity is already logged for that day |
| Port 5432 already in use | Stop local Postgres or change the port mapping in `docker-compose.yml` |
| Expo QR not loading on device | Make sure the device is on the same Wi-Fi as your machine |
