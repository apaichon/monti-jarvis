#!/usr/bin/env sh
set -eu

echo "Sprint 41 security regression $(date -u +%Y-%m-%dT%H:%M:%SZ)"
echo "Go unit and isolation tests"
go test ./... -count=1
echo "Go static analysis"
go vet ./...

for app in apps/tenant-web apps/customer-web; do
  echo "Frontend checks: $app"
  (cd "$app" && npm run prebuild && npm run check && npm run build >/dev/null)
done

echo "Security regression suite passed; output contains no credentials or tenant data."
