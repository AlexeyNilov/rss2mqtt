package githubtrending

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/AlexeyNilov/rss2mqtt/internal/discovery"
	"github.com/PuerkitoBio/goquery"
)

const (
	defaultHTTPTimeout = 15 * time.Second
	githubBaseURL      = "https://github.com"
)

type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

type HTTPSource struct {
	Client HTTPDoer
}

func DefaultHTTPClient() *http.Client {
	return &http.Client{Timeout: defaultHTTPTimeout}
}

func (s HTTPSource) Items(ctx context.Context, cfg discovery.Source) ([]discovery.Item, error) {
	client := s.Client
	if client == nil {
		client = DefaultHTTPClient()
	}

	body, err := Fetch(ctx, client, cfg.URL)
	if err != nil {
		return nil, err
	}

	return Parse(strings.NewReader(string(body)), cfg.Name)
}

func Fetch(ctx context.Context, client HTTPDoer, pageURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create GitHub Trending request: %w", err)
	}
	req.Header.Set("User-Agent", "githubtrending2mqtt")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch GitHub Trending page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("fetch GitHub Trending page: unexpected status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read GitHub Trending response: %w", err)
	}

	return body, nil
}

func Parse(reader io.Reader, sourceName string) ([]discovery.Item, error) {
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return nil, fmt.Errorf("parse GitHub Trending HTML: %w", err)
	}

	items := make([]discovery.Item, 0)
	doc.Find("article.Box-row").Each(func(_ int, row *goquery.Selection) {
		item, ok := parseRepository(row, sourceName)
		if ok {
			items = append(items, item)
		}
	})

	return items, nil
}

func parseRepository(row *goquery.Selection, sourceName string) (discovery.Item, bool) {
	link := row.Find("h2 a[href]").First()
	href, ok := link.Attr("href")
	if !ok {
		return discovery.Item{}, false
	}

	fullName, ok := repositoryFullName(href)
	if !ok {
		return discovery.Item{}, false
	}

	return discovery.Item{
		SourceName:  sourceName,
		Title:       fullName,
		Description: repositoryDescription(row),
		Link:        githubBaseURL + "/" + fullName,
		Identity:    strings.ToLower(fullName),
	}, true
}

func repositoryFullName(href string) (string, bool) {
	parts := strings.Split(strings.Trim(href, "/"), "/")
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return "", false
	}

	return parts[0] + "/" + parts[1], true
}

func repositoryDescription(row *goquery.Selection) string {
	parts := make([]string, 0, 2)
	if description := cleanText(row.Find("p").First().Text()); description != "" {
		parts = append(parts, description)
	}
	if language := cleanText(row.Find("[itemprop='programmingLanguage']").First().Text()); language != "" {
		parts = append(parts, "Language: "+language)
	}

	return strings.Join(parts, "\n")
}

func cleanText(text string) string {
	return strings.Join(strings.Fields(text), " ")
}
