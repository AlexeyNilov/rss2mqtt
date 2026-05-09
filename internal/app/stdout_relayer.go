package app

import (
	"context"
	"io"

	"github.com/AlexeyNilov/rss2mqtt/internal/feed"
	"github.com/AlexeyNilov/rss2mqtt/internal/output"
)

type stdoutRelayer struct {
	writer io.Writer
}

func (r stdoutRelayer) Publish(ctx context.Context, item feed.Item) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	return output.WriteItem(r.writer, item)
}
