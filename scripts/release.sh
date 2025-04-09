#!/bin/bash
set -e

# Check if a version is provided
if [ -z "$1" ]; then
  echo "Usage: $0 <version>"
  echo "Example: $0 0.1.0"
  exit 1
fi

VERSION=$1
TAG="v$VERSION"

# Check if the current directory is the root of the repo
if [ ! -f "go.mod" ]; then
  echo "Error: This script must be run from the root of the repository"
  exit 1
fi

# Ensure the working directory is clean
if [ -n "$(git status --porcelain)" ]; then
  echo "Error: Working directory is not clean. Commit or stash changes first."
  exit 1
fi

echo "Creating release $TAG..."

# Update version in the version.go file
echo "Updating version in pkg/version/version.go..."
sed -i '' "s/Version = \".*\"/Version = \"$VERSION\"/g" pkg/version/version.go
git_commit=$(git rev-parse --short HEAD)
build_date=$(date -u +%Y-%m-%dT%H:%M:%SZ)
sed -i '' "s/GitCommit = \".*\"/GitCommit = \"$git_commit\"/g" pkg/version/version.go
sed -i '' "s/BuildDate = \".*\"/BuildDate = \"$build_date\"/g" pkg/version/version.go

# Update CHANGELOG.md
echo "Updating CHANGELOG.md..."
DATE=$(date +%Y-%m-%d)
sed -i '' "s/## \[Unreleased\]/## \[Unreleased\]\n\n## \[$VERSION\] - $DATE/g" CHANGELOG.md

# Commit the version changes
echo "Committing version changes..."
git add pkg/version/version.go CHANGELOG.md
git commit -m "chore: prepare release $TAG"

# Create and push the tag
echo "Creating and pushing tag $TAG..."
git tag -a "$TAG" -m "Release $TAG"
git push origin main
git push origin "$TAG"

echo "Release $TAG created and pushed!"
echo "The GitHub Actions workflow will now build and create the release."
echo "Check the progress at: https://github.com/dalekurt/kube-ai/actions" 