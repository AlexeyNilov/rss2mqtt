# Decisions

## Why record decisions

Write down key development decisions while the context is fresh. A short note today can save hours later by explaining what was chosen, what was rejected, and why the trade-off made sense at the time.

## Guidance

Use a lightweight Architecture Decision Record (ADR) style:

* Record decisions that affect architecture, data flow, public APIs, dependencies, deployment, security, or long-term maintenance.
* Write the decision when it is made, not after the context has faded.
* Prefer short entries that explain the context, decision, alternatives, and consequences.
* Include enough reasoning for a future maintainer to understand the trade-off.
* Do not document every small implementation detail; focus on choices that would be costly or confusing to rediscover.
* Update or supersede earlier decisions instead of silently rewriting history.

## Entry template

```markdown
### YYYY-MM-DD: Decision title

**Status:** Proposed | Accepted | Superseded

**Context:** What problem, constraint, or trade-off led to this decision?

**Decision:** What was chosen?

**Alternatives considered:** What other options were rejected, and why?

**Consequences:** What becomes easier, harder, riskier, or more constrained?
```

## Actual decisions

### 2026-05-09: Use MQTT output configured by local .env

**Status:** Accepted

**Context:** The production relay target is MQTT. Connection settings are local deployment details and should not live in `rss.yaml`, which describes feed behavior rather than output credentials or broker endpoints.

**Decision:** Read MQTT settings from `.env` in the working directory. Require `MQTT_BROKER_URL` and `MQTT_TOPIC`, and allow optional `MQTT_CLIENT_ID`.

**Alternatives considered:** Storing MQTT settings in `rss.yaml` would keep all configuration in one file, but it mixes feed rules with transport settings. Environment variables alone would work under systemd, but a local `.env` file is easier to manage consistently with the existing deployment scripts.

**Consequences:** Deployments must provide both `rss.yaml` and `.env`. The existing `.env` git ignore rule remains important because broker credentials may be added later.

### 2026-05-09: Use Eclipse Paho for MQTT publishing

**Status:** Accepted

**Context:** MQTT output requires a maintained client that handles broker connections, publish acknowledgements, and network transport details. The app should not implement MQTT protocol behavior directly.

**Decision:** Use `github.com/eclipse/paho.mqtt.golang` behind `internal/mqttout`.

**Alternatives considered:** Implementing MQTT directly is not reasonable for this project. Shelling out to an MQTT CLI would add operational dependencies on the Raspberry Pi. Keeping output as stdout would not satisfy the relay requirement.

**Consequences:** The app gains an MQTT client dependency, isolated behind a small package. Tests use fakes and do not require a real broker.

### 2026-05-09: Publish MQTT messages with QoS 1

**Status:** Accepted

**Context:** RSS items should not be lost silently once they pass filtering and duplicate suppression. MQTT QoS 0 would be lower overhead, but it provides no broker acknowledgement.

**Decision:** Publish approved RSS items with MQTT QoS 1.

**Alternatives considered:** QoS 0 is simpler and cheaper, but it is weaker than the reliability expected from a relay. QoS 2 is stronger but adds extra protocol overhead that is not justified for this lightweight Raspberry Pi deployment.

**Consequences:** Publishing waits for broker acknowledgement before duplicate state is updated. A network failure may cause the same item to be retried later, which is preferable to marking an item processed before the broker has accepted it.

### 2026-05-09: Mark duplicates only after successful relay

**Status:** Accepted

**Context:** The run loop filters RSS items, suppresses duplicates, relays approved items, and persists state. The ordering matters because incorrect state updates can either hide relevant items or cause repeated output.

**Decision:** Filter items before checking duplicate state. Do not record filtered-out items. For approved non-duplicate items, write the item to the current output first, then mark it as processed in duplicate state.

**Alternatives considered:** Recording all fetched items would prevent old items from appearing after filters change, which is surprising during tuning. Marking an item before output would avoid reprinting it if the process crashes after output starts, but it risks losing an item when output fails before the user sees it.

**Consequences:** Filter changes can surface older items that now match, which is useful during configuration changes. If output succeeds and the later state save fails, the item may appear again on the next run; that is preferable to silently dropping an item that was never relayed.

### 2026-05-09: Store duplicate suppression state in local JSON

**Status:** Accepted

**Context:** The application runs once per systemd timer invocation. It needs persisted local state to avoid printing or publishing the same RSS item during later hourly runs. The state should be small, easy to inspect, and avoid pulling in another storage dependency.

**Decision:** Store duplicate suppression state in `.rss2mqtt-state.json` in the working directory. The file maps each configured feed name to a bounded list of SHA-256 hashes of item identities. Keep up to 256 item hashes per feed. Treat a missing state file as empty state, and treat a corrupt state file as an error.

**Alternatives considered:** Storing only the latest item per feed would be smaller, but it would fail when feeds publish multiple relevant items between hourly runs or reorder items. SQLite would be robust, but it is unnecessary for a small bounded state file. Storing raw item URLs or GUIDs would be easier to inspect, but hashing avoids leaking full item identifiers into local state.

**Consequences:** Duplicate suppression is deterministic and lightweight. The trade-off is that the 256-item retention limit is a heuristic; very high-volume feeds could still reprocess older items after they fall out of the retained set.

### 2026-05-09: Use gofeed for RSS and Atom parsing

**Status:** Accepted

**Context:** The application needs to parse RSS feeds reliably, including common extensions such as WordPress `content:encoded`, Dublin Core metadata, GUIDs, links, publication dates, and Atom feeds. The sample feed already includes namespaces and extension fields that make custom XML parsing a poor default.

**Decision:** Use `github.com/mmcdole/gofeed` for RSS and Atom parsing, and normalize only the fields needed by this application.

**Alternatives considered:** Parsing RSS directly with `encoding/xml` would avoid an external dependency, but it would push feed-format edge cases into this codebase. Supporting only the exact sample feed shape would be too brittle for a tool intended to subscribe to several feeds.

**Consequences:** The feed package remains focused on normalization and error handling rather than XML details. The project accepts one parser dependency, which should be kept behind `internal/feed` so it can be replaced later if needed.

### 2026-05-09: Use yaml.v3 for configuration parsing

**Status:** Accepted

**Context:** The application reads hand-edited YAML configuration. Correct YAML parsing includes details such as indentation, lists, scalars, and useful parse errors. Implementing even a narrow YAML parser locally would add fragile parsing logic that is not part of the project's core value.

**Decision:** Use `gopkg.in/yaml.v3` for parsing `rss.yaml`.

**Alternatives considered:** A custom parser for the current small config shape would avoid one dependency, but it would be brittle and would not behave like normal YAML. Switching to JSON would avoid a YAML dependency but conflicts with the accepted configuration format.

**Consequences:** The project gains one small external dependency for configuration parsing. The parser behavior is conventional and easier to test, but dependency updates become part of normal maintenance.

### 2026-05-09: Use GitHub module path

**Status:** Accepted

**Context:** The Go project needs a stable module path before code and internal imports are created.

**Decision:** Use `github.com/AlexeyNilov/rss2mqtt` as the Go module path.

**Alternatives considered:** A local-only module name would work during early development, but it would create unnecessary churn once the repository is pushed or imported from GitHub.

**Consequences:** The module can be initialized directly with the final public import path. Package imports should use this path consistently once code is added.

### 2026-05-09: Suppress duplicate items across scheduled runs

**Status:** Accepted

**Context:** The MVP will be run periodically by a systemd timer. Without persisted state, each hourly run could print or later publish the same feed items repeatedly.

**Decision:** The application will suppress duplicate items across separate invocations by storing local state per configured feed name. The exact state format is not decided yet, but a likely approach is to store an identifier or hash for recently processed items per feed.

**Alternatives considered:** Relying only on RSS publication timestamps would be fragile because feeds may omit timestamps, edit old items, or publish items out of order. Doing no duplicate suppression would keep the MVP simpler, but it would make hourly scheduled use noisy and less useful.

**Consequences:** The application will need a small local persistence boundary even before MQTT support is added. The unresolved design questions are where the state file lives, how much history is kept, and how item identity is computed.

### 2026-05-09: Use `rss.yaml` as the local config file

**Status:** Accepted

**Context:** The MVP needs a simple, predictable configuration path for local runs and systemd timer execution.

**Decision:** The application will read RSS feed configuration from a local working-directory file named `rss.yaml`.

**Alternatives considered:** A command-line config path would be more flexible, and a system path such as `/etc/rss2mqtt/rss.yaml` would fit packaged Linux deployment better. For the first version, a local file is simpler and avoids introducing install-layout decisions too early.

**Consequences:** Local testing is straightforward. Systemd units will need to set the working directory explicitly or run the binary from the directory containing `rss.yaml`.

### 2026-05-09: Run the MVP once per invocation

**Status:** Accepted

**Context:** The first version does not need to own scheduling. The intended deployment model is to install the application behind a systemd timer that runs every hour during daytime.

**Decision:** The MVP will process the configured feeds once and then exit.

**Alternatives considered:** A long-running daemon with internal polling would centralize scheduling inside the application, but it adds lifecycle, retry, sleep, and observability concerns before the core feed filtering behavior is proven.

**Consequences:** The application stays simpler and easier to supervise with standard Linux tools. The trade-off is that cross-run concerns, especially duplicate suppression, need a deliberate design if repeated hourly runs would otherwise reprint or republish old feed items.

### 2026-05-09: Use simple any-substring, case-insensitive filtering over title and description

**Status:** Accepted

**Context:** Each RSS feed has a configured list of substrings that decide whether an item should be relayed. The matching rule needs to be simple enough for the MVP while still useful for broad topic filtering.

**Decision:** An item is approved when its title or RSS `description` contains any configured filter substring, using case-insensitive matching.

**Alternatives considered:** All-filter matching would be stricter but easier to misconfigure and could silently miss relevant items. Whole-word matching would reduce false positives, but substring matching is simpler and acceptable for the first version. Case-sensitive matching would be more exact, but it is a poor default for human-written RSS text and filter lists.

**Consequences:** The MVP favors recall over precision. Users may receive some false positives when a substring appears inside an unrelated word, but the behavior is simple to understand and test.

### 2026-05-09: Print human-readable output in the MVP

**Status:** Superseded by "2026-05-09: Use MQTT output configured by local .env"

**Context:** The stdout stage exists to verify feed loading and filtering before MQTT support is added. The user should be able to inspect results directly in a terminal or timer logs.

**Decision:** Approved RSS items will be printed as human-readable text during the MVP.

**Alternatives considered:** JSON lines would be easier for downstream automation and snapshot tests, but it is less convenient for manual inspection. Since MQTT will become the production output later, stdout is optimized for early human verification.

**Consequences:** Manual testing is straightforward. If automated consumers later depend on stdout, a structured output mode may need to be added explicitly instead of relying on the human-readable format.

### 2026-05-09: Build the service in Go

**Status:** Accepted

**Context:** The application needs to run on Raspberry Pi Zero 2 class hardware with low operational overhead. The project should be easy to deploy as a small standalone process.

**Decision:** Implement the application in Go.

**Alternatives considered:** A scripting language such as Python could make early prototyping faster, but it adds an interpreter dependency and usually a larger runtime footprint. A shell-based implementation would be smaller initially, but it would become brittle once RSS parsing, filtering, MQTT publishing, configuration validation, and duplicate handling are added.

**Consequences:** Go supports a single compiled binary and a small runtime footprint, which fits the target device. The trade-off is that early iteration may be slightly more verbose than a script, but the production deployment path is cleaner.

### 2026-05-09: Start with stdout before MQTT

**Status:** Superseded by "2026-05-09: Use MQTT output configured by local .env"

**Context:** The target behavior is to relay approved RSS items to MQTT, but the filtering and feed-processing behavior can be validated independently from MQTT connectivity.

**Decision:** The first version will print approved RSS items to stdout instead of publishing them to MQTT.

**Alternatives considered:** Implementing MQTT immediately would test the final transport earlier, but it would mix feed parsing, filtering, configuration, network output, and broker configuration before the core behavior is proven.

**Consequences:** The MVP can validate feed loading and filtering with fewer moving parts. MQTT integration remains a future step and should be added behind a clear output boundary so stdout can continue to be useful for testing.

### 2026-05-09: Configure feeds and filters with YAML

**Status:** Accepted

**Context:** The user needs to subscribe to several RSS feeds and assign substring filters to each feed. The configuration should be editable on a small Linux device without recompiling the application.

**Decision:** Store the list of RSS feeds and their corresponding filter substrings in a YAML configuration file.

**Alternatives considered:** Command-line flags would be awkward for multiple feeds and filters. JSON would work, but it is less convenient for hand-edited configuration. Environment variables are suitable for secrets and deployment-specific values, but not for structured feed lists.

**Consequences:** YAML is practical for hand-written configuration and supports structured feed definitions. The project should keep the schema simple and validate it at startup so configuration mistakes fail clearly.

### 2026-05-03: Target Raspberry Pi Zero 2 WH with Raspberry Pi OS Lite 64-bit

**Status:** Accepted

**Context:** The project needs a minimal, reliable operating system target for deploying the Go bot on a Raspberry Pi Zero 2 WH. The main requirement is to run a compiled Go application with low operational overhead. The main options considered were Raspberry Pi OS Lite, DietPi, Alpine Linux, and Arch Linux ARM. The original Raspberry Pi Zero constraints around ARMv6 do not apply to the Zero 2 WH, which uses an ARMv8 CPU.

**Decision:** Standardize on Raspberry Pi OS Lite 64-bit as the deployment target for the Raspberry Pi Zero 2 WH, and build the bot as a `linux/arm64` binary.

**Alternatives considered:** Raspberry Pi OS Lite 32-bit would be a more conservative compatibility choice, but it adds no clear benefit for a pure Go bot on ARMv8 hardware. DietPi is attractive for aggressive minimalism, but it adds project-specific tooling on top of Debian without solving a real problem for this deployment. Alpine Linux is smaller, but its `musl`-based environment introduces avoidable compatibility risk if the project later adds native dependencies. Arch Linux ARM is flexible, but it is a weaker default for a small deployment-focused project that benefits more from predictability than from distribution minimalism for its own sake.

**Consequences:** This keeps the runtime small while preserving the official Raspberry Pi kernel, package ecosystem, and hardware support path. It also simplifies deployment by aligning the project with a standard `arm64` Linux target. The trade-off is that the base image is not as stripped down as Alpine or DietPi, but the operational risk is lower and the environment is more conventional.
