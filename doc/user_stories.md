# User stories

## User story format

Write each story in the form:

> **As a** `<user_type>`, **I want** `<goal>` **so that** `<benefit>`.

This captures:

* **Who** the user is
* **What** they want to achieve
* **Why** it matters

## Glossary

* Device: Raspberry Pi Zero 2.
* Feed: One RSS source configured by URL.
* Item: One message/article/entry received from an RSS feed.
* Filter: A configured list of substrings used to decide whether an RSS item should be relayed.
* Relay: The act of forwarding an approved RSS item to an output destination.
* MVP: First implementation stage that prints approved items to stdout instead of publishing them to MQTT.

## Actual stories

### MVP: stdout relay

* **As a** user running the application on a Raspberry Pi Zero 2, **I want** the application to read RSS feed definitions from a YAML config file **so that** I can change monitored feeds without rebuilding the program.
* **As a** user monitoring multiple RSS feeds, **I want** each configured feed to have its own substring filters **so that** different sources can be approved using source-specific criteria.
* **As a** user monitoring RSS feeds, **I want** the application to approve only RSS items that match the configured filters **so that** irrelevant items are not relayed.
* **As a** user running the application hourly, **I want** already processed RSS items to be suppressed **so that** the same item is not printed or relayed repeatedly.
* **As a** user testing the first version, **I want** approved RSS items to be printed to stdout **so that** filtering behavior can be verified before MQTT is added.
* **As a** user scheduling the application with systemd, **I want** the MVP to run once and exit **so that** an external timer can control when feed checks happen.
* **As a** user deploying to a small device, **I want** the application to stay lightweight **so that** it can run reliably on a Raspberry Pi Zero 2.

### Future: MQTT relay

* **As a** user integrating RSS updates with home automation or other MQTT consumers, **I want** approved RSS items to be published to a configured MQTT topic **so that** downstream systems can consume them.

## Open questions

* How should duplicate suppression state be stored and expired across scheduled runs?
