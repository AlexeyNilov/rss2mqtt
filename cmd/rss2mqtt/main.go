package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/AlexeyNilov/rss2mqtt/internal/app"
	"github.com/AlexeyNilov/rss2mqtt/internal/mqttout"
)

const (
	outputMQTT   = "mqtt"
	outputStdout = "stdout"
)

type cliOptions struct {
	output string
}

func main() {
	if err := run(context.Background(), os.Args[1:], os.Stdout, os.Stderr); err != nil {
		log.New(os.Stderr, "", 0).Printf("rss2mqtt: %v", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	cliOpts, err := parseCLIOptions(args)
	if err != nil {
		return err
	}

	relayer, err := buildRelayer(cliOpts, stdout)
	if err != nil {
		return err
	}

	return app.Run(ctx, app.Options{
		Stdout:       stdout,
		Stderr:       stderr,
		DiscoveryLog: discoveryLog(cliOpts, stdout),
		Relayer:      relayer,
	})
}

func parseCLIOptions(args []string) (cliOptions, error) {
	flags := flag.NewFlagSet("rss2mqtt", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	opts := cliOptions{}
	flags.StringVar(&opts.output, "output", outputMQTT, "output target: mqtt or stdout")

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

func discoveryLog(opts cliOptions, stdout io.Writer) io.Writer {
	if opts.output == outputMQTT {
		return stdout
	}

	return nil
}
