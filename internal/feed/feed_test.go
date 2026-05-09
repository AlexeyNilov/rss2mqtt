package feed

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestParseReadsSampleFeedItems(t *testing.T) {
	file := openSampleFeed(t)
	defer file.Close()

	items, err := Parse(file, "oreilly-radar")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(items) != 15 {
		t.Fatalf("len(items) = %d, want 15", len(items))
	}

	first := items[0]
	if first.FeedName != "oreilly-radar" {
		t.Fatalf("FeedName = %q, want %q", first.FeedName, "oreilly-radar")
	}
	if first.Title != "Fighting Tool Sprawl: The Case for AI Tool Registries" {
		t.Fatalf("Title = %q", first.Title)
	}
	if first.Link != "https://www.oreilly.com/radar/fighting-tool-sprawl-the-case-for-ai-tool-registries/" {
		t.Fatalf("Link = %q", first.Link)
	}
	if first.GUID != "https://www.oreilly.com/radar/?p=18667" {
		t.Fatalf("GUID = %q", first.GUID)
	}
	if first.Identity != first.GUID {
		t.Fatalf("Identity = %q, want GUID %q", first.Identity, first.GUID)
	}
	if first.Published.IsZero() || first.Published.Year() != 2026 {
		t.Fatalf("Published = %v, want parsed 2026 timestamp", first.Published)
	}
}

func TestParseUsesDescriptionAsItemDescription(t *testing.T) {
	file := openSampleFeed(t)
	defer file.Close()

	items, err := Parse(file, "oreilly-radar")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	description := items[0].Description
	if !strings.Contains(description, "As enterprise AI agent adoption scales") {
		t.Fatalf("Description = %q, want RSS description text", description)
	}
	if strings.Contains(description, "<p>") {
		t.Fatalf("Description contains full content markup, want RSS description excerpt: %q", description)
	}
}

func TestParseReportsInvalidFeed(t *testing.T) {
	_, err := Parse(strings.NewReader("<rss><channel><item>"), "broken")
	if err == nil {
		t.Fatal("Parse() error = nil, want parse error")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "parse") {
		t.Fatalf("Parse() error = %q, want parse context", err)
	}
}

func TestFetchReturnsResponseBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", r.Method)
		}
		_, _ = w.Write([]byte("<rss></rss>"))
	}))
	defer server.Close()

	body, err := Fetch(context.Background(), server.Client(), server.URL)
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if string(body) != "<rss></rss>" {
		t.Fatalf("Fetch() body = %q", body)
	}
}

func TestFetchReportsNonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unavailable", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	_, err := Fetch(context.Background(), server.Client(), server.URL)
	if err == nil {
		t.Fatal("Fetch() error = nil, want status error")
	}
	if !strings.Contains(err.Error(), "503") {
		t.Fatalf("Fetch() error = %q, want status code", err)
	}
}

func TestDefaultHTTPClientHasTimeout(t *testing.T) {
	client := DefaultHTTPClient()
	if client.Timeout <= 0 {
		t.Fatal("DefaultHTTPClient().Timeout = 0, want conservative timeout")
	}
	if client.Timeout > 30*time.Second {
		t.Fatalf("DefaultHTTPClient().Timeout = %v, want <= 30s", client.Timeout)
	}
}

func openSampleFeed(t *testing.T) *os.File {
	t.Helper()

	path := filepath.Join("..", "..", "sample", "feed.xml")
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("open sample feed: %v", err)
	}

	return file
}
