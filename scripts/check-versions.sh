#!/bin/bash
set -e

# This script checks that the versions in different parts of the repository are in sync

# Extract version from Git tag (if in CI environment)
if [ -n "$GITHUB_REF" ]; then
  GIT_VERSION=${GITHUB_REF#refs/tags/v}
else
  # If not in CI, use the latest tag
  GIT_VERSION=$(git describe --tags --abbrev=0 2>/dev/null | sed 's/^v//' || echo "0.0.0")
fi

# Extract version from npm package.json
NPM_VERSION=$(node -p "require('./npm/package.json').version")

# Extract version from goreleaser config (if needed)
# GORELEASER_VERSION=$(grep -oP 'version: \K.*' .goreleaser.yml || echo "")

# Check if versions match
echo "Git tag version: $GIT_VERSION"
echo "NPM package version: $NPM_VERSION"

# Compare versions
if [ "$GIT_VERSION" != "$NPM_VERSION" ]; then
  echo "Error: Version mismatch!"
  echo "Git version: $GIT_VERSION"
  echo "NPM version: $NPM_VERSION"
  exit 1
fi

echo "✓ Versions are in sync"
exit 0