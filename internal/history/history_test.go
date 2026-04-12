package history

import (
	"path/filepath"
	"testing"
	"time"
)

func TestTouchAndSortRepos(t *testing.T) {
	h := &History{}

	h.Touch("acme/first")
	time.Sleep(time.Millisecond)
	h.Touch("acme/second")

	repos := []string{"acme/third", "acme/first", "acme/second"}
	sorted := h.SortRepos(repos)

	// Most recent first
	if sorted[0] != "acme/second" {
		t.Errorf("sorted[0] = %q, want acme/second", sorted[0])
	}
	if sorted[1] != "acme/first" {
		t.Errorf("sorted[1] = %q, want acme/first", sorted[1])
	}
	// Unknown repos at end
	if sorted[2] != "acme/third" {
		t.Errorf("sorted[2] = %q, want acme/third", sorted[2])
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history.json")

	h := &History{}
	h.Touch("acme/demo")
	if err := h.SaveTo(path); err != nil {
		t.Fatal(err)
	}

	loaded := LoadFrom(path)
	if len(loaded.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(loaded.Entries))
	}
	if loaded.Entries[0].Repository != "acme/demo" {
		t.Errorf("repo = %q, want acme/demo", loaded.Entries[0].Repository)
	}
}

func TestTouchUpdatesExisting(t *testing.T) {
	h := &History{}
	h.Touch("acme/demo")
	first := h.Entries[0].LastUsed

	time.Sleep(time.Millisecond)
	h.Touch("acme/demo")

	if len(h.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(h.Entries))
	}
	if !h.Entries[0].LastUsed.After(first) {
		t.Error("lastUsed should have been updated")
	}
}
