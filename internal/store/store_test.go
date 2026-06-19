package store

import (
	"path/filepath"
	"sync"
	"testing"
)

func openTestStore(t *testing.T) *Store {
	t.Helper()

	store, err := Open(filepath.Join(t.TempDir(), "readmarker.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
	})
	return store
}

func TestUnknownSourceReadsAsNotYetRead(t *testing.T) {
	store := openTestStore(t)

	got, ok, err := store.Get("slack:workspace:channel")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if ok {
		t.Fatal("Get() ok = true, want false")
	}
	if got != 0 {
		t.Fatalf("Get() = %d, want 0", got)
	}
}

func TestAdvanceNeverRewinds(t *testing.T) {
	store := openTestStore(t)

	got, err := store.Advance("github:owner/repo#1", 100)
	if err != nil {
		t.Fatalf("Advance() error = %v", err)
	}
	if got != 100 {
		t.Fatalf("Advance() = %d, want 100", got)
	}

	got, err = store.Advance("github:owner/repo#1", 99)
	if err != nil {
		t.Fatalf("Advance() error = %v", err)
	}
	if got != 100 {
		t.Fatalf("Advance() = %d, want 100", got)
	}
}

func TestSetOverridesCursor(t *testing.T) {
	store := openTestStore(t)

	if _, err := store.Advance("chatwork:room", 100); err != nil {
		t.Fatalf("Advance() error = %v", err)
	}
	if err := store.Set("chatwork:room", 3); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	got, ok, err := store.Get("chatwork:room")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if !ok {
		t.Fatal("Get() ok = false, want true")
	}
	if got != 3 {
		t.Fatalf("Get() = %d, want 3", got)
	}
}

func TestListReturnsCursorsInKeyOrder(t *testing.T) {
	store := openTestStore(t)

	if err := store.Set("slack:b", 2); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	if err := store.Set("slack:a", 1); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	got, err := store.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	want := []Cursor{
		{SourceKey: "slack:a", Cursor: 1},
		{SourceKey: "slack:b", Cursor: 2},
	}
	if len(got) != len(want) {
		t.Fatalf("List() len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("List()[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestConcurrentAdvanceStaysAtomic(t *testing.T) {
	store := openTestStore(t)
	const workers = 128

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(pos uint64) {
			defer wg.Done()
			if _, err := store.Advance("slack:workspace:channel", pos); err != nil {
				t.Errorf("Advance(%d) error = %v", pos, err)
			}
		}(uint64(i))
	}
	wg.Wait()

	got, ok, err := store.Get("slack:workspace:channel")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if !ok {
		t.Fatal("Get() ok = false, want true")
	}
	if got != workers-1 {
		t.Fatalf("Get() = %d, want %d", got, workers-1)
	}
}
