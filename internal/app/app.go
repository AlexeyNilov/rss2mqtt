package app

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"

	"github.com/AlexeyNilov/rss2mqtt/internal/config"
	"github.com/AlexeyNilov/rss2mqtt/internal/feed"
	"github.com/AlexeyNilov/rss2mqtt/internal/filter"
	"github.com/AlexeyNilov/rss2mqtt/internal/state"
)

const DefaultConfigPath = "rss.yaml"

type Options struct {
	ConfigPath   string
	StatePath    string
	Stdout       io.Writer
	Stderr       io.Writer
	DiscoveryLog io.Writer
	Source       FeedSource
	State        DuplicateState
	Relayer      Relayer
}

type FeedSource interface {
	Items(ctx context.Context, cfg config.Feed) ([]feed.Item, error)
}

type DuplicateState interface {
	Seen(feedName, identity string) bool
	Mark(feedName, identity string)
	Save() error
}

type Relayer interface {
	Publish(ctx context.Context, item feed.Item) error
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

	if err := processFeeds(ctx, cfg, opts.Source, duplicateState, opts.Relayer, opts.Stderr, opts.DiscoveryLog); err != nil {
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
	if opts.Relayer == nil {
		opts.Relayer = NewStdoutRelayer(opts.Stdout)
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
	relayer Relayer,
	stderr io.Writer,
	discoveryLog io.Writer,
) error {
	for _, configuredFeed := range cfg.Feeds {
		items, err := source.Items(ctx, configuredFeed)
		if err != nil {
			writeDiagnostic(stderr, configuredFeed.Name, err)
			continue
		}
		if err := processItems(ctx, configuredFeed, items, duplicates, relayer, discoveryLog); err != nil {
			return err
		}
	}

	return nil
}

func processItems(
	ctx context.Context,
	configuredFeed config.Feed,
	items []feed.Item,
	duplicates DuplicateState,
	relayer Relayer,
	discoveryLog io.Writer,
) error {
	for _, item := range items {
		if !filter.Matches(item, configuredFeed.Filters) {
			continue
		}
		if duplicates.Seen(configuredFeed.Name, item.Identity) {
			continue
		}
		if err := relayer.Publish(ctx, item); err != nil {
			return err
		}
		if err := writeDiscoveryTitle(discoveryLog, item); err != nil {
			return err
		}

		duplicates.Mark(configuredFeed.Name, item.Identity)
	}

	return nil
}

func writeDiagnostic(stderr io.Writer, feedName string, err error) {
	log.New(stderr, "", 0).Printf("Feed %q failed: %v", feedName, err)
}

func writeDiscoveryTitle(writer io.Writer, item feed.Item) error {
	if writer == nil {
		return nil
	}
	if _, err := fmt.Fprintln(writer, item.Title); err != nil {
		return fmt.Errorf("write discovery log: %w", err)
	}

	return nil
}
