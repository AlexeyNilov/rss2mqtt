package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/AlexeyNilov/rss2mqtt/internal/app"
	"github.com/AlexeyNilov/rss2mqtt/internal/discovery"
	"github.com/AlexeyNilov/rss2mqtt/internal/githubtrending"
	"github.com/AlexeyNilov/rss2mqtt/internal/mqttout"
)

const (
	outputMQTT   = "mqtt"
	outputStdout = "stdout"
)

type cliOptions struct {
	output     string
	configPath string
}

func main() {
	if err := run(context.Background(), os.Args[1:], os.Stdout, os.Stderr); err != nil {
		log.New(os.Stderr, "", 0).Printf("githubtrending2mqtt: %v", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	cliOpts, err := parseCLIOptions(args)
	if err != nil {
		return err
	}

	cfg, err := githubtrending.Load(cliOpts.configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	relayer, err := buildRelayer(cliOpts, stdout)
	if err != nil {
		return err
	}

	return app.RunSources(ctx, pageSources(cfg), app.Options{
		StatePath:    githubtrending.DefaultStatePath,
		Stdout:       stdout,
		Stderr:       stderr,
		DiscoveryLog: discoveryLog(cliOpts, stdout),
		Source:       githubtrending.HTTPSource{Client: githubtrending.DefaultHTTPClient()},
		Relayer:      relayer,
	})
}

func parseCLIOptions(args []string) (cliOptions, error) {
	flags := flag.NewFlagSet("githubtrending2mqtt", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	opts := cliOptions{configPath: githubtrending.DefaultConfigPath}
	flags.StringVar(&opts.output, "output", outputMQTT, "output target: mqtt or stdout")
	flags.StringVar(&opts.configPath, "config", opts.configPath, "path to GitHub Trending config")

	if err := flags.Parse(args); err != nil {
		return cliOptions{}, err
	}
	if opts.output != outputMQTT && opts.output != outputStdout {
		return cliOptions{}, fmt.Errorf("invalid output %q: use %q or %q", opts.output, outputMQTT, outputStdout)
	}

	return opts, nil
}

func buildRelayer(opts cliOptions, stdout io.Writer) (app.Relayer, error) {
	if opts.output == outputStdout {
		return app.NewStdoutRelayer(stdout), nil
	}

	mqttConfig, err := mqttout.LoadConfig(mqttout.DefaultEnvPath)
	if err != nil {
		return nil, fmt.Errorf("load mqtt config: %w", err)
	}

	return mqttout.NewPublisher(mqttConfig), nil
}

func pageSources(cfg githubtrending.Config) []discovery.Source {
	sources := make([]discovery.Source, 0, len(cfg.Pages))
	for _, page := range cfg.Pages {
		sources = append(sources, discovery.Source{
			Name:    page.Name,
			URL:     page.ResolvedURL(),
			Filters: page.Filters,
		})
	}

	return sources
}

func discoveryLog(opts cliOptions, stdout io.Writer) io.Writer {
	if opts.output == outputMQTT {
		return stdout
	}

	return nil
}
