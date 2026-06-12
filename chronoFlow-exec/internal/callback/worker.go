package callback

import (
	"context"
	"time"

	"chronoFlow-exec/internal/conf"
	"chronoFlow-exec/internal/store"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

var WorkerProviderSet = wire.NewSet(NewWorker)

type Worker struct {
	client *Client
	store  *store.PendingStore
	conf   *conf.Callback
	log    *log.Helper
	stop   chan struct{}
}

func NewWorker(client *Client, store *store.PendingStore, c *conf.Callback, logger log.Logger) *Worker {
	return &Worker{
		client: client,
		store:  store,
		conf:   c,
		log:    log.NewHelper(logger),
		stop:   make(chan struct{}),
	}
}

func (w *Worker) Start(ctx context.Context) error {
	w.retryOnce()
	w.loop()
	return nil
}

func (w *Worker) Stop(ctx context.Context) error {
	select {
	case <-w.stop:
	default:
		close(w.stop)
	}
	return nil
}

func (w *Worker) loop() {
	ticker := time.NewTicker(time.Duration(w.retryIntervalSeconds()) * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			w.retryOnce()
		case <-w.stop:
			return
		}
	}
}

func (w *Worker) retryOnce() {
	items, err := w.store.ListPending()
	if err != nil {
		w.log.Errorf("list pending callbacks failed: %v", err)
		return
	}
	for _, item := range items {
		if w.isExpired(item.CreatedAt) {
			continue
		}
		if err := w.client.Send(item); err != nil {
			w.log.Errorf("callback retry failed: log_id=%d err=%v", item.LogID, err)
			continue
		}
		_ = w.store.DeletePending(item.LogID)
	}
}

func (w *Worker) retryIntervalSeconds() int32 {
	if w.conf != nil && w.conf.RetryIntervalSeconds > 0 {
		return w.conf.RetryIntervalSeconds
	}
	return 30
}

func (w *Worker) isExpired(createdAt time.Time) bool {
	days := int32(7)
	if w.conf != nil && w.conf.PendingRetentionDays > 0 {
		days = w.conf.PendingRetentionDays
	}
	return !createdAt.IsZero() && time.Since(createdAt) > time.Duration(days)*24*time.Hour
}
