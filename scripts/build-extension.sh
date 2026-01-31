#!/bin/bash
set -e

echo "🚀 Building MindHit Chrome Extension..."
echo ""

# Check if pnpm is installed
if ! command -v pnpm &> /dev/null; then
    echo "❌ pnpm is not installed"
    echo "📦 Installing pnpm..."
    npm install -g pnpm
fi

# Check if moon is installed
if ! command -v moon &> /dev/null; then
    echo "❌ moon task runner is not installed"
    echo "📦 Installing moon..."
    npm install -g @moonrepo/cli
fi

# Check if .env exists
if [ ! -f .env ]; then
    echo "⚙️  No .env file found. Creating from .env.example..."
    cp .env.example .env
    echo "⚠️  Please edit .env and set your API_SECRET_KEY, VITE_API_KEY, and other values"
    echo ""
fi

# Install dependencies if needed
if [ ! -d "node_modules" ]; then
    echo "📦 Installing dependencies..."
    pnpm install
fi

# Build extension
echo "🔨 Building extension..."
pnpm run --filter @mindhit/extension build

echo ""
echo "✅ Extension built successfully!"
echo ""
echo "📁 Extension location: apps/extension/dist/"
echo ""
echo "📖 To install in Chrome:"
echo "   1. Open chrome://extensions"
echo "   2. Enable 'Developer mode'"
echo "   3. Click 'Load unpacked'"
echo "   4. Select: $(pwd)/apps/extension/dist"
echo ""
