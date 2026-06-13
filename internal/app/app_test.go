package app

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AlexeyNilov/rss2mqtt/internal/config"
	"github.com/AlexeyNilov/rss2mqtt/internal/feed"
)

func TestRunPrintsApprovedNonDuplicateItemsAndSavesState(t *testing.T) {
	configPath := writeConfig(t, `
- name: oreilly-radar
  url: https://example.com/rss.xml
  filters:
    - AI
`)
	source := fakeSource{
		items: map[string][]feed.Item{
			"oreilly-radar": {
				{FeedName: "oreilly-radar", Title: "Local AI", Identity: "item-1"},
				{FeedName: "oreilly-radar", Title: "Gardening", Identity: "item-2"},
				{FeedName: "oreilly-radar", Title: "AI duplicate", Identity: "item-3"},
			},
		},
	}
	state := newFakeState()
	state.seen["oreilly-radar:item-3"] = true
	var stdout bytes.Buffer
	relayer := newFakeRelayer(&stdout)

	err := Run(context.Background(), Options{
		ConfigPath: configPath,
		State:      state,
		Source:     source,
		Relayer:    relayer,
		Stderr:     &bytes.Buffer{},
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	got := stdout.String()
	if !strings.Contains(got, "Title: Local AI") {
		t.Fatalf("stdout = %q, want approved new item", got)
	}
	if strings.Contains(got, "Gardening") {
		t.Fatalf("stdout = %q, want filtered item omitted", got)
	}
	if strings.Contains(got, "AI duplicate") {
		t.Fatalf("stdout = %q, want duplicate item omitted", got)
	}
	if !state.marked["oreilly-radar:item-1"] {
		t.Fatal("new approved item was not marked")
	}
	if state.marked["oreilly-radar:item-2"] {
		t.Fatal("filtered item was marked")
	}
	if !state.saved {
		t.Fatal("state was not saved")
	}
}

func TestRunLogsNewApprovedItemTitlesWhenDiscoveryLogConfigured(t *testing.T) {
	configPath := writeConfig(t, `
- name: oreilly-radar
  url: https://example.com/rss.xml
  filters:
    - AI
`)
	source := fakeSource{
		items: map[string][]feed.Item{
			"oreilly-radar": {
				{FeedName: "oreilly-radar", Title: "Local AI", Identity: "item-1"},
				{FeedName: "oreilly-radar", Title: "Gardening", Identity: "item-2"},
				{FeedName: "oreilly-radar", Title: "AI duplicate", Identity: "item-3"},
			},
		},
	}
	state := newFakeState()
	state.seen["oreilly-radar:item-3"] = true
	var output bytes.Buffer
	var discoveryLog bytes.Buffer

	err := Run(context.Background(), Options{
		ConfigPath:   configPath,
		State:        state,
		Source:       source,
		Relayer:      newFakeRelayer(&output),
		DiscoveryLog: &discoveryLog,
		Stderr:       &bytes.Buffer{},
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if got := discoveryLog.String(); got != "Local AI\n" {
		t.Fatalf("discovery log = %q, want only new approved title", got)
	}
}

func TestRunContinuesWhenOneFeedFails(t *testing.T) {
	configPath := writeConfig(t, `
- name: broken
  url: https://example.com/broken.xml
  filters:
    - AI
- name: working
  url: https://example.com/working.xml
  filters:
    - Go
`)
	source := fakeSource{
		items: map[string][]feed.Item{
			"working": {{FeedName: "working", Title: "Go release", Identity: "go-1"}},
		},
		errs: map[string]error{
			"broken": errors.New("network failed"),
		},
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	relayer := newFakeRelayer(&stdout)

	err := Run(context.Background(), Options{
		ConfigPath: configPath,
		State:      newFakeState(),
		Source:     source,
		Relayer:    relayer,
		Stderr:     &stderr,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !strings.Contains(stdout.String(), "Title: Go release") {
		t.Fatalf("stdout = %q, want working feed item", stdout.String())
	}
	if strings.Contains(stdout.String(), "broken") {
		t.Fatalf("stdout = %q, want diagnostics excluded from stdout", stdout.String())
	}
	if !strings.Contains(stderr.String(), "broken") {
		t.Fatalf("stderr = %q, want broken feed diagnostic", stderr.String())
	}
}

func TestRunReturnsErrorForInvalidConfig(t *testing.T) {
	configPath := writeConfig(t, `
- name: broken
  filters:
    - AI
`)

	err := Run(context.Background(), Options{
		ConfigPath: configPath,
		State:      newFakeState(),
		Source:     fakeSource{},
		Stdout:     &bytes.Buffer{},
		Stderr:     &bytes.Buffer{},
	})
	if err == nil {
		t.Fatal("Run() error = nil, want config error")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "config") {
		t.Fatalf("Run() error = %q, want config context", err)
	}
}

func TestRunReturnsErrorWhenStateCannotLoad(t *testing.T) {
	configPath := writeConfig(t, `
- name: oreilly-radar
  url: https://example.com/rss.xml
  filters:
    - AI
`)
	statePath := filepath.Join(t.TempDir(), ".rss2mqtt-state.json")
	if err := os.WriteFile(statePath, []byte("{not-json"), 0o600); err != nil {
		t.Fatalf("write corrupt state: %v", err)
	}

	err := Run(context.Background(), Options{
		ConfigPath: configPath,
		StatePath:  statePath,
		Source:     fakeSource{},
		Stdout:     &bytes.Buffer{},
		Stderr:     &bytes.Buffer{},
	})
	if err == nil {
		t.Fatal("Run() error = nil, want state load error")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "state") {
		t.Fatalf("Run() error = %q, want state context", err)
	}
}

func TestRunDoesNotMarkItemWhenOutputFails(t *testing.T) {
	configPath := writeConfig(t, `
- name: oreilly-radar
  url: https://example.com/rss.xml
  filters:
    - AI
`)
	source := fakeSource{
		items: map[string][]feed.Item{
			"oreilly-radar": {{FeedName: "oreilly-radar", Title: "Local AI", Identity: "item-1"}},
		},
	}
	state := newFakeState()
	relayer := &fakeRelayer{err: errors.New("publish failed")}
	var discoveryLog bytes.Buffer

	err := Run(context.Background(), Options{
		ConfigPath:   configPath,
		State:        state,
		Source:       source,
		Relayer:      relayer,
		DiscoveryLog: &discoveryLog,
		Stderr:       &bytes.Buffer{},
	})
	if err == nil {
		t.Fatal("Run() error = nil, want output error")
	}
	if state.marked["oreilly-radar:item-1"] {
		t.Fatal("item was marked despite output failure")
	}
	if discoveryLog.String() != "" {
		t.Fatalf("discovery log = %q, want no title logged after output failure", discoveryLog.String())
	}
}

func TestRunDoesNotMarkItemWhenDiscoveryLogFails(t *testing.T) {
	configPath := writeConfig(t, `
- name: oreilly-radar
  url: https://example.com/rss.xml
  filters:
    - AI
`)
	source := fakeSource{
		items: map[string][]feed.Item{
			"oreilly-radar": {{FeedName: "oreilly-radar", Title: "Local AI", Identity: "item-1"}},
		},
	}
	state := newFakeState()
	var output bytes.Buffer

	err := Run(context.Background(), Options{
		ConfigPath:   configPath,
		State:        state,
		Source:       source,
		Relayer:      newFakeRelayer(&output),
		DiscoveryLog: failingWriter{err: errors.New("log failed")},
		Stderr:       &bytes.Buffer{},
	})
	if err == nil {
		t.Fatal("Run() error = nil, want discovery log error")
	}
	if state.marked["oreilly-radar:item-1"] {
		t.Fatal("item was marked despite discovery log failure")
	}
}

func TestRunReturnsErrorWhenStateSaveFails(t *testing.T) {
	configPath := writeConfig(t, `
- name: oreilly-radar
  url: https://example.com/rss.xml
  filters:
    - AI
`)
	source := fakeSource{
		items: map[string][]feed.Item{
			"oreilly-radar": {{FeedName: "oreilly-radar", Title: "Local AI", Identity: "item-1"}},
		},
	}
	state := newFakeState()
	state.saveErr = errors.New("disk full")
	var stdout bytes.Buffer

	err := Run(context.Background(), Options{
		ConfigPath: configPath,
		State:      state,
		Source:     source,
		Relayer:    newFakeRelayer(&stdout),
		Stderr:     &bytes.Buffer{},
	})
	if err == nil {
		t.Fatal("Run() error = nil, want save error")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "save state") {
		t.Fatalf("Run() error = %q, want save state context", err)
	}
}

type fakeSource struct {
	items map[string][]feed.Item
	errs  map[string]error
}

func (s fakeSource) Items(_ context.Context, cfg config.Feed) ([]feed.Item, error) {
	if err := s.errs[cfg.Name]; err != nil {
		return nil, err
	}

	return s.items[cfg.Name], nil
}

type fakeState struct {
	seen    map[string]bool
	marked  map[string]bool
	saved   bool
	saveErr error
}

func newFakeState() *fakeState {
	return &fakeState{seen: make(map[string]bool), marked: make(map[string]bool)}
}

func (s *fakeState) Seen(feedName, identity string) bool {
	return s.seen[stateKey(feedName, identity)]
}

func (s *fakeState) Mark(feedName, identity string) {
	key := stateKey(feedName, identity)
	s.marked[key] = true
	s.seen[key] = true
}

func (s *fakeState) Save() error {
	s.saved = true
	return s.saveErr
}

type fakeRelayer struct {
	writer *bytes.Buffer
	err    error
}

func newFakeRelayer(writer *bytes.Buffer) *fakeRelayer {
	return &fakeRelayer{writer: writer}
}

func (r *fakeRelayer) Publish(_ context.Context, item feed.Item) error {
	if r.err != nil {
		return r.err
	}
	r.writer.WriteString("Title: " + item.Title + "\n")
	return nil
}

type failingWriter struct {
	err error
}

func (w failingWriter) Write([]byte) (int, error) {
	return 0, w.err
}

func stateKey(feedName, identity string) string {
	return feedName + ":" + identity
}

func writeConfig(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "rss.yaml")
	if err := os.WriteFile(path, []byte(strings.TrimSpace(content)+"\n"), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	return path
}
