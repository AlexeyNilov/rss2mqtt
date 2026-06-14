package githubtrending

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	DefaultConfigPath = "github-trending.yaml"
	DefaultStatePath  = ".githubtrending2mqtt-state.json"
	defaultSince      = "weekly"
)

type Config struct {
	Pages []Page
}

type Page struct {
	Name     string   `yaml:"name"`
	URL      string   `yaml:"url"`
	Language string   `yaml:"language"`
	Since    string   `yaml:"since"`
	Filters  []string `yaml:"filters"`
}

func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	var pages []Page
	if err := yaml.Unmarshal(data, &pages); err != nil {
		return Config{}, fmt.Errorf("parse yaml config: %w", err)
	}

	cfg := Config{Pages: pages}
	if err := validate(cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (p Page) ResolvedURL() string {
	if strings.TrimSpace(p.URL) != "" {
		return strings.TrimSpace(p.URL)
	}

	since := normalizedSince(p.Since)
	language := strings.Trim(strings.TrimSpace(p.Language), "/")
	if language == "" {
		return "https://github.com/trending?since=" + url.QueryEscape(since)
	}

	return "https://github.com/trending/" + url.PathEscape(language) + "?since=" + url.QueryEscape(since)
}

func validate(cfg Config) error {
	if len(cfg.Pages) == 0 {
		return fmt.Errorf("config must contain at least one page")
	}

	seen := make(map[string]struct{}, len(cfg.Pages))
	for i, page := range cfg.Pages {
		if err := validatePage(page, i, seen); err != nil {
			return err
		}
	}

	return nil
}

func validatePage(page Page, index int, seen map[string]struct{}) error {
	if strings.TrimSpace(page.Name) == "" {
		return fmt.Errorf("page %d: name is required", index)
	}
	if !validSince(normalizedSince(page.Since)) {
		return fmt.Errorf("page %q: since must be daily, weekly, or monthly", page.Name)
	}
	if len(page.Filters) == 0 {
		return fmt.Errorf("page %q: filters must not be empty", page.Name)
	}
	if hasBlankFilter(page.Filters) {
		return fmt.Errorf("page %q: filters must not contain blank values", page.Name)
	}
	if _, ok := seen[page.Name]; ok {
		return fmt.Errorf("duplicate page name %q", page.Name)
	}

	seen[page.Name] = struct{}{}
	return nil
}

func normalizedSince(since string) string {
	trimmed := strings.TrimSpace(since)
	if trimmed == "" {
		return defaultSince
	}

	return trimmed
}

func validSince(since string) bool {
	return since == "daily" || since == "weekly" || since == "monthly"
}

func hasBlankFilter(filters []string) bool {
	for _, filter := range filters {
		if strings.TrimSpace(filter) == "" {
			return true
		}
	}

	return false
}
