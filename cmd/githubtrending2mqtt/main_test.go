package main

import (
	"bytes"
	"testing"

	"github.com/AlexeyNilov/rss2mqtt/internal/githubtrending"
)

func TestParseCLIOptionsDefaultsToMQTTAndDefaultConfig(t *testing.T) {
	opts, err := parseCLIOptions([]string{})
	if err != nil {
		t.Fatalf("parseCLIOptions() error = %v", err)
	}

	if opts.output != outputMQTT {
		t.Fatalf("output = %q, want %q", opts.output, outputMQTT)
	}
	if opts.configPath != githubtrending.DefaultConfigPath {
		t.Fatalf("configPath = %q, want %q", opts.configPath, githubtrending.DefaultConfigPath)
	}
}

func TestParseCLIOptionsAcceptsStdoutAndConfigPath(t *testing.T) {
	opts, err := parseCLIOptions([]string{"--output", "stdout", "--config", "custom.yaml"})
	if err != nil {
		t.Fatalf("parseCLIOptions() error = %v", err)
	}

	if opts.output != outputStdout {
		t.Fatalf("output = %q, want %q", opts.output, outputStdout)
	}
	if opts.configPath != "custom.yaml" {
		t.Fatalf("configPath = %q, want custom.yaml", opts.configPath)
	}
}

func TestParseCLIOptionsRejectsInvalidOutput(t *testing.T) {
	_, err := parseCLIOptions([]string{"--output", "file"})
	if err == nil {
		t.Fatal("parseCLIOptions() error = nil, want invalid output error")
	}
}

func TestPageSourcesUseResolvedURLs(t *testing.T) {
	sources := pageSources(githubtrending.Config{Pages: []githubtrending.Page{
		{Name: "python-weekly", Language: "python", Since: "weekly", Filters: []string{"*"}},
	}})

	if len(sources) != 1 {
		t.Fatalf("len(sources) = %d, want 1", len(sources))
	}
	if sources[0].Name != "python-weekly" {
		t.Fatalf("source name = %q", sources[0].Name)
	}
	if sources[0].URL != "https://github.com/trending/python?since=weekly" {
		t.Fatalf("source URL = %q", sources[0].URL)
	}
}

func TestDiscoveryLogUsesStdoutForMQTTOutput(t *testing.T) {
	var stdout bytes.Buffer

	log := discoveryLog(cliOptions{output: outputMQTT}, &stdout)

	if log != &stdout {
		t.Fatal("discoveryLog() did not return stdout for mqtt output")
	}
}

func TestDiscoveryLogDisabledForStdoutOutput(t *testing.T) {
	var stdout bytes.Buffer

	log := discoveryLog(cliOptions{output: outputStdout}, &stdout)

	if log != nil {
		t.Fatalf("discoveryLog() = %v, want nil for stdout output", log)
	}
}
