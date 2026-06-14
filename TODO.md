# githubtrending2mqtt Implementation Plan

## Phase 1: Preserve existing behavior before refactoring

- [x] Add or tighten tests around the current `rss2mqtt` run pipeline:
  config load, source fetch, filtering, duplicate suppression, publish, state save.
- [x] Add payload-format tests that make accidental MQTT/stdout output changes visible.
- [x] Confirm `go test ./...` is green before moving shared types.

## Phase 2: Extract the shared discovery pipeline

- [x] Introduce a neutral discovered item type outside the RSS package.
- [x] Move filtering, output formatting, MQTT publishing, and app orchestration to depend on the neutral item type.
- [x] Keep RSS parsing source-specific: RSS parser maps feed entries into neutral discovered items.
- [x] Rename only where it improves clarity; avoid a broad cosmetic rename while behavior is moving.
- [x] Keep `rss2mqtt` defaults compatible: `rss.yaml`, `.rss2mqtt-state.json`, and existing CLI behavior.

## Phase 3: Add GitHub Trending discovery

- [x] Create `internal/githubtrending` with a parser that accepts HTML and returns normalized discovered repository items.
- [x] Use saved fixture HTML in tests; do not test against live GitHub in unit tests.
- [x] Capture stable identities from repository owner/name, not from display text alone.
- [x] Include useful payload fields in title/description/link without making MQTT output GitHub-specific.
- [x] Add a fetcher that can request pages such as `https://github.com/trending/python?since=weekly`.

## Phase 4: Add githubtrending2mqtt command and config

- [x] Add `cmd/githubtrending2mqtt`.
- [x] Add a separate config file, for example `github-trending.yaml`.
- [x] Support multiple trending pages, for example language plus `since` interval.
- [x] Use a separate state file, for example `.githubtrending2mqtt-state.json`.
- [x] Support the same output choices as `rss2mqtt`: MQTT by default and stdout for inspection.

## Phase 5: Raspberry Pi deployment

- [x] Build `githubtrending2mqtt` for Linux ARM64.
- [ ] Deploy it with `APP_NAME=githubtrending2mqtt scripts/deploy_pi.sh`.
- [ ] Use `scripts/setup_githubtrending_service_pi.sh` to install a Sunday 16:00 timer.
- [ ] Verify generated units with `scripts/setup_githubtrending_service_pi.sh --print-only`.
- [ ] Confirm the Pi has the binary, config file, MQTT `.env`, service, and timer installed.

## Phase 6: Documentation and release check

- [x] Update `README.md` with the final GitHub Trending config schema once implemented.
- [x] Add sample config for GitHub Trending.
- [x] Document the separate state file and systemd timer.
- [x] Run `go test ./...`.
- [ ] Run shell syntax checks for deployment scripts.
