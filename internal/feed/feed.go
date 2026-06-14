package feed

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/AlexeyNilov/rss2mqtt/internal/discovery"
	"github.com/mmcdole/gofeed"
)

const defaultHTTPTimeout = 15 * time.Second

type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

type Item = discovery.Item

func DefaultHTTPClient() *http.Client {
	return &http.Client{Timeout: defaultHTTPTimeout}
}

func Fetch(ctx context.Context, client HTTPDoer, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create feed request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch feed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("fetch feed: unexpected status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read feed response: %w", err)
	}

	return body, nil
}

func Parse(reader io.Reader, feedName string) ([]Item, error) {
	parsed, err := gofeed.NewParser().Parse(reader)
	if err != nil {
		return nil, fmt.Errorf("parse feed: %w", err)
	}

	items := make([]Item, 0, len(parsed.Items))
	for _, parsedItem := range parsed.Items {
		items = append(items, normalizeItem(feedName, parsedItem))
	}

	return items, nil
}

func normalizeItem(feedName string, item *gofeed.Item) Item {
	normalized := Item{
		SourceName:  feedName,
		Title:       item.Title,
		Description: item.Description,
		Link:        item.Link,
		GUID:        item.GUID,
		Identity:    identity(item),
	}

	if item.PublishedParsed != nil {
		normalized.Published = *item.PublishedParsed
	}

	return normalized
}

func identity(item *gofeed.Item) string {
	switch {
	case item.GUID != "":
		return item.GUID
	case item.Link != "":
		return item.Link
	default:
		return item.Title
	}
}
