package app

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/AlexeyNilov/rss2mqtt/internal/config"
	"github.com/AlexeyNilov/rss2mqtt/internal/feed"
	"github.com/AlexeyNilov/rss2mqtt/internal/filter"
	"github.com/AlexeyNilov/rss2mqtt/internal/output"
	"github.com/AlexeyNilov/rss2mqtt/internal/state"
)

const DefaultConfigPath = "rss.yaml"

type Options struct {
	ConfigPath string
	StatePath  string
	Stdout     io.Writer
	Stderr     io.Writer
	Source     FeedSource
	State      DuplicateState
}

type FeedSource interface {
	Items(ctx context.Context, cfg config.Feed) ([]feed.Item, error)
}

type DuplicateState interface {
	Seen(feedName, identity string) bool
	Mark(feedName, identity string)
	Save() error
}

type HTTPFeedSource struct {
	Client feed.HTTPDoer
}

func Run(ctx context.Context, opts Options) error {
	opts = withDefaults(opts)

	cfg, err := config.Load(opts.ConfigPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	duplicateState, err := duplicateState(opts)
	if err != nil {
		return fmt.Errorf("load state: %w", err)
	}

	if err := processFeeds(ctx, cfg, opts.Source, duplicateState, opts.Stdout, opts.Stderr); err != nil {
		return err
	}
	if err := duplicateState.Save(); err != nil {
		return fmt.Errorf("save state: %w", err)
	}

	return nil
}

func (s HTTPFeedSource) Items(ctx context.Context, cfg config.Feed) ([]feed.Item, error) {
	client := s.Client
	if client == nil {
		client = feed.DefaultHTTPClient()
	}

	body, err := feed.Fetch(ctx, client, cfg.URL)
	if err != nil {
		return nil, err
	}

	return feed.Parse(bytes.NewReader(body), cfg.Name)
}

func withDefaults(opts Options) Options {
	if opts.ConfigPath == "" {
		opts.ConfigPath = DefaultConfigPath
	}
	if opts.StatePath == "" {
		opts.StatePath = state.DefaultPath
	}
	if opts.Stdout == nil {
		opts.Stdout = io.Discard
	}
	if opts.Stderr == nil {
		opts.Stderr = io.Discard
	}
	if opts.Source == nil {
		opts.Source = HTTPFeedSource{Client: feed.DefaultHTTPClient()}
	}

	return opts
}

func duplicateState(opts Options) (DuplicateState, error) {
	if opts.State != nil {
		return opts.State, nil
	}

	return state.Load(opts.StatePath)
}

func processFeeds(
	ctx context.Context,
	cfg config.Config,
	source FeedSource,
	duplicates DuplicateState,
	stdout io.Writer,
	stderr io.Writer,
) error {
	for _, configuredFeed := range cfg.Feeds {
		items, err := source.Items(ctx, configuredFeed)
		if err != nil {
			writeDiagnostic(stderr, configuredFeed.Name, err)
			continue
		}
		if err := processItems(configuredFeed, items, duplicates, stdout); err != nil {
			return err
		}
	}

	return nil
}

func processItems(configuredFeed config.Feed, items []feed.Item, duplicates DuplicateState, stdout io.Writer) error {
	for _, item := range items {
		if !filter.Matches(item, configuredFeed.Filters) {
			continue
		}
		if duplicates.Seen(configuredFeed.Name, item.Identity) {
			continue
		}
		if err := output.WriteItem(stdout, item); err != nil {
			return err
		}

		duplicates.Mark(configuredFeed.Name, item.Identity)
	}

	return nil
}

func writeDiagnostic(stderr io.Writer, feedName string, err error) {
	_, _ = fmt.Fprintf(stderr, "Feed %q failed: %v\n", feedName, err)
}
