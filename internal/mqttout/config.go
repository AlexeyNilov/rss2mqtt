package mqttout

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	DefaultEnvPath  = ".env"
	defaultClientID = "rss2mqtt"
	defaultTimeout  = 10 * time.Second
)

type Config struct {
	BrokerURL string
	Topic     string
	ClientID  string
	Timeout   time.Duration
}

func LoadConfig(path string) (Config, error) {
	values, err := readEnvFile(path)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		BrokerURL: values["MQTT_BROKER_URL"],
		Topic:     values["MQTT_TOPIC"],
		ClientID:  valueOrDefault(values["MQTT_CLIENT_ID"], defaultClientID),
		Timeout:   defaultTimeout,
	}
	if err := validateConfig(cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func readEnvFile(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("read env file: %w", err)
	}
	defer file.Close()

	values := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return nil, fmt.Errorf("parse env file: invalid line %q", line)
		}

		values[strings.TrimSpace(key)] = strings.Trim(strings.TrimSpace(value), `"'`)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read env file: %w", err)
	}

	return values, nil
}

func validateConfig(cfg Config) error {
	if strings.TrimSpace(cfg.BrokerURL) == "" {
		return fmt.Errorf("MQTT_BROKER_URL is required")
	}
	if strings.TrimSpace(cfg.Topic) == "" {
		return fmt.Errorf("MQTT_TOPIC is required")
	}

	return nil
}

func valueOrDefault(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}

	return value
}
