#!/usr/bin/env bash
set -euo pipefail

# Generate docs via mise task
mise run doc

# Check for uncommitted changes in docs/
if [ -n "$(git status --porcelain docs/)" ]; then
  echo "::error::Documentation is not up to date. Please run 'mise run doc' and commit the changes."
  git status docs/
  exit 1
fi

echo "Documentation is up to date."
