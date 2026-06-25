#!/usr/bin/env bash
set -euo pipefail
ROOT="$(dirname "$0")/.."
RECORDER_ASSETS="$ROOT/../api-recorder/src/assets"
WEB_ASSETS="$ROOT/web/public/assets"

if [[ -d "$RECORDER_ASSETS" ]]; then
  mkdir -p "$WEB_ASSETS"
  cp "$RECORDER_ASSETS"/icon16.png "$RECORDER_ASSETS"/icon48.png "$RECORDER_ASSETS"/icon128.png "$WEB_ASSETS"/
fi

cd "$ROOT/web"
if [[ ! -d node_modules ]]; then
  npm install
fi
npm run build
echo "✅ Web UI built → web/dist"
