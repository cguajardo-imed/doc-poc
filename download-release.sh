#!/bin/sh
# download-release.sh — Downloads the latest summarizer-api release from GitHub.
#
# Required env vars (or set below):
#   GH_TOKEN   — GitHub personal access token (repo scope)
#   GH_OWNER   — GitHub org or user (e.g. my-org)
#   GH_REPO    — Repository name (e.g. summarizer)
#
# Optional:
#   GH_TAG     — Specific tag to download (default: latest release)
#   INSTALL_DIR — Directory to place the binary (default: current directory)

set -eu

# ── Configuration ────────────────────────────────────────────────────────────

GH_TOKEN="${GH_TOKEN:-}"
GH_OWNER="${GH_OWNER:-cguajardo-imed}"
GH_REPO="${GH_REPO:-doc-poc}"
GH_TAG="${GH_TAG:-}"          # empty = latest
INSTALL_DIR="${INSTALL_DIR:-.}"

# Load .env if present and vars are still missing
if [ -f .env ] && { [ -z "$GH_TOKEN" ] || [ -z "$GH_OWNER" ] || [ -z "$GH_REPO" ]; }; then
  set -a
  # shellcheck disable=SC1091
  . ./.env
  set +a
fi

# ── Validation ────────────────────────────────────────────────────────────────

die() { echo "[download] ERROR: $*" >&2; exit 1; }

[ -n "$GH_TOKEN" ] || die "GH_TOKEN is not set."
[ -n "$GH_OWNER" ] || die "GH_OWNER is not set."
[ -n "$GH_REPO"  ] || die "GH_REPO is not set."

command -v curl >/dev/null 2>&1 || die "'curl' is required but not installed."
command -v jq   >/dev/null 2>&1 || die "'jq' is required but not installed."

# ── Detect OS / arch ─────────────────────────────────────────────────────────

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
  aarch64 | arm64) ARCH="arm64" ;;
  x86_64)          ARCH="amd64" ;;
  *) die "Unsupported architecture: $ARCH" ;;
esac

case "$OS" in
  linux)  EXT="" ;;
  darwin) EXT="" ;;
  mingw* | msys* | cygwin* | windows*) OS="windows"; EXT=".exe" ;;
  *) die "Unsupported OS: $OS" ;;
esac

ASSET_NAME="summarizer-api-${OS}-${ARCH}${EXT}"

echo "[download] Target asset: $ASSET_NAME"

# ── Resolve tag ───────────────────────────────────────────────────────────────

AUTH_HEADER="Authorization: Bearer $GH_TOKEN"
ACCEPT_JSON="Accept: application/vnd.github+json"
API_BASE="https://api.github.com/repos/$GH_OWNER/$GH_REPO"

if [ -z "$GH_TAG" ]; then
  echo "[download] Fetching latest release tag..."
  GH_TAG=$(
    curl -sf \
      -H "$AUTH_HEADER" \
      -H "$ACCEPT_JSON" \
      "$API_BASE/releases/latest" |
    jq -r '.tag_name'
  )
  [ -n "$GH_TAG" ] && [ "$GH_TAG" != "null" ] || die "Could not determine latest release tag."
fi

echo "[download] Release tag: $GH_TAG"

# ── Resolve asset ID ──────────────────────────────────────────────────────────

echo "[download] Resolving asset ID..."
ASSET_ID=$(
  curl -sf \
    -H "$AUTH_HEADER" \
    -H "$ACCEPT_JSON" \
    "$API_BASE/releases/tags/$GH_TAG" |
  jq -r --arg name "$ASSET_NAME" '.assets[] | select(.name == $name) | .id'
)

[ -n "$ASSET_ID" ] && [ "$ASSET_ID" != "null" ] || \
  die "Asset '$ASSET_NAME' not found in release '$GH_TAG'. Check that the build completed."

echo "[download] Asset ID: $ASSET_ID"

# ── Download ──────────────────────────────────────────────────────────────────

mkdir -p "$INSTALL_DIR"
DEST="$INSTALL_DIR/summarizer-api${EXT}"

echo "[download] Downloading to $DEST ..."
curl -fL \
  -H "$AUTH_HEADER" \
  -H "Accept: application/octet-stream" \
  "$API_BASE/releases/assets/$ASSET_ID" \
  -o "$DEST"

chmod +x "$DEST"

echo "[download] Done. Binary saved to: $DEST"
echo "[download] Run with: ./start.sh $DEST"
