# rss2mqtt

RSS to MQTT relay.

## MVP

The first version is a small Go command-line application that reads RSS feed definitions from a local `rss.yaml` file, filters feed items by case-insensitive substrings, prints approved items to stdout, and exits. MQTT output is planned for a later phase.

The application is intended to run on Raspberry Pi Zero 2 class hardware. Scheduling is external: the MVP runs once per invocation and can later be launched by a systemd timer.

## Configuration

Create a local `rss.yaml` file next to the binary or run the binary from a working directory that contains it:

```yaml
- name: oreilly-radar
  url: https://www.oreilly.com/radar/feed/
  filters:
    - AI
    - agent
```

`name` must be unique per feed. Filters are matched against the RSS item title and `description`.

See `rss.yaml.example` for a sample file.

## Development

Run the test suite:

```powershell
go test ./...
```

Build the local binary:

```powershell
go build ./cmd/rss2mqtt
```

## Raspberry Pi Service Timer

After copying the `linux/arm64` binary and creating `rss.yaml` on the Raspberry Pi, install the systemd service and timer:

```bash
scripts/setup_service_pi.sh raspberrypi.local pi /home/pi/rss2mqtt rss2mqtt
```

The installer creates a run-once `rss2mqtt.service` and enables `rss2mqtt.timer`. By default, the timer runs hourly from 08:00 through 20:00 local device time:

```bash
TIMER_ON_CALENDAR='*-*-* 08..20:00:00' scripts/setup_service_pi.sh raspberrypi.local
```

Use `--print-only` to inspect the generated units without installing them.

To deploy a new binary after the timer is installed:

```bash
scripts/build_arm.sh
scripts/deploy_pi.sh raspberrypi.local pi /home/pi/rss2mqtt/rss2mqtt rss2mqtt
```

The deploy script copies the binary, sets executable mode, and restarts the timer if it is already installed. It does not stop or start the run-once service directly.
