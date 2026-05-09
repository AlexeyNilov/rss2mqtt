package main

import (
	"testing"
)

func TestParseCLIOptionsDefaultsToMQTT(t *testing.T) {
	opts, err := parseCLIOptions([]string{})
	if err != nil {
		t.Fatalf("parseCLIOptions() error = %v", err)
	}

	if opts.output != outputMQTT {
		t.Fatalf("output = %q, want %q", opts.output, outputMQTT)
	}
}

func TestParseCLIOptionsAcceptsStdout(t *testing.T) {
	opts, err := parseCLIOptions([]string{"--output", "stdout"})
	if err != nil {
		t.Fatalf("parseCLIOptions() error = %v", err)
	}

	if opts.output != outputStdout {
		t.Fatalf("output = %q, want %q", opts.output, outputStdout)
	}
}

func TestParseCLIOptionsRejectsInvalidOutput(t *testing.T) {
	_, err := parseCLIOptions([]string{"--output", "file"})
	if err == nil {
		t.Fatal("parseCLIOptions() error = nil, want invalid output error")
	}
}
