#!/bin/bash
set -e

# This script checks that the versions in different parts of the repository are in sync

# Extract version from Git tag (if in CI environment and it's a tag)
if [[ -n "$GITHUB_REF" && "$GITHUB_REF" == refs/tags/v* ]]; then
  # This is a tag-based run, extract version from the tag
  GIT_VERSION=${GITHUB_REF#refs/tags/v}
  echo "Running in tag context: $GITHUB_REF"
else
  # Not a tag or not in CI, use the latest tag or fallback to package.json
  if git describe --tags --abbrev=0 2>/dev/null; then
    GIT_VERSION=$(git describe --tags --abbrev=0 2>/dev/null | sed 's/^v//' || echo "0.0.0")
    echo "Using latest Git tag: v$GIT_VERSION"
  else
    # No tags found, don't fail the check in PR builds or branch builds
    # Just use NPM version for comparison to itself (will always match)
    echo "No Git tags found, using NPM package version as reference"
    GIT_VERSION=$(node -p "require('./npm/package.json').version")
  fi
fi

# Extract version from npm package.json
NPM_VERSION=$(node -p "require('./npm/package.json').version")

# Extract version from goreleaser config (if needed)
# GORELEASER_VERSION=$(grep -oP 'version: \K.*' .goreleaser.yml || echo "")

# Check if versions match
echo "Git tag version: $GIT_VERSION"
echo "NPM package version: $NPM_VERSION"

# Determine if we're in a PR or branch build vs a release build
if [[ "$GITHUB_REF" == refs/heads/* || "$GITHUB_REF" == refs/pull/* ]]; then
  # For branch or PR builds, mismatch is allowed but we should show a warning
  if [ "$GIT_VERSION" != "$NPM_VERSION" ]; then
    echo "Warning: Version mismatch in branch/PR build"
    echo "Git version: $GIT_VERSION"
    echo "NPM version: $NPM_VERSION"
    echo "This will need to be fixed before release."
    # Exit with success for PRs and branches
    exit 0
  fi
else
  # For tag builds (releases), versions must match exactly
  if [ "$GIT_VERSION" != "$NPM_VERSION" ]; then
    echo "Error: Version mismatch in release build!"
    echo "Git version: $GIT_VERSION"
    echo "NPM version: $NPM_VERSION"
    echo "Versions must match exactly for releases."
    exit 1
  fi
fi

echo "✓ Versions are in sync"
exit 0