#!/bin/bash
set -e

# Ensure we're on main
BRANCH=$(git rev-parse --abbrev-ref HEAD)
if [[ "$BRANCH" != "main" ]]; then
  echo "Error: You must be on the main branch to release."
  exit 1
fi

# Ensure working directory is clean
if [[ -n $(git status -s) ]]; then
  echo "Error: Working directory is not clean. Commit or stash changes first."
  exit 1
fi

# Run checks
echo "Running release checks..."
make release-check

# Check for outdated dependencies
echo "Checking for outdated dependencies..."
./scripts/check_updates.sh
UPDATES=$(go list -u -m -f '{{if .Update}}{{.Path}}{{end}}' all)
if [[ -n "$UPDATES" ]]; then
    echo ""
    read -p "The above dependencies have updates available. Proceed anyway? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Release cancelled. Please update dependencies and try again."
        exit 1
    fi
fi

# Get current version
CURRENT_VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
echo "Current version: $CURRENT_VERSION"

# Ask for bump type
echo "Select release type:"
select type in "patch" "minor" "major"; do
    case $type in
        patch|minor|major) BUMP_TYPE=$type; break;;
        *) echo "Invalid selection";;
    esac
done

# Calculate new version
# Remove 'v' prefix
VERSION_NUM=${CURRENT_VERSION#v}
IFS='.' read -r MAJOR MINOR PATCH <<< "$VERSION_NUM"

if [[ "$BUMP_TYPE" == "major" ]]; then
    MAJOR=$((MAJOR + 1))
    MINOR=0
    PATCH=0
elif [[ "$BUMP_TYPE" == "minor" ]]; then
    MINOR=$((MINOR + 1))
    PATCH=0
elif [[ "$BUMP_TYPE" == "patch" ]]; then
    PATCH=$((PATCH + 1))
fi

NEW_VERSION="v$MAJOR.$MINOR.$PATCH"
echo "New version will be: $NEW_VERSION"

# Confirm
read -p "Are you sure you want to release $NEW_VERSION? (y/N) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Release cancelled."
    exit 1
fi

# Create tag and push
echo "Tagging release $NEW_VERSION..."
git tag -a "$NEW_VERSION" -m "Release $NEW_VERSION"
git push origin "$NEW_VERSION"

echo "Release $NEW_VERSION triggered! Check GitHub Actions for progress."
