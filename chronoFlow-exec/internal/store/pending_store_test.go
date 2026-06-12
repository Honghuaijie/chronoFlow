package store

import (
	"testing"
	"time"
)

func TestPendingStoreSaveListDelete(t *testing.T) {
	s := NewPendingStore(t.TempDir())
	item := &CallbackItem{
		LogID:       1,
		JobID:       2,
		CallbackURL: "http://admin/internal/job-runs/callback",
		Status:      "success",
		CreatedAt:   time.Now(),
	}

	if err := s.Save(item); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}
	items, err := s.ListPending()
	if err != nil {
		t.Fatalf("ListPending returned error: %v", err)
	}
	if len(items) != 1 || items[0].LogID != 1 {
		t.Fatalf("unexpected items: %+v", items)
	}
	if err := s.DeletePending(1); err != nil {
		t.Fatalf("DeletePending returned error: %v", err)
	}
	items, err = s.ListPending()
	if err != nil {
		t.Fatalf("ListPending returned error: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected empty pending list, got %+v", items)
	}
}
