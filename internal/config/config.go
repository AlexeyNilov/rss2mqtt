package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Feeds []Feed
}

type Feed struct {
	Name    string   `yaml:"name"`
	URL     string   `yaml:"url"`
	Filters []string `yaml:"filters"`
}

func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	var feeds []Feed
	if err := yaml.Unmarshal(data, &feeds); err != nil {
		return Config{}, fmt.Errorf("parse yaml config: %w", err)
	}

	cfg := Config{Feeds: feeds}
	if err := validate(cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func validate(cfg Config) error {
	if len(cfg.Feeds) == 0 {
		return fmt.Errorf("config must contain at least one feed")
	}

	seen := make(map[string]struct{}, len(cfg.Feeds))
	for i, feed := range cfg.Feeds {
		if err := validateFeed(feed, i, seen); err != nil {
			return err
		}
	}

	return nil
}

func validateFeed(feed Feed, index int, seen map[string]struct{}) error {
	if strings.TrimSpace(feed.Name) == "" {
		return fmt.Errorf("feed %d: name is required", index)
	}
	if strings.TrimSpace(feed.URL) == "" {
		return fmt.Errorf("feed %q: url is required", feed.Name)
	}
	if len(feed.Filters) == 0 {
		return fmt.Errorf("feed %q: filters must not be empty", feed.Name)
	}
	if _, ok := seen[feed.Name]; ok {
		return fmt.Errorf("duplicate feed name %q", feed.Name)
	}

	seen[feed.Name] = struct{}{}
	return nil
}
