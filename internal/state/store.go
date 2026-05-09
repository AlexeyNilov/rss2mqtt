package state

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	DefaultPath            = ".rss2mqtt-state.json"
	DefaultMaxItemsPerFeed = 256
	stateVersion           = 1
)

type Store struct {
	path            string
	maxItemsPerFeed int
	feeds           map[string][]string
}

type stateFile struct {
	Version int                 `json:"version"`
	Feeds   map[string][]string `json:"feeds"`
}

func Load(path string) (*Store, error) {
	return load(path, DefaultMaxItemsPerFeed)
}

func load(path string, maxItemsPerFeed int) (*Store, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return newStore(path, maxItemsPerFeed), nil
	}
	if err != nil {
		return nil, fmt.Errorf("read state: %w", err)
	}

	var file stateFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("parse state: %w", err)
	}
	if file.Feeds == nil {
		file.Feeds = make(map[string][]string)
	}

	return &Store{path: path, maxItemsPerFeed: maxItemsPerFeed, feeds: file.Feeds}, nil
}

func (s *Store) Seen(feedName, identity string) bool {
	hash, ok := hashIdentity(identity)
	if !ok {
		return false
	}

	return containsHash(s.feeds[feedName], hash)
}

func (s *Store) Mark(feedName, identity string) {
	hash, ok := hashIdentity(identity)
	if !ok {
		return
	}

	feedHashes := removeHash(s.feeds[feedName], hash)
	feedHashes = append([]string{hash}, feedHashes...)
	s.feeds[feedName] = limit(feedHashes, s.maxItemsPerFeed)
}

func (s *Store) Save() error {
	file := stateFile{Version: stateVersion, Feeds: s.feeds}
	data, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return fmt.Errorf("encode state: %w", err)
	}

	return writeFileAtomic(s.path, append(data, '\n'))
}

func newStore(path string, maxItemsPerFeed int) *Store {
	return &Store{
		path:            path,
		maxItemsPerFeed: maxItemsPerFeed,
		feeds:           make(map[string][]string),
	}
}

func hashIdentity(identity string) (string, bool) {
	trimmed := strings.TrimSpace(identity)
	if trimmed == "" {
		return "", false
	}

	sum := sha256.Sum256([]byte(trimmed))
	return hex.EncodeToString(sum[:]), true
}

func containsHash(hashes []string, hash string) bool {
	for _, candidate := range hashes {
		if candidate == hash {
			return true
		}
	}

	return false
}

func removeHash(hashes []string, hash string) []string {
	filtered := hashes[:0]
	for _, candidate := range hashes {
		if candidate != hash {
			filtered = append(filtered, candidate)
		}
	}

	return filtered
}

func limit(hashes []string, max int) []string {
	if max <= 0 || len(hashes) <= max {
		return hashes
	}

	return hashes[:max]
}

func writeFileAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create state directory: %w", err)
	}

	tmp, err := os.CreateTemp(dir, filepath.Base(path)+".tmp-")
	if err != nil {
		return fmt.Errorf("create temporary state file: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if err := writeAndClose(tmp, data); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		if removeErr := os.Remove(path); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
			return fmt.Errorf("replace state file: %w", removeErr)
		}
		if retryErr := os.Rename(tmpPath, path); retryErr != nil {
			return fmt.Errorf("replace state file: %w", retryErr)
		}
	}

	return nil
}

func writeAndClose(file *os.File, data []byte) error {
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		return fmt.Errorf("write temporary state file: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close temporary state file: %w", err)
	}

	return nil
}
