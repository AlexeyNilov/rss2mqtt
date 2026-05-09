#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUTPUT_DIR="$ROOT_DIR/bin"
OUTPUT_FILE="$OUTPUT_DIR/rss2mqtt.exe"

mkdir -p "$OUTPUT_DIR"

go build -o "$OUTPUT_FILE" ./cmd/rss2mqtt

printf 'Built %s\n' "$OUTPUT_FILE"
