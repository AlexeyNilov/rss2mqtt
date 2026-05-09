# rss2mqtt

`rss2mqtt` reads RSS feeds, filters feed items by configured substrings, suppresses duplicates across scheduled runs, and relays approved items to MQTT.

The app is designed for Raspberry Pi Zero 2 class hardware. It runs once and exits, so scheduling is handled outside the process, typically by a systemd timer.

## Features

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

## Development

Run tests:

```
go test ./...
```

Build locally:

```
go build ./cmd/rss2mqtt
```

Build for Raspberry Pi Zero 2:

```bash
scripts/build_arm.sh
```

## Raspberry Pi Deployment

Copy a new binary to the Pi:

```bash
scripts/deploy_pi.sh raspberrypi.local pi /home/pi/rss2mqtt/rss2mqtt rss2mqtt
```

Install the systemd service and timer:

```bash
scripts/setup_service_pi.sh raspberrypi.local pi /home/pi/rss2mqtt rss2mqtt
```

The installer creates a run-once `rss2mqtt.service` and enables `rss2mqtt.timer`. By default, the timer runs hourly from 08:00 through 20:00 local device time.

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
