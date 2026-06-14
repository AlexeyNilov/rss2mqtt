#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

POSITIONAL_ARGS=()
for arg in "$@"; do
  case "$arg" in
    --print-only|--help|-h)
      ;;
    *)
      POSITIONAL_ARGS+=("$arg")
      ;;
  esac
done

PI_USER_VALUE="${PI_USER:-pi}"
if [[ "${#POSITIONAL_ARGS[@]}" -ge 2 ]]; then
  PI_USER_VALUE="${POSITIONAL_ARGS[1]}"
fi

REMOTE_APP_DIR_VALUE="${REMOTE_APP_DIR:-/home/${PI_USER_VALUE}/githubtrending2mqtt}"
if [[ "${#POSITIONAL_ARGS[@]}" -ge 3 ]]; then
  REMOTE_APP_DIR_VALUE="${POSITIONAL_ARGS[2]}"
fi

export SERVICE_NAME="${SERVICE_NAME:-githubtrending2mqtt}"
export SERVICE_DESCRIPTION="${SERVICE_DESCRIPTION:-githubtrending2mqtt GitHub Trending check}"
export TIMER_DESCRIPTION="${TIMER_DESCRIPTION:-Run githubtrending2mqtt GitHub Trending check}"
export TIMER_ON_CALENDAR="${TIMER_ON_CALENDAR:-Sun *-*-* 16:00:00}"
export REMOTE_APP_DIR="${REMOTE_APP_DIR:-$REMOTE_APP_DIR_VALUE}"
export REMOTE_BIN_PATH="${REMOTE_BIN_PATH:-$REMOTE_APP_DIR_VALUE/githubtrending2mqtt}"
export REMOTE_CONFIG_PATH="${REMOTE_CONFIG_PATH:-$REMOTE_APP_DIR_VALUE/github-trending.yaml}"

exec "$SCRIPT_DIR/setup_service_pi.sh" "$@"
