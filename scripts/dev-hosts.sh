#!/usr/bin/env bash
# Manage local dev hostname for Monti Jarvis (macOS/Linux /etc/hosts).
set -euo pipefail

HOSTNAME="${DEV_HOSTNAME:-monti-jarvis-dev.local}"
IP="${DEV_HOST_IP:-127.0.0.1}"
MARKER="# monti-jarvis dev"
HOSTS_FILE="/etc/hosts"

usage() {
  cat <<EOF
Usage: $(basename "$0") <add|remove|status>

  add     Add ${IP} ${HOSTNAME} to ${HOSTS_FILE} (requires sudo)
  remove  Remove ${HOSTNAME} entry from ${HOSTS_FILE} (requires sudo)
  status  Show whether the dev hostname is configured

Override: DEV_HOSTNAME=my.test.local DEV_HOST_IP=127.0.0.1 $0 add
EOF
}

has_entry() {
  grep -qE "^[[:space:]]*${IP}[[:space:]]+${HOSTNAME}([[:space:]]|$)" "$HOSTS_FILE" 2>/dev/null
}

cmd="${1:-}"
case "$cmd" in
  add)
    if has_entry; then
      printf "already configured: %s -> %s\n" "$HOSTNAME" "$IP"
      exit 0
    fi
    printf "adding %s %s %s\n" "$IP" "$HOSTNAME" "$MARKER"
    echo "$IP $HOSTNAME $MARKER" | sudo tee -a "$HOSTS_FILE" >/dev/null
    printf "ok — open http://%s:8091 (set APP_PUBLIC_URL in infra/.env.dev)\n" "$HOSTNAME"
    ;;
  remove)
    if ! has_entry; then
      printf "not configured: %s\n" "$HOSTNAME"
      exit 0
    fi
    if sed --version >/dev/null 2>&1; then
      sudo sed -i "/[[:space:]]${HOSTNAME}/d" "$HOSTS_FILE"
    else
      sudo sed -i '' "/[[:space:]]${HOSTNAME}/d" "$HOSTS_FILE"
    fi
    printf "removed %s from %s\n" "$HOSTNAME" "$HOSTS_FILE"
    ;;
  status)
    if has_entry; then
      printf "configured: %s -> %s\n" "$HOSTNAME" "$IP"
      grep "$HOSTNAME" "$HOSTS_FILE" || true
    else
      printf "not configured: %s\n" "$HOSTNAME"
      exit 1
    fi
    ;;
  *)
    usage
    exit 1
    ;;
esac