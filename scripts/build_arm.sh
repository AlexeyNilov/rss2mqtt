#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUTPUT_DIR="$ROOT_DIR/bin"
OUTPUT_FILE="$OUTPUT_DIR/rss2mqtt-linux-arm64"

mkdir -p "$OUTPUT_DIR"

export GOOS=linux
export GOARCH=arm64
export CGO_ENABLED=0

go build -o "$OUTPUT_FILE" ./cmd/rss2mqtt
ls -lh "$OUTPUT_FILE"

printf 'Built %s\n' "$OUTPUT_FILE"
