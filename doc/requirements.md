# Requirements

## EARS (Easy Approach to Requirements Syntax)

Use the EARS structure for precise requirements:

> **While** `<optional precondition>`, **when** `<optional trigger>`, **the system shall** `<system response>`.

This helps ensure requirements are:

* Context-aware
* Trigger-based
* Action-specific

## Actual requirements

### Runtime and deployment

* The system shall be implemented in Go.
* The Go module path shall be `github.com/AlexeyNilov/rss2mqtt`.
* While running on the target device, the system shall be suitable for Raspberry Pi Zero 2 class hardware.
* The system shall avoid unnecessary background services and runtime dependencies.

### Configuration

* When the application starts, the system shall read its configuration from `rss.yaml` in the local working directory.
* When the YAML file contains multiple RSS feed entries, the system shall load all configured feed entries.
* The YAML configuration shall be a list of RSS feed entries.
* Each configured RSS feed entry shall include a unique RSS name.
* Each configured RSS feed entry shall include the feed URL.
* Each configured RSS feed entry shall include a list of substrings used for filtering.
* If the configuration file cannot be read, the system shall report the error and stop.
* If the configuration file is invalid, the system shall report the validation error and stop.

Initial YAML shape:

```yaml
- name: example-feed
  url: https://example.com/rss.xml
  filters:
    - important substring
    - another match
```

### RSS input

* When a configured RSS feed is processed, the system shall fetch items from that feed URL.
* When all configured RSS feeds have been processed, the system shall exit.
* If a configured RSS feed cannot be fetched, the system shall report the error without preventing other configured feeds from being processed.
* If a configured RSS feed response cannot be parsed as RSS, the system shall report the error without preventing other configured feeds from being processed.

### Filtering

* When an RSS item is received from a configured feed, the system shall evaluate the item against that feed's configured filter substrings.
* When an RSS item's title or RSS `description` contains at least one configured filter substring, the system shall approve the item for relay.
* When an RSS item does not match the configured filter rules, the system shall not relay the item.
* When evaluating filter substrings, the system shall match substrings case-insensitively.

### Duplicate suppression

* When an RSS item has already been processed in an earlier scheduled run, the system shall not relay that item again.
* The system shall persist enough local state to suppress duplicates across separate application invocations.
* The system shall associate duplicate suppression state with the configured RSS feed name.

### MVP output

* While MQTT support is not yet implemented, when an RSS item is approved for relay, the system shall print the approved item to stdout as human-readable text.
* When an RSS item is not approved for relay, the system shall not print that item to stdout.

### Future MQTT output

* When MQTT support is implemented, the system shall publish approved RSS items to a configured MQTT topic.
* When MQTT support is implemented, MQTT connection settings shall be configurable.

## Open requirements

None at this time.
