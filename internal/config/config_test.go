package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadReturnsConfiguredFeeds(t *testing.T) {
	path := writeConfig(t, `
- name: oreilly-radar
  url: https://www.oreilly.com/radar/feed/
  filters:
    - AI
    - agent
- name: go-blog
  url: https://go.dev/blog/feed.atom
  filters:
    - release
`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(cfg.Feeds) != 2 {
		t.Fatalf("len(cfg.Feeds) = %d, want 2", len(cfg.Feeds))
	}

	first := cfg.Feeds[0]
	if first.Name != "oreilly-radar" {
		t.Fatalf("first feed name = %q, want %q", first.Name, "oreilly-radar")
	}
	if first.URL != "https://www.oreilly.com/radar/feed/" {
		t.Fatalf("first feed URL = %q", first.URL)
	}
	if got := strings.Join(first.Filters, ","); got != "AI,agent" {
		t.Fatalf("first feed filters = %q, want %q", got, "AI,agent")
	}
}

func TestLoadRejectsInvalidFeedConfig(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr string
	}{
		{
			name: "missing feed name",
			content: `
- url: https://www.oreilly.com/radar/feed/
  filters:
    - AI
`,
			wantErr: "name",
		},
		{
			name: "missing URL",
			content: `
- name: oreilly-radar
  filters:
    - AI
`,
			wantErr: "url",
		},
		{
			name: "empty filters",
			content: `
- name: oreilly-radar
  url: https://www.oreilly.com/radar/feed/
  filters: []
`,
			wantErr: "filters",
		},
		{
			name: "blank filter",
			content: `
- name: oreilly-radar
  url: https://www.oreilly.com/radar/feed/
  filters:
    - " "
`,
			wantErr: "filter",
		},
		{
			name: "duplicate feed name",
			content: `
- name: oreilly-radar
  url: https://www.oreilly.com/radar/feed/
  filters:
    - AI
- name: oreilly-radar
  url: https://example.com/feed.xml
  filters:
    - go
`,
			wantErr: "duplicate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeConfig(t, tt.content)

			_, err := Load(path)
			if err == nil {
				t.Fatal("Load() error = nil, want validation error")
			}
			if !strings.Contains(strings.ToLower(err.Error()), tt.wantErr) {
				t.Fatalf("Load() error = %q, want it to contain %q", err, tt.wantErr)
			}
		})
	}
}

func TestLoadReportsMissingFile(t *testing.T) {
	_, err := Load(filepath.Join(t.TempDir(), "missing.yaml"))
	if err == nil {
		t.Fatal("Load() error = nil, want missing file error")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "read") {
		t.Fatalf("Load() error = %q, want read context", err)
	}
}

func TestLoadReportsInvalidYAML(t *testing.T) {
	path := writeConfig(t, `
- name: oreilly-radar
  url: https://www.oreilly.com/radar/feed/
  filters:
    - AI
  - broken
`)

	_, err := Load(path)
	if err == nil {
		t.Fatal("Load() error = nil, want YAML parse error")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "yaml") {
		t.Fatalf("Load() error = %q, want YAML context", err)
	}
}

func writeConfig(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "rss.yaml")
	if err := os.WriteFile(path, []byte(strings.TrimSpace(content)+"\n"), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	return path
}
