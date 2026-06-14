package githubtrending

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/AlexeyNilov/rss2mqtt/internal/discovery"
)

func TestParseReturnsTrendingRepositories(t *testing.T) {
	items, err := Parse(strings.NewReader(sampleTrendingHTML), "python-weekly")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("len(items) = %d, want 2", len(items))
	}

	first := items[0]
	if first.SourceName != "python-weekly" {
		t.Fatalf("SourceName = %q", first.SourceName)
	}
	if first.Title != "owner-one/project-alpha" {
		t.Fatalf("Title = %q", first.Title)
	}
	if first.Link != "https://github.com/owner-one/project-alpha" {
		t.Fatalf("Link = %q", first.Link)
	}
	if first.Identity != "owner-one/project-alpha" {
		t.Fatalf("Identity = %q", first.Identity)
	}
	if !strings.Contains(first.Description, "Fast alpha tooling") {
		t.Fatalf("Description = %q, want repository description", first.Description)
	}
	if !strings.Contains(first.Description, "Language: Python") {
		t.Fatalf("Description = %q, want language", first.Description)
	}
}

func TestParseSkipsRowsWithoutRepositoryLinks(t *testing.T) {
	items, err := Parse(strings.NewReader(`
<article class="Box-row"><h2>No link</h2></article>
<article class="Box-row"><h2><a href="/owner/repo">owner / repo</a></h2></article>
`), "mixed")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("len(items) = %d, want 1", len(items))
	}
}

func TestHTTPSourceFetchesAndParsesPage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", r.Method)
		}
		_, _ = w.Write([]byte(sampleTrendingHTML))
	}))
	defer server.Close()

	source := HTTPSource{Client: server.Client()}
	items, err := source.Items(context.Background(), discovery.Source{Name: "python-weekly", URL: server.URL})
	if err != nil {
		t.Fatalf("Items() error = %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("len(items) = %d, want 2", len(items))
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

const sampleTrendingHTML = `
<html>
  <body>
    <article class="Box-row">
      <h2>
        <a href="/owner-one/project-alpha">
          owner-one / project-alpha
        </a>
      </h2>
      <p>Fast alpha tooling for Python agents.</p>
      <span itemprop="programmingLanguage">Python</span>
      <a href="/owner-one/project-alpha/stargazers">1,234</a>
      <span>321 stars today</span>
    </article>
    <article class="Box-row">
      <h2>
        <a href="/owner-two/project-beta">owner-two / project-beta</a>
      </h2>
      <p>Repository beta description.</p>
      <span itemprop="programmingLanguage">Go</span>
    </article>
  </body>
</html>
`
