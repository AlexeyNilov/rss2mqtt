package output

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/AlexeyNilov/rss2mqtt/internal/feed"
)

func TestWriteItemFormatsApprovedItem(t *testing.T) {
	var buf bytes.Buffer
	item := feed.Item{
		FeedName:    "oreilly-radar",
		Title:       "Fighting Tool Sprawl",
		Link:        "https://example.com/article",
		Description: "As enterprise AI agent adoption scales, tool infrastructure matters.",
		Published:   time.Date(2026, 5, 8, 11, 20, 31, 0, time.UTC),
	}

	if err := WriteItem(&buf, item); err != nil {
		t.Fatalf("WriteItem() error = %v", err)
	}

	got := buf.String()
	for _, want := range []string{
		"Feed: oreilly-radar",
		"Title: Fighting Tool Sprawl",
		"Link: https://example.com/article",
		"Published: 2026-05-08T11:20:31Z",
		"Description: As enterprise AI agent adoption scales, tool infrastructure matters.",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("formatted item = %q, want it to contain %q", got, want)
		}
	}
}

func TestWriteItemOmitsEmptyOptionalFields(t *testing.T) {
	var buf bytes.Buffer
	item := feed.Item{
		FeedName: "oreilly-radar",
		Title:    "Local AI",
	}

	if err := WriteItem(&buf, item); err != nil {
		t.Fatalf("WriteItem() error = %v", err)
	}

	got := buf.String()
	if strings.Contains(got, "Link:") {
		t.Fatalf("formatted item = %q, want empty link omitted", got)
	}
	if strings.Contains(got, "Published:") {
		t.Fatalf("formatted item = %q, want zero published omitted", got)
	}
	if strings.Contains(got, "Description:") {
		t.Fatalf("formatted item = %q, want empty description omitted", got)
	}
}

func TestWriteItemTruncatesLongDescription(t *testing.T) {
	var buf bytes.Buffer
	item := feed.Item{
		FeedName:    "oreilly-radar",
		Title:       "Long article",
		Description: strings.Repeat("a", 260),
	}

	if err := WriteItem(&buf, item); err != nil {
		t.Fatalf("WriteItem() error = %v", err)
	}

	got := buf.String()
	if len(got) > 360 {
		t.Fatalf("formatted item length = %d, want concise excerpt", len(got))
	}
	if !strings.Contains(got, "...") {
		t.Fatalf("formatted item = %q, want truncation marker", got)
	}
}

func TestWriteItemPropagatesWriterError(t *testing.T) {
	errWriter := failingWriter{err: errors.New("write failed")}

	err := WriteItem(errWriter, feed.Item{FeedName: "feed", Title: "title"})
	if err == nil {
		t.Fatal("WriteItem() error = nil, want writer error")
	}
	if !strings.Contains(err.Error(), "write") {
		t.Fatalf("WriteItem() error = %q, want write context", err)
	}
}

type failingWriter struct {
	err error
}

func (w failingWriter) Write([]byte) (int, error) {
	return 0, w.err
}
