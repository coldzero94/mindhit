# Chrome Extension Installation

Build and install the MindHit Chrome extension from source.

## Quick Start (Recommended)

For the easiest installation, use the provided build script:

**macOS/Linux:**

```bash
./scripts/build-extension.sh
```

**Windows:**

```bash
scripts\build-extension.bat
```

The script will:

- Check and install pnpm if needed
- Create `.env` from `.env.example` if missing
- Install dependencies
- Build the extension
- Show you exactly where to find the built extension

After running the script, follow the instructions to load `apps/extension/dist/` in Chrome.

## Manual Installation

If you prefer to build manually or the script doesn't work, follow these steps:

### Prerequisites

- Node.js 22+
- Moon task runner (install with `npm install -g @moonrepo/cli`)

## Build from Source

### 1. Install Moon Task Runner

```bash
npm install -g @moonrepo/cli
```

### 2. Configure Environment

Make sure your root `.env` file has the extension variables set:

```bash
VITE_API_URL=http://localhost:9000/v1
VITE_WEB_APP_URL=http://localhost:3000
VITE_API_KEY=your-api-key
```

### 3. Build

```bash
moonx extension:build
```

This creates a `dist/` folder in `apps/extension/` with the built extension.

## Load in Chrome

1. Open Chrome and go to `chrome://extensions`
2. Enable **Developer mode** (toggle in top right)
3. Click **Load unpacked**
4. Select the `apps/extension/dist` folder

The MindHit extension icon should now appear in your browser toolbar.

## Usage

### Starting a Recording Session

1. Click the MindHit extension icon
2. Click **Start Recording**
3. Browse the web normally
4. Click **Stop** when done

### Viewing Your Data

1. Click **Dashboard** in the extension popup
2. Or go directly to http://localhost:3000 (or your web app URL)
3. View your sessions and generated mindmaps

## Development Mode

For development with hot reload:

```bash
moonx extension:dev
```

Load the `dist/` folder in Chrome as above. Changes will automatically rebuild.

## Troubleshooting

### Extension not connecting to API

1. Check that the backend API is running
2. Verify `VITE_API_URL` points to your API server
3. Check `VITE_API_KEY` matches your `API_SECRET_KEY`

### Build errors

1. Make sure you're using Node.js 22+
2. Run `moonx extension:build` again
3. Check for TypeScript errors: `moonx extension:typecheck`

### Extension not appearing

1. Make sure you loaded the `dist/` folder, not the `apps/extension` folder
2. Check for errors in `chrome://extensions`
3. Try reloading the extension

## Updating the Extension

After pulling new code:

```bash
moonx extension:build
```

Then reload the extension in `chrome://extensions`.
