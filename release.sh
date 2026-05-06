#!/usr/bin/env bash
set -euo pipefail

# Usage: ./release.sh [version]
# Example: ./release.sh v1.2.0
# If no version is provided, it increments the latest tag's patch number.

get_next_version() {
  local latest
  latest=$(git tag --sort=-v:refname | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$' | head -n1)
  if [[ -z "$latest" ]]; then
    echo "v0.1.0"
    return
  fi
  local major minor patch
  IFS='.' read -r major minor patch <<< "${latest#v}"
  echo "v${major}.${minor}.$((patch + 1))"
}

VERSION="${1:-}"

if [[ -z "$VERSION" ]]; then
  VERSION=$(get_next_version)
  echo "No version provided. Using next version: $VERSION"
fi

if ! [[ "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "Error: version must match vX.Y.Z (got: $VERSION)"
  exit 1
fi

if git tag | grep -qx "$VERSION"; then
  echo "Error: tag $VERSION already exists locally."
  exit 1
fi

echo "Creating and pushing tag $VERSION..."
git tag "$VERSION"
git push origin "$VERSION"

echo ""
echo "Release workflow triggered for $VERSION"
echo "Track it at: $(gh repo view --json url -q .url)/actions"
