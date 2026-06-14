# githubtrending2mqtt Implementation Plan

## Phase 1: Preserve existing behavior before refactoring

- [ ] Add or tighten tests around the current `rss2mqtt` run pipeline:
  config load, source fetch, filtering, duplicate suppression, publish, state save.
- [ ] Add payload-format tests that make accidental MQTT/stdout output changes visible.
- [ ] Confirm `go test ./...` is green before moving shared types.

## Phase 2: Extract the shared discovery pipeline

- [ ] Introduce a neutral discovered item type outside the RSS package.
- [ ] Move filtering, output formatting, MQTT publishing, and app orchestration to depend on the neutral item type.
- [ ] Keep RSS parsing source-specific: RSS parser maps feed entries into neutral discovered items.
- [ ] Rename only where it improves clarity; avoid a broad cosmetic rename while behavior is moving.
- [ ] Keep `rss2mqtt` defaults compatible: `rss.yaml`, `.rss2mqtt-state.json`, and existing CLI behavior.

## Phase 3: Add GitHub Trending discovery

- [ ] Create `internal/githubtrending` with a parser that accepts HTML and returns normalized discovered repository items.
- [ ] Use saved fixture HTML in tests; do not test against live GitHub in unit tests.
- [ ] Capture stable identities from repository owner/name, not from display text alone.
- [ ] Include useful payload fields in title/description/link without making MQTT output GitHub-specific.
- [ ] Add a fetcher that can request pages such as `https://github.com/trending/python?since=weekly`.

## Phase 4: Add githubtrending2mqtt command and config

- [ ] Add `cmd/githubtrending2mqtt`.
- [ ] Add a separate config file, for example `github-trending.yaml`.
- [ ] Support multiple trending pages, for example language plus `since` interval.
- [ ] Use a separate state file, for example `.githubtrending2mqtt-state.json`.
- [ ] Support the same output choices as `rss2mqtt`: MQTT by default and stdout for inspection.

## Phase 5: Raspberry Pi deployment

- [ ] Build `githubtrending2mqtt` with `scripts/build_arm.sh githubtrending2mqtt`.
- [ ] Deploy it with `APP_NAME=githubtrending2mqtt scripts/deploy_pi.sh`.
- [ ] Use `scripts/setup_githubtrending_service_pi.sh` to install a Sunday 16:00 timer.
- [ ] Verify generated units with `scripts/setup_githubtrending_service_pi.sh --print-only`.
- [ ] Confirm the Pi has the binary, config file, MQTT `.env`, service, and timer installed.

## Phase 6: Documentation and release check

- [ ] Update `README.md` with the final GitHub Trending config schema once implemented.
- [ ] Add sample config for GitHub Trending.
- [ ] Document the separate state file and systemd timer.
- [ ] Run `go test ./...`.
- [ ] Run shell syntax checks for deployment scripts.
