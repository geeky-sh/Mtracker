#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

echo "==> Mtracker local setup"

# 1. Copy .env if it doesn't exist
if [ ! -f .env ]; then
  cp .env.example .env
  echo "    Created .env from .env.example — fill in your Google credentials before starting."
fi

# 2. Check required tools
for cmd in docker docker-compose node npm go; do
  if ! command -v "$cmd" &>/dev/null; then
    echo "ERROR: '$cmd' is not installed or not in PATH." >&2
    exit 1
  fi
done

echo "    All required tools found."

# 3. Download Go dependencies
echo "==> Downloading Go modules…"
(cd apps/api && go mod download)

# 4. Install mobile dependencies
echo "==> Installing mobile dependencies…"
(cd apps/mobile && npm install)

echo ""
echo "Setup complete!"
echo ""
echo "Next steps:"
echo "  1. Edit .env and set JWT_SECRET, GOOGLE_CLIENT_ID, and the Expo vars."
echo "  2. Start the backend:  docker-compose up -d"
echo "  3. Start the mobile:   cd apps/mobile && npx expo start"
echo "  4. Open Expo Go on your Android device/emulator and scan the QR code."
