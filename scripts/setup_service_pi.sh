#!/usr/bin/env bash

set -euo pipefail

SERVICE_NAME="${SERVICE_NAME:-rss2mqtt}"
SERVICE_DESCRIPTION="${SERVICE_DESCRIPTION:-rss2mqtt bot}"
PRINT_ONLY=0

usage() {
  printf 'Usage: %s [--print-only] <pi-host> [pi-user] [remote-app-dir] [service-name]\n' "$(basename "$0")" >&2
  printf 'Example: %s raspberrypi.local pi /home/pi/rss2mqtt rss2mqtt\n' "$(basename "$0")" >&2
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --print-only)
      PRINT_ONLY=1
      shift
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    *)
      break
      ;;
  esac
done

PI_HOST="${1:-${PI_HOST:-zero-control.local}}"
PI_USER="${2:-${PI_USER:-pi}}"
REMOTE_APP_DIR="${3:-${REMOTE_APP_DIR:-/home/${PI_USER}/rss2mqtt}}"
SERVICE_NAME="${4:-${SERVICE_NAME}}"
REMOTE_BIN_PATH="${REMOTE_BIN_PATH:-$REMOTE_APP_DIR/rss2mqtt}"
REMOTE_ENV_PATH="${REMOTE_ENV_PATH:-$REMOTE_APP_DIR/.env}"
SERVICE_PATH="/etc/systemd/system/${SERVICE_NAME}.service"

render_service_unit() {
  cat <<EOF
[Unit]
Description=${SERVICE_DESCRIPTION}
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=${PI_USER}
Group=${PI_USER}
WorkingDirectory=${REMOTE_APP_DIR}
EnvironmentFile=${REMOTE_ENV_PATH}
ExecStart=${REMOTE_BIN_PATH}
Restart=on-failure
RestartSec=5
NoNewPrivileges=true
PrivateTmp=true

[Install]
WantedBy=multi-user.target
EOF
}

if [[ "$PRINT_ONLY" -eq 1 ]]; then
  render_service_unit
  exit 0
fi

if [[ -z "$PI_HOST" ]]; then
  usage
  exit 1
fi

REMOTE_TARGET="${PI_USER}@${PI_HOST}"
TMP_SERVICE_PATH="/tmp/${SERVICE_NAME}.service"

render_service_unit | ssh "$REMOTE_TARGET" "cat > '$TMP_SERVICE_PATH'"

ssh "$REMOTE_TARGET" "
  set -euo pipefail
  if [[ ! -f '$REMOTE_BIN_PATH' ]]; then
    printf 'Remote binary not found: %s\n' '$REMOTE_BIN_PATH' >&2
    exit 1
  fi
  if [[ ! -f '$REMOTE_ENV_PATH' ]]; then
    printf 'Remote env file not found: %s\n' '$REMOTE_ENV_PATH' >&2
    exit 1
  fi
  sudo mv '$TMP_SERVICE_PATH' '$SERVICE_PATH'
  sudo chown root:root '$SERVICE_PATH'
  sudo chmod 644 '$SERVICE_PATH'
  sudo systemctl daemon-reload
  sudo systemctl enable '$SERVICE_NAME'
  sudo systemctl restart '$SERVICE_NAME'
  sudo systemctl status '$SERVICE_NAME' --no-pager
"

printf 'Configured %s on %s\n' "$SERVICE_PATH" "$REMOTE_TARGET"
