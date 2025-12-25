#!/bin/bash
set -e

echo "Generating all code..."
pnpm run generate

echo "Checking for uncommitted changes..."
if [[ -n $(git status --porcelain) ]]; then
  echo "❌ Generated files are out of sync!"
  git diff
  exit 1
fi

echo "✅ All generated files are up to date"
