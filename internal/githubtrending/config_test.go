package githubtrending

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadReturnsConfiguredPages(t *testing.T) {
	path := writeConfig(t, `
- name: python-weekly
  language: python
  since: weekly
  filters:
    - "*"
- name: all-monthly
  since: monthly
  filters:
    - database
`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(cfg.Pages) != 2 {
		t.Fatalf("len(cfg.Pages) = %d, want 2", len(cfg.Pages))
	}
	if got := cfg.Pages[0].ResolvedURL(); got != "https://github.com/trending/python?since=weekly" {
		t.Fatalf("ResolvedURL() = %q", got)
	}
	if got := cfg.Pages[1].ResolvedURL(); got != "https://github.com/trending?since=monthly" {
		t.Fatalf("ResolvedURL() = %q", got)
	}
}

func TestLoadUsesExplicitURL(t *testing.T) {
	path := writeConfig(t, `
- name: custom
  url: https://github.com/trending/go?since=daily
  filters:
    - cli
`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if got := cfg.Pages[0].ResolvedURL(); got != "https://github.com/trending/go?since=daily" {
		t.Fatalf("ResolvedURL() = %q", got)
	}
}

func TestLoadRejectsInvalidPageConfig(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr string
	}{
		{
			name: "missing name",
			content: `
- language: python
  filters:
    - "*"
`,
			wantErr: "name",
		},
		{
			name: "bad since",
			content: `
- name: python
  language: python
  since: yearly
  filters:
    - "*"
`,
			wantErr: "since",
		},
		{
			name: "empty filters",
			content: `
- name: python
  language: python
  filters: []
`,
			wantErr: "filters",
		},
		{
			name: "duplicate name",
			content: `
- name: python
  language: python
  filters:
    - "*"
- name: python
  language: go
  filters:
    - "*"
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

func writeConfig(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "github-trending.yaml")
	if err := os.WriteFile(path, []byte(strings.TrimSpace(content)+"\n"), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	return path
}
