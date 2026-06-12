package callback

import (
	"testing"
	"time"

	"chronoFlow-exec/internal/conf"
	"chronoFlow-exec/internal/store"

	"github.com/go-kratos/kratos/v2/log"
)

func TestWorkerDeletesExpiredPendingCallbacks(t *testing.T) {
	pendingStore := store.NewPendingStore(t.TempDir())
	expired := &store.CallbackItem{
		LogID:     1,
		JobID:     2,
		Status:    "success",
		CreatedAt: time.Now().Add(-48 * time.Hour),
	}
	if err := pendingStore.Save(expired); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}
	worker := NewWorker(nil, pendingStore, &conf.Callback{PendingRetentionDays: 1}, log.DefaultLogger)

	worker.retryOnce()

	items, err := pendingStore.ListPending()
	if err != nil {
		t.Fatalf("ListPending returned error: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected expired pending callback deleted, got %+v", items)
	}
}
