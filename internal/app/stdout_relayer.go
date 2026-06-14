package app

import (
	"context"
	"io"

	"github.com/AlexeyNilov/rss2mqtt/internal/discovery"
	"github.com/AlexeyNilov/rss2mqtt/internal/output"
)

type stdoutRelayer struct {
	writer io.Writer
}

func NewStdoutRelayer(writer io.Writer) Relayer {
	return stdoutRelayer{writer: writer}
}

func (r stdoutRelayer) Publish(ctx context.Context, item discovery.Item) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	return output.WriteItem(r.writer, item)
}
