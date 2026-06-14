# rss2mqtt

This repository is evolving from a single RSS utility into a small family of
"discover new items, suppress duplicates, publish to MQTT" tools.

`rss2mqtt` is the first tool in that family. It reads RSS feeds, filters feed
items by configured substrings, suppresses duplicates across scheduled runs, and
relays approved items to MQTT.

The tools are designed for Raspberry Pi Zero 2 class hardware. Each command runs
once and exits, so scheduling is handled outside the process, typically by a
systemd timer.

Tools:

* `rss2mqtt`: discover matching RSS feed items.
* `githubtrending2mqtt`: discover repositories from GitHub Trending pages and
  publish them through the same MQTT/state pipeline.

## Shared Design

The long-term shape of the project is source-specific discovery commands over a
shared delivery pipeline:

* Source packages parse external pages or feeds into normalized discovered
  items.
* Filtering decides which discovered items are eligible for delivery.
* Local state suppresses duplicates across scheduled runs.
* Output relayers publish approved items to MQTT or stdout.

RSS parsing should remain isolated from GitHub Trending parsing. MQTT publishing,
state storage, output formatting, deployment scripts, and the run-once scheduling
model should be shared.

## rss2mqtt Features

* Multiple RSS feeds configured in `rss.yaml`.
* Per-feed substring filters.
* Case-insensitive matching against item title and RSS `description`.
* Duplicate suppression with local `.rss2mqtt-state.json` state.
* MQTT output by default.
* Optional stdout output for local inspection.
* MQTT QoS 1 publishing.
* systemd service/timer installation script for Raspberry Pi.

## Configuration

Run the binary from a directory containing `rss.yaml`.

```yaml
- name: oreilly-radar
  url: https://www.oreilly.com/radar/feed/
  filters:
    - AI
    - agent
```

`name` must be unique per feed. Filters are matched against the RSS item title and `description`.

For MQTT output, create `.env` in the same working directory:

```env
MQTT_BROKER_URL=tcp://localhost:1883
MQTT_TOPIC=rss/approved
MQTT_CLIENT_ID=rss2mqtt
```

`MQTT_BROKER_URL` and `MQTT_TOPIC` are required. `MQTT_CLIENT_ID` is optional and defaults to `rss2mqtt`.

## Running

MQTT is the default output:

```bash
rss2mqtt
```

Use stdout output for local inspection without requiring `.env`:

```bash
rss2mqtt --output stdout
```

Supported output values are `mqtt` and `stdout`.

## githubtrending2mqtt Configuration

Run the binary from a directory containing `github-trending.yaml`.

```yaml
- name: python-weekly
  language: python
  since: weekly
  filters:
    - "*"

- name: go-monthly
  language: go
  since: monthly
  filters:
    - mqtt
    - cli
```

`name` must be unique per page. `language` is optional; omit it to read the
global Trending page. `since` is optional and defaults to `weekly`; supported
values are `daily`, `weekly`, and `monthly`.

You can also provide an explicit GitHub Trending URL:

```yaml
- name: python-weekly
  url: https://github.com/trending/python?since=weekly
  filters:
    - "*"
```

Filters are matched against the repository title and description. Use `"*"` to
approve every discovered repository from a configured page.

For MQTT output, `githubtrending2mqtt` uses the same `.env` format as
`rss2mqtt`.

## Development

Run tests:

```
go test ./...
```

Build locally:

```
go build ./cmd/rss2mqtt
go build ./cmd/githubtrending2mqtt
```

Build `rss2mqtt` for Raspberry Pi Zero 2:

```bash
scripts/build_arm.sh
```

Build `githubtrending2mqtt`:

```bash
scripts/build_arm.sh githubtrending2mqtt
```

## Raspberry Pi Deployment

Copy a new binary to the Pi:

```bash
scripts/deploy_pi.sh raspberrypi.local pi /home/pi/rss2mqtt/rss2mqtt rss2mqtt
```

Copy `githubtrending2mqtt` to the Pi:

```bash
APP_NAME=githubtrending2mqtt scripts/deploy_pi.sh raspberrypi.local pi /home/pi/githubtrending2mqtt/githubtrending2mqtt githubtrending2mqtt
```

Install the systemd service and timer:

```bash
scripts/setup_service_pi.sh raspberrypi.local pi /home/pi/rss2mqtt rss2mqtt
```

Install the GitHub Trending service and timer:

```bash
scripts/setup_githubtrending_service_pi.sh raspberrypi.local pi /home/pi/githubtrending2mqtt
```

That installer schedules `githubtrending2mqtt` every Sunday at 16:00 local device
time.

The RSS installer creates a run-once `rss2mqtt.service` and enables
`rss2mqtt.timer`. By default, the timer runs hourly from 08:00 through 20:00
local device time.

Override the timer schedule:

```bash
TIMER_ON_CALENDAR='*-*-* 08..20:00:00' scripts/setup_service_pi.sh raspberrypi.local
```

Inspect generated systemd units without installing them:

```bash
scripts/setup_service_pi.sh --print-only raspberrypi.local
```

The deploy script copies the binary, sets executable mode, and restarts the timer if it is already installed. It does not stop or start the run-once service directly.

## Local Files

These files are intentionally local and ignored by git:

* `.env`
* `rss.yaml`
* `.rss2mqtt-state.json`
* `github-trending.yaml`
* `.githubtrending2mqtt-state.json`
