#!/bin/sh
# start.sh — Starts summarizer-api and ollama serve in parallel.
# Usage: ./start.sh [path/to/summarizer-api]
#
# Environment variables (can be set in .env or exported before running):
#   API_BINARY   — path to the API binary (default: ./summarizer-api)
#   API_PORT     — port the API listens on (default: 3000)
#   OLLAMA_HOST  — ollama bind address (default: 0.0.0.0:11434)

set -eu

API_BINARY="${API_BINARY:-./summarizer-api}"
OLLAMA_HOST="${OLLAMA_HOST:-0.0.0.0:11434}"

# Load .env if present
if [ -f .env ]; then
  # Export variables, ignoring comments and blank lines
  set -a
  # shellcheck disable=SC1091
  . ./.env
  set +a
fi

# Validate API binary exists
if [ ! -f "$API_BINARY" ]; then
  echo "[start.sh] ERROR: API binary not found at '$API_BINARY'"
  echo "[start.sh] Set API_BINARY env var or pass the path as first argument."
  exit 1
fi

if [ -n "${1:-}" ]; then
  API_BINARY="$1"
fi

# Make sure it is executable
chmod +x "$API_BINARY"

cleanup() {
  echo ""
  echo "[start.sh] Shutting down..."
  # Kill the whole process group spawned by this script
  kill 0
}

trap cleanup INT TERM

echo "[start.sh] Starting ollama serve (OLLAMA_HOST=$OLLAMA_HOST)..."
OLLAMA_HOST="$OLLAMA_HOST" ollama serve &
OLLAMA_PID=$!

# Give ollama a moment to bind the port before the API tries to reach it
sleep 2

echo "[start.sh] Starting summarizer-api ($API_BINARY)..."
"$API_BINARY" &
API_PID=$!

echo "[start.sh] Running. PIDs: ollama=$OLLAMA_PID  api=$API_PID"
echo "[start.sh] Press Ctrl+C to stop both."

# Wait for either process to exit; exit code reflects the first failure
wait $OLLAMA_PID $API_PID
