package main

import (
	"context"
	"log"
	"os"

	"github.com/AlexeyNilov/rss2mqtt/internal/app"
)

func main() {
	if err := app.Run(context.Background(), app.Options{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}); err != nil {
		log.New(os.Stderr, "", 0).Printf("rss2mqtt: %v", err)
		os.Exit(1)
	}
}
