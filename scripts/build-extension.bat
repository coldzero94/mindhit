@echo off
echo 🚀 Building MindHit Chrome Extension...
echo.

REM Check if pnpm is installed
where pnpm >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo ❌ pnpm is not installed
    echo 📦 Installing pnpm...
    call npm install -g pnpm
)

REM Check if moon is installed
where moon >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo ❌ moon task runner is not installed
    echo 📦 Installing moon...
    call npm install -g @moonrepo/cli
)

REM Check if .env exists
if not exist .env (
    echo ⚙️  No .env file found. Creating from .env.example...
    copy .env.example .env
    echo ⚠️  Please edit .env and set your API_SECRET_KEY, VITE_API_KEY, and other values
    echo.
)

REM Install dependencies if needed
if not exist node_modules (
    echo 📦 Installing dependencies...
    call pnpm install
)

REM Build extension
echo 🔨 Building extension...
call pnpm run --filter @mindhit/extension build

echo.
echo ✅ Extension built successfully!
echo.
echo 📁 Extension location: apps\extension\dist\
echo.
echo 📖 To install in Chrome:
echo    1. Open chrome://extensions
echo    2. Enable 'Developer mode'
echo    3. Click 'Load unpacked'
echo    4. Select: %CD%\apps\extension\dist
echo.
