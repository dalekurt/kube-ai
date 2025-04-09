#!/bin/bash
set -e

# Check if version is provided
if [ -z "$1" ]; then
  echo "Usage: $0 <version> [extract-notes]"
  echo "Example: $0 0.1.0"
  echo "Example: $0 0.1.0 extract-notes (only extracts release notes from CHANGELOG.md)"
  exit 1
fi

VERSION=$1
TAG="v$VERSION"
EXTRACT_NOTES=${2:-""}

# Function to extract release notes from CHANGELOG.md
extract_changelog() {
  awk -v ver="$VERSION" '
    BEGIN { found=0; capturing=0; }
    $0 ~ "## \\[" ver "\\]" { found=1; capturing=1; next; }
    found && $0 ~ "^## " { capturing=0; }
    found && capturing { print $0; }
  ' CHANGELOG.md
}

# If extract-notes is specified, only output the changelog content and exit
if [ "$EXTRACT_NOTES" = "extract-notes" ]; then
  extract_changelog
  exit 0
fi

# Check if GitHub CLI is installed
if ! command -v gh &> /dev/null; then
  echo "GitHub CLI not found. Please install it first:"
  echo "  https://cli.github.com/manual/installation"
  exit 1
fi

# Check if the current directory is the root of the repo
if [ ! -f "go.mod" ]; then
  echo "Error: This script must be run from the root of the repository"
  exit 1
fi

# Check if the tag exists
if ! git rev-parse "$TAG" >/dev/null 2>&1; then
  echo "Error: Tag $TAG does not exist. Create it first with:"
  echo "  task release -- $VERSION"
  exit 1
fi

echo "Creating GitHub release for $TAG..."

# Extract release notes from CHANGELOG.md
echo "Extracting release notes from CHANGELOG.md..."
CHANGELOG_CONTENT=$(extract_changelog)

if [ -z "$CHANGELOG_CONTENT" ]; then
  echo "Error: Could not find release notes for version $VERSION in CHANGELOG.md"
  exit 1
fi

# Create a temporary file for the release notes
RELEASE_NOTES_FILE=$(mktemp)
cat > "$RELEASE_NOTES_FILE" << EOF
# Kube-AI $VERSION

$CHANGELOG_CONTENT

## Installation

### Linux (amd64)
\`\`\`bash
curl -L https://github.com/dalekurt/kube-ai/releases/download/$TAG/kube-ai-linux-amd64 -o kube-ai
chmod +x kube-ai
sudo mv kube-ai /usr/local/bin/
\`\`\`

### macOS (Apple Silicon)
\`\`\`bash
curl -L https://github.com/dalekurt/kube-ai/releases/download/$TAG/kube-ai-darwin-arm64 -o kube-ai
chmod +x kube-ai
sudo mv kube-ai /usr/local/bin/
\`\`\`

### Docker
\`\`\`bash
docker pull dalekurt/kube-ai:$VERSION
\`\`\`

See the [documentation](README.md) for more details.
EOF

echo "Building binaries for all platforms..."
if ! task build:all > /dev/null; then
  echo "Error: Failed to build binaries"
  rm "$RELEASE_NOTES_FILE"
  exit 1
fi

echo "Creating GitHub release..."
gh release create "$TAG" \
  --title "Kube-AI $VERSION" \
  --notes-file "$RELEASE_NOTES_FILE" \
  --draft \
  bin/*

echo "Building and pushing Docker image..."
if task docker:build && task docker:push; then
  echo "Docker image pushed successfully"
else
  echo "Warning: Failed to build or push Docker image"
fi

# Clean up
rm "$RELEASE_NOTES_FILE"

echo "Release draft created successfully!"
echo "Please review and publish the release at: https://github.com/dalekurt/kube-ai/releases" 