package discovery

import "time"

type Source struct {
	Name    string
	URL     string
	Filters []string
}

type Item struct {
	SourceName  string
	Title       string
	Description string
	Link        string
	GUID        string
	Identity    string
	Published   time.Time
}
