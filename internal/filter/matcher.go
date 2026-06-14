package filter

import (
	"strings"

	"github.com/AlexeyNilov/rss2mqtt/internal/discovery"
)

func Matches(item discovery.Item, filters []string) bool {
	text := strings.ToLower(item.Title + "\n" + item.Description)

	for _, filter := range filters {
		normalized := strings.ToLower(strings.TrimSpace(filter))
		if normalized == "" {
			continue
		}
		if normalized == "*" {
			return true
		}
		if strings.Contains(text, normalized) {
			return true
		}
	}

	return false
}
