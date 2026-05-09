package state

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStoreAllowsNewItem(t *testing.T) {
	store := newTestStore(t, 256)

	if store.Seen("oreilly-radar", "https://example.com/item-1") {
		t.Fatal("Seen() = true, want false for new item")
	}
}

func TestStoreSuppressesItemSeenInEarlierRun(t *testing.T) {
	path := statePath(t)

	firstRun, err := load(path, 256)
	if err != nil {
		t.Fatalf("load first run: %v", err)
	}
	firstRun.Mark("oreilly-radar", "https://example.com/item-1")
	if err := firstRun.Save(); err != nil {
		t.Fatalf("save first run: %v", err)
	}

	secondRun, err := load(path, 256)
	if err != nil {
		t.Fatalf("load second run: %v", err)
	}
	if !secondRun.Seen("oreilly-radar", "https://example.com/item-1") {
		t.Fatal("Seen() = false, want true for item saved in earlier run")
	}
}

func TestStoreKeepsFeedsSeparate(t *testing.T) {
	store := newTestStore(t, 256)
	store.Mark("oreilly-radar", "same-identity")

	if store.Seen("go-blog", "same-identity") {
		t.Fatal("Seen() = true, want false for same identity in different feed")
	}
}

func TestLoadMissingStateReturnsEmptyStore(t *testing.T) {
	store, err := Load(statePath(t))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if store.Seen("oreilly-radar", "https://example.com/item-1") {
		t.Fatal("Seen() = true, want false for empty missing state")
	}
}

func TestLoadCorruptStateReturnsError(t *testing.T) {
	path := statePath(t)
	if err := os.WriteFile(path, []byte("{not-json"), 0o600); err != nil {
		t.Fatalf("write corrupt state: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("Load() error = nil, want corrupt state error")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "state") {
		t.Fatalf("Load() error = %q, want state context", err)
	}
}

func TestStoreRetainsBoundedItemsPerFeed(t *testing.T) {
	path := statePath(t)
	store, err := load(path, 2)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	store.Mark("oreilly-radar", "item-1")
	store.Mark("oreilly-radar", "item-2")
	store.Mark("oreilly-radar", "item-3")

	if store.Seen("oreilly-radar", "item-1") {
		t.Fatal("Seen() = true, want oldest item evicted")
	}
	if !store.Seen("oreilly-radar", "item-2") {
		t.Fatal("Seen() = false, want retained item-2")
	}
	if !store.Seen("oreilly-radar", "item-3") {
		t.Fatal("Seen() = false, want retained item-3")
	}
}

func TestSaveWritesStateFileWithoutRawIdentity(t *testing.T) {
	path := statePath(t)
	store, err := load(path, 256)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	store.Mark("oreilly-radar", "https://example.com/private-item")

	if err := store.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read state file: %v", err)
	}
	if !strings.Contains(string(data), "oreilly-radar") {
		t.Fatalf("state file = %q, want feed name", data)
	}
	if strings.Contains(string(data), "https://example.com/private-item") {
		t.Fatalf("state file contains raw identity: %q", data)
	}
}

func newTestStore(t *testing.T, maxItems int) *Store {
	t.Helper()

	store, err := load(statePath(t), maxItems)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	return store
}

func statePath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), DefaultPath)
}
