package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/AlexeyNilov/rss2mqtt/internal/discovery"
)

const descriptionLimit = 160

func WriteItem(writer io.Writer, item discovery.Item) error {
	if _, err := fmt.Fprintf(writer, "Source: %s\n", item.SourceName); err != nil {
		return fmt.Errorf("write feed item: %w", err)
	}
	if _, err := fmt.Fprintf(writer, "Title: %s\n", item.Title); err != nil {
		return fmt.Errorf("write feed item: %w", err)
	}
	if item.Link != "" {
		if _, err := fmt.Fprintf(writer, "Link: %s\n", item.Link); err != nil {
			return fmt.Errorf("write feed item: %w", err)
		}
	}
	if !item.Published.IsZero() {
		if _, err := fmt.Fprintf(writer, "Published: %s\n", item.Published.UTC().Format("2006-01-02T15:04:05Z")); err != nil {
			return fmt.Errorf("write feed item: %w", err)
		}
	}
	if strings.TrimSpace(item.Description) != "" {
		if _, err := fmt.Fprintf(writer, "Description: %s\n", excerpt(item.Description)); err != nil {
			return fmt.Errorf("write feed item: %w", err)
		}
	}
	if _, err := fmt.Fprintln(writer); err != nil {
		return fmt.Errorf("write feed item: %w", err)
	}

	return nil
}

func excerpt(text string) string {
	trimmed := strings.Join(strings.Fields(text), " ")
	if len(trimmed) <= descriptionLimit {
		return trimmed
	}

	return strings.TrimSpace(trimmed[:descriptionLimit]) + "..."
}
