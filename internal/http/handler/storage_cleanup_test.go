package handler

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"fleet/internal/platform/storage"
)

func saveObject(t *testing.T, s storage.Storage, key string) string {
	t.Helper()
	if _, err := s.Save(context.Background(), key, strings.NewReader("payload"), "text/plain"); err != nil {
		t.Fatalf("save: %v", err)
	}
	return s.URL(key)
}

func TestReclaimObjectsDeletesStoredObjects(t *testing.T) {
	s := storage.NewLocal(t.TempDir(), "/media")
	url := saveObject(t, s, "uploads/1/old.txt")

	reclaimObjects(context.Background(), s, slog.Default(), url)

	if _, err := s.Stat(context.Background(), "uploads/1/old.txt"); !errors.Is(err, storage.ErrNotFound) {
		t.Errorf("object still present: %v", err)
	}
}

func TestReclaimObjectsIgnoresWhatItDoesNotOwn(t *testing.T) {
	s := storage.NewLocal(t.TempDir(), "/media")
	kept := saveObject(t, s, "uploads/1/keep.txt")

	// Empty values, foreign hosts and an already-deleted key must all be
	// no-ops: reclamation runs after a successful write and must never fail it.
	reclaimObjects(context.Background(), s, slog.Default(),
		"", "https://example.com/somebody-elses.png", "/media/uploads/1/never-existed.txt")

	if _, err := s.Stat(context.Background(), "uploads/1/keep.txt"); err != nil {
		t.Errorf("unrelated object was disturbed: %v", err)
	}
	if kept == "" {
		t.Error("expected a URL for the kept object")
	}
}

func TestReclaimObjectsToleratesNoStorage(t *testing.T) {
	// A store wired without storage must not panic on a delete path.
	reclaimObjects(context.Background(), nil, slog.Default(), "/media/uploads/1/x.txt")
}

func TestChangedURL(t *testing.T) {
	old, current, other := "old.png", "old.png", "new.png"

	tests := []struct {
		name      string
		before    *string
		after     *string
		wantURL   string
		wantMatch bool
	}{
		{"replaced", &old, &other, "old.png", true},
		{"cleared", &old, nil, "old.png", true},
		{"unchanged", &old, &current, "", false},
		{"was empty", nil, &other, "", false},
		{"both empty", nil, nil, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, ok := changedURL(tt.before, tt.after)
			if ok != tt.wantMatch || url != tt.wantURL {
				t.Errorf("changedURL = (%q, %v), want (%q, %v)", url, ok, tt.wantURL, tt.wantMatch)
			}
		})
	}
}
