package main

import (
	"context"
	"fmt"
	"os"

	"github.com/AlexeyNilov/rss2mqtt/internal/app"
)

func main() {
	if err := app.Run(context.Background(), app.Options{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "rss2mqtt: %v\n", err)
		os.Exit(1)
	}
}
