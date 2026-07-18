#!/usr/bin/env bash
# Build all FEAT-0017 embed packages (core first).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")" && pwd)"
cd "$ROOT/embed-core"
npm install --no-fund --no-audit
npm test
for pkg in embed-vue embed-react embed-svelte embed-web-component; do
  echo "==> $pkg"
  cd "$ROOT/$pkg"
  npm install --legacy-peer-deps --no-fund --no-audit
  npm run build
done
echo "All embed packages built."
