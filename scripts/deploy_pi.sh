#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
APP_NAME="${APP_NAME:-rss2mqtt}"
LOCAL_FILE="${LOCAL_FILE:-$ROOT_DIR/bin/${APP_NAME}-linux-arm64}"
PI_HOST="${1:-${PI_HOST:-zero-control.local}}"
PI_USER="${2:-${PI_USER:-pi}}"
PI_PATH="${3:-${PI_PATH:-/home/pi/${APP_NAME}/${APP_NAME}}}"
SERVICE_NAME="${4:-${SERVICE_NAME:-$APP_NAME}}"
SERVICE_NAME="${SERVICE_NAME%.service}"
SERVICE_NAME="${SERVICE_NAME%.timer}"
TIMER_NAME="${SERVICE_NAME}.timer"

if [[ -z "$PI_HOST" ]]; then
  printf 'Usage: %s <pi-host> [pi-user] [remote-path] [service-name]\n' "$(basename "$0")" >&2
  printf 'Example: %s raspberrypi.local pi /home/pi/rss2mqtt/rss2mqtt rss2mqtt\n' "$(basename "$0")" >&2
  printf 'Example: APP_NAME=githubtrending2mqtt %s raspberrypi.local pi /home/pi/githubtrending2mqtt/githubtrending2mqtt githubtrending2mqtt\n' "$(basename "$0")" >&2
  exit 1
fi

if [[ ! -f "$LOCAL_FILE" ]]; then
  printf 'Build artifact not found: %s\n' "$LOCAL_FILE" >&2
  printf 'Run scripts/build_arm.sh first or set LOCAL_FILE.\n' >&2
  exit 1
fi

REMOTE_TARGET="${PI_USER}@${PI_HOST}"
REMOTE_PATH_ARG="$(printf '%q' "$PI_PATH")"
TIMER_NAME_ARG="$(printf '%q' "$TIMER_NAME")"

scp "$LOCAL_FILE" "${REMOTE_TARGET}:${PI_PATH}"
ssh "$REMOTE_TARGET" "
  set -euo pipefail
  chmod 755 -- $REMOTE_PATH_ARG
  if systemctl cat -- $TIMER_NAME_ARG >/dev/null 2>&1; then
    sudo systemctl daemon-reload
    sudo systemctl restart -- $TIMER_NAME_ARG
    sudo systemctl list-timers --no-pager $TIMER_NAME_ARG
  else
    printf 'Timer unit not installed yet: %s\n' '$TIMER_NAME' >&2
    printf 'Run scripts/setup_service_pi.sh to install the systemd timer.\n' >&2
  fi
"

printf 'Copied %s to %s:%s and set mode 755\n' "$LOCAL_FILE" "$REMOTE_TARGET" "$PI_PATH"
