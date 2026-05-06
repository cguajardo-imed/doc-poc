#!/bin/sh
# build.sh — Builds the summarizer-api binary using Docker (no local Go required).
#
# The binary is extracted from the Docker build stage directly,
# so you get the exact same artifact that would run in the container.
#
# Usage:
#   ./build.sh                        # native OS/arch
#   GOARCH=arm64 ./build.sh           # cross-compile (e.g. for EC2 aarch64)
#   GOOS=linux GOARCH=arm64 ./build.sh
#
# Output:
#   dist/summarizer-api-<os>-<arch>[.exe]

set -eu

GOOS="${GOOS:-linux}"
GOARCH="${GOARCH:-$(uname -m)}"
IMAGE_TAG="summarizer-api-builder:local"
DIST_DIR="dist"

# Normalise arch naming to match Go conventions
case "$GOARCH" in
  aarch64) GOARCH="arm64" ;;
  x86_64)  GOARCH="amd64" ;;
esac

EXT=""
[ "$GOOS" = "windows" ] && EXT=".exe"

BINARY_NAME="summarizer-api-${GOOS}-${GOARCH}${EXT}"

echo "[build] Target: GOOS=$GOOS  GOARCH=$GOARCH"
echo "[build] Output: $DIST_DIR/$BINARY_NAME"

# ── Build the Docker image (builder stage only) ───────────────────────────────

docker build \
  --target builder \
  --build-arg GOOS="$GOOS" \
  --build-arg GOARCH="$GOARCH" \
  --build-arg CGO_ENABLED=0 \
  -t "$IMAGE_TAG" \
  -f api.Dockerfile \
  ./api

# ── Extract binary from the image ────────────────────────────────────────────

mkdir -p "$DIST_DIR"

CONTAINER_ID=$(docker create "$IMAGE_TAG")
docker cp "$CONTAINER_ID:/app/summarizer-api" "$DIST_DIR/$BINARY_NAME"
docker rm "$CONTAINER_ID" >/dev/null

chmod +x "$DIST_DIR/$BINARY_NAME"

echo "[build] Done → $DIST_DIR/$BINARY_NAME"
