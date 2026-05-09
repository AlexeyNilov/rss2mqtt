package filter

import (
	"testing"

	"github.com/AlexeyNilov/rss2mqtt/internal/feed"
)

func TestMatchesApprovesWhenAnyFilterMatches(t *testing.T) {
	item := feed.Item{
		Title:       "Weekly Go release notes",
		Description: "Compiler and runtime changes",
	}

	if !Matches(item, []string{"python", "runtime"}) {
		t.Fatal("Matches() = false, want true when any filter substring matches")
	}
}

func TestMatchesApprovesEveryItemWhenWildcardFilterConfigured(t *testing.T) {
	item := feed.Item{
		Title:       "Gardening weekly",
		Description: "Soil and irrigation notes",
	}

	if !Matches(item, []string{"*"}) {
		t.Fatal("Matches() = false, want true when wildcard filter is configured")
	}
}

func TestMatchesIsCaseInsensitive(t *testing.T) {
	item := feed.Item{
		Title:       "Fighting Tool Sprawl",
		Description: "Enterprise AI agent adoption scales",
	}

	if !Matches(item, []string{"enterprise ai"}) {
		t.Fatal("Matches() = false, want true for case-insensitive description match")
	}
}

func TestMatchesSearchesTitleAndDescription(t *testing.T) {
	tests := []struct {
		name string
		item feed.Item
	}{
		{
			name: "title",
			item: feed.Item{Title: "Local AI", Description: "offline models"},
		},
		{
			name: "description",
			item: feed.Item{Title: "Local compute", Description: "AI models on device"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !Matches(tt.item, []string{"AI"}) {
				t.Fatal("Matches() = false, want true")
			}
		})
	}
}

func TestMatchesRejectsWhenNoFilterMatches(t *testing.T) {
	item := feed.Item{
		Title:       "Gardening weekly",
		Description: "Soil and irrigation notes",
	}

	if Matches(item, []string{"AI", "agent"}) {
		t.Fatal("Matches() = true, want false when no filter matches")
	}
}

func TestMatchesRejectsEmptyFilters(t *testing.T) {
	item := feed.Item{
		Title:       "Local AI",
		Description: "AI models on device",
	}

	if Matches(item, nil) {
		t.Fatal("Matches() = true, want false for empty filters")
	}
}

func TestMatchesIgnoresBlankFilters(t *testing.T) {
	item := feed.Item{
		Title:       "Local AI",
		Description: "AI models on device",
	}

	if Matches(item, []string{" ", "\t"}) {
		t.Fatal("Matches() = true, want false for blank filters")
	}
}
