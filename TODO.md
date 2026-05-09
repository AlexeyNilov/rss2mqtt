# TODO

Step-by-step implementation plan for the MVP.

## Assumptions

* The MVP is one Go command-line binary named `rss2mqtt`.
* The Go module path is `github.com/AlexeyNilov/rss2mqtt`.
* The app reads `rss.yaml` from the current working directory.
* The app runs once, processes all configured feeds, writes approved items to stdout, saves duplicate state, and exits.
* RSS item filtering searches the item title and RSS `description`.
* Filters are case-insensitive substrings; matching any configured substring approves the item.
* Stdout output is human-readable and intended for manual inspection or systemd logs, not as a stable machine API.
* Duplicate suppression state is stored in `.rss2mqtt-state.json` as bounded per-feed item identity hashes.
* `sample/feed.xml` is the canonical local RSS fixture for early parser and filtering tests.

## Sample Feed Reference

Use `sample/feed.xml` as the first parser fixture. It is a WordPress-style RSS 2.0 feed with 15 items and includes these useful item fields:

* `title`
* `link`
* `pubDate`
* `dc:creator`
* `category`
* `guid`
* `description`
* `content:encoded`
* `media:content`
* `media:thumbnail`

Use `description` as the MVP content field. Do not filter against `content:encoded` or media fields unless a later requirement explicitly asks for it.

## Proposed Repository Layout

```text
cmd/rss2mqtt/main.go
internal/config/
internal/feed/
internal/filter/
internal/state/
internal/output/
doc/
rss.yaml.example
go.mod
README.md
TODO.md
```

Keep `main.go` thin. Business logic belongs in `internal/*` packages so it can be tested without running the binary.

## Phase 1: Project Skeleton

* [x] Initialize the Go module with `github.com/AlexeyNilov/rss2mqtt`.
* [x] Create the minimal directory layout under `cmd/rss2mqtt` and `internal`.
* [x] Add `.gitignore` entries for local config, local state, binaries, and test artifacts.
* [x] Add `rss.yaml.example` showing the initial config shape.
* [x] Update `README.md` with the MVP purpose, run-once behavior, config file name, and local development commands.
* [x] Run `go test ./...` to verify the empty skeleton is valid.

## Phase 2: Configuration Package

Target package: `internal/config`.

* [x] Write failing tests for loading a valid `rss.yaml` with multiple feeds.
* [x] Write failing tests for rejecting missing feed name, missing URL, empty filters, and duplicate feed names.
* [x] Write failing tests for clear errors when the file is missing or YAML is invalid.
* [x] Implement config structs and YAML loading.
* [x] Implement validation with small focused functions.
* [x] Keep the config API independent from the current working directory; pass the file path into the loader.
* [x] Run `go test ./internal/config`.

Initial config shape:

```yaml
- name: example-feed
  url: https://example.com/rss.xml
  filters:
    - important substring
    - another match
```

## Phase 3: Feed Item Model and RSS Fetching

Target package: `internal/feed`.

* [x] Choose the RSS parsing approach before coding:
  * Prefer a small, maintained Go RSS/Atom parser if it avoids fragile XML handling.
  * If adding a dependency, record the decision in `doc/decisions.md`.
* [x] Define a normalized `Item` type with at least feed name, title, content, link, published time if available, and a stable raw identity candidate if available.
* [x] Write failing tests using `sample/feed.xml`, not live network calls.
* [x] Write failing tests for extracting `description` as the item content.
* [x] Write failing tests for parse failures returning useful errors.
* [x] Implement parsing from an `io.Reader` first.
* [x] Implement HTTP fetching separately using an injectable `http.Client`.
* [x] Set conservative HTTP timeouts.
* [x] Run `go test ./internal/feed`.

## Phase 4: Filtering Package

Target package: `internal/filter`.

* [x] Write failing tests showing that matching any configured substring approves an item.
* [x] Write failing tests showing matching is case-insensitive.
* [x] Write failing tests showing title and description are both searched.
* [x] Write failing tests showing non-matching items are rejected.
* [x] Implement the matcher as pure logic with no I/O.
* [x] Run `go test ./internal/filter`.

## Phase 5: Duplicate Suppression State

Target package: `internal/state`.

* [x] Decide the local state design before implementation:
  * state file path
  * state file format
  * item identity algorithm
  * number of remembered items per feed or retention window
  * behavior when state file is missing or corrupt
* [x] Record the accepted state design in `doc/decisions.md`.
* [x] Write failing tests for allowing a new item.
* [x] Write failing tests for suppressing an item seen in an earlier run.
* [x] Write failing tests for keeping state separated by feed name.
* [x] Write failing tests for loading missing state as empty state.
* [x] Write failing tests for saving state atomically enough for a small local file.
* [x] Implement the state store with a small public API. Defer an explicit Go interface until the orchestration package has a consumer that needs it.
* [x] Run `go test ./internal/state`.

Accepted state design:

```text
.rss2mqtt-state.json
feed name -> bounded list of item identity hashes
hash input -> normalized item identity from the feed package
retention -> latest 256 item hashes per feed
missing state -> empty state
corrupt state -> error
```

## Phase 6: Human-Readable Output

Target package: `internal/output`.

* [x] Write failing tests for formatting one approved item.
* [x] Write failing tests for including enough context to debug matches: feed name, title, link if available, and content excerpt if useful.
* [x] Keep the formatter independent from stdout; accept an `io.Writer`.
* [x] Implement human-readable output.
* [x] Run `go test ./internal/output`.

## Phase 7: Application Orchestration

Target location: either `cmd/rss2mqtt/main.go` only, or add `internal/app` if orchestration becomes too large.

* [x] Write tests for orchestration if the logic cannot stay as thin wiring in `main.go`.
* [x] Load `rss.yaml`.
* [x] Load duplicate state.
* [x] For each configured feed, fetch and parse items.
* [x] For each item, filter first, then check duplicate state.
* [x] Print approved, non-duplicate items.
* [x] Mark processed approved items in state.
* [x] Save state at the end.
* [x] Continue processing other feeds when one feed fetch or parse fails.
* [x] Return a non-zero exit code only for startup/config failures or other fatal application errors.
* [x] Run `go test ./...`.

Important behavior to settle:

* [x] Decide whether filtered-out items should be recorded in duplicate state. Decision: no, because changing filters later should allow older matching items to appear.
* [x] Decide whether duplicate state is updated before or after successful stdout write. Decision: after successful write.

## Phase 8: CLI and Runtime Polish

* [ ] Keep the initial CLI minimal: no flags unless a real need appears.
* [ ] Use logging for errors and operational messages; use stdout only for approved item output.
* [ ] Send errors and diagnostics to stderr.
* [ ] Confirm the app exits after one full pass.

## Phase 9: Verification

* [ ] Run `go test ./...`.
* [ ] Run the app against a local fixture or controlled test feed.
* [ ] Verify invalid config exits clearly.
* [ ] Verify one broken feed does not stop other feeds.
* [ ] Verify duplicate suppression across two invocations.
* [ ] Verify human-readable stdout contains approved items only.
* [ ] Verify diagnostics do not pollute stdout.
* [ ] Build the binary for the target platform.

## Deferred Work

* [ ] MQTT output package and configuration.
* [ ] systemd service and timer files.
* [ ] More advanced duplicate retention policy if the simple local state grows too large.
