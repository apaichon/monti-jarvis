#!/usr/bin/env bash
set -euo pipefail

# Thin entrypoint kept with Monti so the checkout being deployed is explicit.
# The host-level Docker, compose, and Nginx orchestration remains owned by the
# HarvestMax deployment repository.

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
DEPLOYMENT_ROOT="${HARVEST_DEPLOYMENT_ROOT:-/Users/apaichon/Projects/libra/harvest-max/harvest-deployment}"

if [[ ! -x "$DEPLOYMENT_ROOT/scripts/deploy-monti.sh" ]]; then
    echo "ERROR: HarvestMax deploy script not found or not executable: $DEPLOYMENT_ROOT/scripts/deploy-monti.sh" >&2
    echo "Set HARVEST_DEPLOYMENT_ROOT to the host deployment checkout." >&2
    exit 1
fi

exec "$DEPLOYMENT_ROOT/scripts/deploy-monti.sh" \
    --source "$ROOT_DIR" \
    "$@"
