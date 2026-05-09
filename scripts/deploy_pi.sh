#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
LOCAL_FILE="${LOCAL_FILE:-$ROOT_DIR/bin/rss2mqtt-linux-arm64}"
PI_HOST="${1:-${PI_HOST:-zero-control.local}}"
PI_USER="${2:-${PI_USER:-pi}}"
PI_PATH="${3:-${PI_PATH:-/home/pi/rss2mqtt/rss2mqtt}}"
SERVICE_NAME="${4:-${SERVICE_NAME:-rss2mqtt.service}}"

if [[ -z "$PI_HOST" ]]; then
  printf 'Usage: %s <pi-host> [pi-user] [remote-path] [service-name]\n' "$(basename "$0")" >&2
  printf 'Example: %s raspberrypi.local pi /home/pi/rss2mqtt/rss2mqtt rss2mqtt.service\n' "$(basename "$0")" >&2
  exit 1
fi

if [[ ! -f "$LOCAL_FILE" ]]; then
  printf 'Build artifact not found: %s\n' "$LOCAL_FILE" >&2
  printf 'Run scripts/build_arm.sh first or set LOCAL_FILE.\n' >&2
  exit 1
fi

REMOTE_TARGET="${PI_USER}@${PI_HOST}"
REMOTE_PATH_ARG="$(printf '%q' "$PI_PATH")"
SERVICE_NAME_ARG="$(printf '%q' "$SERVICE_NAME")"

start_remote_service() {
  ssh "$REMOTE_TARGET" "sudo systemctl start -- $SERVICE_NAME_ARG"
}

ssh "$REMOTE_TARGET" "sudo systemctl stop -- $SERVICE_NAME_ARG"
trap start_remote_service EXIT

scp "$LOCAL_FILE" "${REMOTE_TARGET}:${PI_PATH}"
ssh "$REMOTE_TARGET" "chmod 755 -- $REMOTE_PATH_ARG"

trap - EXIT
start_remote_service

printf 'Restarted %s after copying %s to %s:%s and setting mode 755\n' "$SERVICE_NAME" "$LOCAL_FILE" "$REMOTE_TARGET" "$PI_PATH"
