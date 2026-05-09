package main

import (
	"context"
	"log"
	"os"

	"github.com/AlexeyNilov/rss2mqtt/internal/app"
	"github.com/AlexeyNilov/rss2mqtt/internal/mqttout"
)

func main() {
	mqttConfig, err := mqttout.LoadConfig(mqttout.DefaultEnvPath)
	if err != nil {
		log.New(os.Stderr, "", 0).Printf("rss2mqtt: load mqtt config: %v", err)
		os.Exit(1)
	}

	if err := app.Run(context.Background(), app.Options{
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
		Relayer: mqttout.NewPublisher(mqttConfig),
	}); err != nil {
		log.New(os.Stderr, "", 0).Printf("rss2mqtt: %v", err)
		os.Exit(1)
	}
}
