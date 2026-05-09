#!/usr/bin/env bash

set -euo pipefail

SERVICE_NAME="${SERVICE_NAME:-rss2mqtt}"
SERVICE_DESCRIPTION="${SERVICE_DESCRIPTION:-rss2mqtt RSS feed check}"
TIMER_DESCRIPTION="${TIMER_DESCRIPTION:-Run rss2mqtt RSS feed check}"
TIMER_ON_CALENDAR="${TIMER_ON_CALENDAR:-*-*-* 08..20:00:00}"
PRINT_ONLY=0

usage() {
  printf 'Usage: %s [--print-only] <pi-host> [pi-user] [remote-app-dir] [service-name]\n' "$(basename "$0")" >&2
  printf 'Example: %s raspberrypi.local pi /home/pi/rss2mqtt rss2mqtt\n' "$(basename "$0")" >&2
  printf '\nEnvironment overrides:\n' >&2
  printf '  TIMER_ON_CALENDAR  systemd OnCalendar value, default: %s\n' "$TIMER_ON_CALENDAR" >&2
  printf '  REMOTE_BIN_PATH     remote binary path, default: <remote-app-dir>/rss2mqtt\n' >&2
  printf '  REMOTE_CONFIG_PATH  remote config path, default: <remote-app-dir>/rss.yaml\n' >&2
  printf '  REMOTE_ENV_PATH     remote MQTT env file path, default: <remote-app-dir>/.env\n' >&2
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
SERVICE_NAME="${SERVICE_NAME%.service}"
SERVICE_NAME="${SERVICE_NAME%.timer}"
REMOTE_BIN_PATH="${REMOTE_BIN_PATH:-$REMOTE_APP_DIR/rss2mqtt}"
REMOTE_CONFIG_PATH="${REMOTE_CONFIG_PATH:-$REMOTE_APP_DIR/rss.yaml}"
REMOTE_ENV_PATH="${REMOTE_ENV_PATH:-$REMOTE_APP_DIR/.env}"
SERVICE_PATH="/etc/systemd/system/${SERVICE_NAME}.service"
TIMER_PATH="/etc/systemd/system/${SERVICE_NAME}.timer"

render_service_unit() {
  cat <<EOF
[Unit]
Description=${SERVICE_DESCRIPTION}
After=network-online.target
Wants=network-online.target

[Service]
Type=oneshot
User=${PI_USER}
Group=${PI_USER}
WorkingDirectory=${REMOTE_APP_DIR}
EnvironmentFile=${REMOTE_ENV_PATH}
ExecStart=${REMOTE_BIN_PATH}
NoNewPrivileges=true
PrivateTmp=true
EOF
}

render_timer_unit() {
  cat <<EOF
[Unit]
Description=${TIMER_DESCRIPTION}

[Timer]
OnCalendar=${TIMER_ON_CALENDAR}
Persistent=true
Unit=${SERVICE_NAME}.service

[Install]
WantedBy=timers.target
EOF
}

if [[ "$PRINT_ONLY" -eq 1 ]]; then
  printf '# %s.service\n' "$SERVICE_NAME"
  render_service_unit
  printf '\n# %s.timer\n' "$SERVICE_NAME"
  render_timer_unit
  exit 0
fi

if [[ -z "$PI_HOST" ]]; then
  usage
  exit 1
fi

REMOTE_TARGET="${PI_USER}@${PI_HOST}"
TMP_SERVICE_PATH="/tmp/${SERVICE_NAME}.service"
TMP_TIMER_PATH="/tmp/${SERVICE_NAME}.timer"

render_service_unit | ssh "$REMOTE_TARGET" "cat > '$TMP_SERVICE_PATH'"
render_timer_unit | ssh "$REMOTE_TARGET" "cat > '$TMP_TIMER_PATH'"

ssh "$REMOTE_TARGET" "
  set -euo pipefail
  if [[ ! -f '$REMOTE_BIN_PATH' ]]; then
    printf 'Remote binary not found: %s\n' '$REMOTE_BIN_PATH' >&2
    exit 1
  fi
  if [[ ! -f '$REMOTE_CONFIG_PATH' ]]; then
    printf 'Remote config not found: %s\n' '$REMOTE_CONFIG_PATH' >&2
    exit 1
  fi
  if [[ ! -f '$REMOTE_ENV_PATH' ]]; then
    printf 'Remote env file not found: %s\n' '$REMOTE_ENV_PATH' >&2
    exit 1
  fi
  sudo mv '$TMP_SERVICE_PATH' '$SERVICE_PATH'
  sudo mv '$TMP_TIMER_PATH' '$TIMER_PATH'
  sudo chown root:root '$SERVICE_PATH'
  sudo chown root:root '$TIMER_PATH'
  sudo chmod 644 '$SERVICE_PATH'
  sudo chmod 644 '$TIMER_PATH'
  sudo systemctl daemon-reload
  sudo systemctl disable --now '$SERVICE_NAME.service' 2>/dev/null || true
  sudo systemctl enable --now '$SERVICE_NAME.timer'
  sudo systemctl list-timers '$SERVICE_NAME.timer' --no-pager
"

printf 'Configured %s and %s on %s\n' "$SERVICE_PATH" "$TIMER_PATH" "$REMOTE_TARGET"
