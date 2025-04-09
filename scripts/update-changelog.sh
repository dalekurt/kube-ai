#!/bin/bash
set -e

# Get the current directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
CHANGELOG_FILE="$ROOT_DIR/CHANGELOG.md"
TODAY=$(date +%Y-%m-%d)

# Check if the CHANGELOG.md file exists
if [ ! -f "$CHANGELOG_FILE" ]; then
  echo "Error: CHANGELOG.md file not found at $CHANGELOG_FILE"
  exit 1
fi

# Function to extract the latest version from CHANGELOG.md
get_latest_version() {
  grep -E "^## \[[0-9]+\.[0-9]+\.[0-9]+\]" "$CHANGELOG_FILE" | head -n 1 | sed -E 's/## \[(.*)\].*/\1/'
}

# Function to get git commits since last version
get_commits_since_last_version() {
  local latest_version=$(get_latest_version)
  local latest_tag="v$latest_version"
  
  # Check if the tag exists
  if git rev-parse "$latest_tag" >/dev/null 2>&1; then
    # Get commits since the last tag
    git log --pretty=format:"- %s (%h)" "$latest_tag"..HEAD | grep -v "Merge pull request" | grep -v "Merge branch"
  else
    # If no tag exists, get all commits
    git log --pretty=format:"- %s (%h)" | grep -v "Merge pull request" | grep -v "Merge branch" | head -n 20
  fi
}

# Function to categorize commits based on conventional commit format
categorize_commits() {
  local commits="$1"
  local added=""
  local changed=""
  local fixed=""
  local other=""
  
  while IFS= read -r line; do
    if [[ "$line" == *"feat"* || "$line" == *"add"* ]]; then
      added+="$line"$'\n'
    elif [[ "$line" == *"fix"* || "$line" == *"bug"* ]]; then
      fixed+="$line"$'\n'
    elif [[ "$line" == *"change"* || "$line" == *"refactor"* || "$line" == *"chore"* ]]; then
      changed+="$line"$'\n'
    else
      other+="$line"$'\n'
    fi
  done <<< "$commits"
  
  # Prepare the output
  echo "### Added"
  if [ -n "$added" ]; then
    echo "$added"
  else
    echo "- No new features added in this release"
  fi
  
  echo ""
  echo "### Changed"
  if [ -n "$changed" ]; then
    echo "$changed"
  else
    echo "- No significant changes in this release"
  fi
  
  echo ""
  echo "### Fixed"
  if [ -n "$fixed" ]; then
    echo "$fixed"
  else
    echo "- No fixes in this release"
  fi
  
  # Add other commits if any
  if [ -n "$other" ]; then
    echo ""
    echo "### Other"
    echo "$other"
  fi
}

# Main function to update the CHANGELOG.md
update_changelog() {
  local version="$1"
  local is_major=0
  local is_minor=0
  local is_patch=1
  
  # Determine version type if not specified
  if [ -z "$version" ]; then
    # Get the latest version
    local latest_version=$(get_latest_version)
    IFS='.' read -r major minor patch <<< "$latest_version"
    
    # Check commit messages to determine version type
    local commits=$(get_commits_since_last_version)
    if echo "$commits" | grep -q -E "BREAKING|breaking"; then
      is_major=1
      is_minor=0
      is_patch=0
    elif echo "$commits" | grep -q -E "feat:|feature:|add:|added:"; then
      is_major=0
      is_minor=1
      is_patch=0
    fi
    
    # Calculate new version
    if [ "$is_major" -eq 1 ]; then
      version="$((major + 1)).0.0"
    elif [ "$is_minor" -eq 1 ]; then
      version="$major.$((minor + 1)).0"
    else
      version="$major.$minor.$((patch + 1))"
    fi
  fi
  
  # Get the commits and categorize them
  local commits=$(get_commits_since_last_version)
  local categorized=$(categorize_commits "$commits")
  
  # Create the new version section
  local new_section="## [$version] - $TODAY

$categorized"
  
  # Insert the new section after the Unreleased section
  awk -v new="$new_section" '
    /^## \[Unreleased\]/ {
      print $0
      print ""
      print new
      next
    }
    { print $0 }
  ' "$CHANGELOG_FILE" > "$CHANGELOG_FILE.new"
  
  mv "$CHANGELOG_FILE.new" "$CHANGELOG_FILE"
  
  echo "CHANGELOG.md updated with new version $version"
}

# Parse command-line arguments
VERSION=""
if [ $# -ge 1 ]; then
  VERSION="$1"
fi

update_changelog "$VERSION" 