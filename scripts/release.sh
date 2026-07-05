#!/bin/bash
# Tag and prepare a new release.
# Usage: ./scripts/release.sh 1.2.3

set -e

VERSION=$1

if [ -z "$VERSION" ]; then
  echo "Usage: $0 <version>"
  echo "Example: $0 1.2.3"
  exit 1
fi

TAG="v$VERSION"

cd "$(dirname "$0")/.."

# Ensure working tree is clean
if [ -n "$(git status --porcelain)" ]; then
  echo "ERROR: Working tree is not clean. Commit or stash changes first."
  exit 1
fi

echo "Building release $TAG..."
bash scripts/build.sh

echo "Staging dist binaries..."
git add dist/mac/SyncApp dist/windows/SyncApp.exe dist/mac/templates dist/windows/templates

echo "Committing..."
git commit -m "Release $TAG"

echo "Tagging $TAG..."
git tag "$TAG"

echo ""
echo "Release $TAG ready. Push with:"
echo "  git push origin main && git push origin $TAG"
