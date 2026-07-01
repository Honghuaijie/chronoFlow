package biz

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

type fakeAlertDispatcher struct {
	dispatched []int64
}

func (d *fakeAlertDispatcher) DispatchJobLogAlert(_ context.Context, logID int64) {
	d.dispatched = append(d.dispatched, logID)
}

func (d *fakeAlertDispatcher) MarkPendingAlertsFailedOnStartup(context.Context) error {
	return nil
}

func TestCallbackUsecaseWritesLogAndFinalizesStatus(t *testing.T) {
	now := time.Now()
	repo := &fakeRunJobLogRepo{created: &JobLog{ID: 1, JobID: 10, Status: JobLogStatusRunning, StartTime: now}}
	store := &fakeCallbackLogStore{}
	uc := NewCallbackUsecase(repo, store, CallbackConfig{MaxLogBytes: 1024}, nil, log.DefaultLogger)

	got, err := uc.ApplyCallback(context.Background(), &CallbackInput{
		LogID:        1,
		JobID:        10,
		Status:       JobLogStatusSuccess,
		ExitCode:     0,
		LogContent:   "hello",
		LogTruncated: false,
		EndTime:      now.Add(time.Second),
		DurationMS:   1000,
	})
	if err != nil {
		t.Fatalf("ApplyCallback returned error: %v", err)
	}
	if got.Status != JobLogStatusSuccess {
		t.Fatalf("expected success, got %q", got.Status)
	}
	if store.written != "hello" {
		t.Fatalf("expected log content written, got %q", store.written)
	}
	if repo.updated.LogPath != "logs/1.log" || repo.updated.LogSizeBytes != 5 {
		t.Fatalf("expected log metadata updated, got %+v", repo.updated)
	}
}

func TestCallbackUsecaseDispatchesAlertWhenFailed(t *testing.T) {
	now := time.Now()
	repo := &fakeRunJobLogRepo{created: &JobLog{ID: 1, JobID: 10, Status: JobLogStatusRunning, StartTime: now}}
	store := &fakeCallbackLogStore{}
	alerts := &fakeAlertDispatcher{}
	uc := NewCallbackUsecase(repo, store, CallbackConfig{MaxLogBytes: 1024}, alerts, log.DefaultLogger)

	_, err := uc.ApplyCallback(context.Background(), &CallbackInput{
		LogID:      1,
		JobID:      10,
		Status:     JobLogStatusFailed,
		ExitCode:   1,
		LogContent: "failed",
		EndTime:    now.Add(time.Second),
	})
	if err != nil {
		t.Fatalf("ApplyCallback returned error: %v", err)
	}
	if !reflect.DeepEqual(alerts.dispatched, []int64{1}) {
		t.Fatalf("expected alert dispatched for log 1, got %+v", alerts.dispatched)
	}
}

func TestCallbackUsecaseDoesNotDispatchAlertWhenSuccess(t *testing.T) {
	now := time.Now()
	repo := &fakeRunJobLogRepo{created: &JobLog{ID: 1, JobID: 10, Status: JobLogStatusRunning, StartTime: now}}
	store := &fakeCallbackLogStore{}
	alerts := &fakeAlertDispatcher{}
	uc := NewCallbackUsecase(repo, store, CallbackConfig{MaxLogBytes: 1024}, alerts, log.DefaultLogger)

	_, err := uc.ApplyCallback(context.Background(), &CallbackInput{
		LogID:      1,
		JobID:      10,
		Status:     JobLogStatusSuccess,
		LogContent: "ok",
		EndTime:    now.Add(time.Second),
	})
	if err != nil {
		t.Fatalf("ApplyCallback returned error: %v", err)
	}
	if len(alerts.dispatched) != 0 {
		t.Fatalf("expected no alert dispatched, got %+v", alerts.dispatched)
	}
}

func TestCallbackUsecaseIgnoresFinalLog(t *testing.T) {
	now := time.Now()
	repo := &fakeRunJobLogRepo{created: &JobLog{ID: 1, JobID: 10, Status: JobLogStatusSuccess, StartTime: now}}
	store := &fakeCallbackLogStore{}
	uc := NewCallbackUsecase(repo, store, CallbackConfig{MaxLogBytes: 1024}, nil, log.DefaultLogger)

	got, err := uc.ApplyCallback(context.Background(), &CallbackInput{
		LogID:      1,
		JobID:      10,
		Status:     JobLogStatusFailed,
		LogContent: "late",
		EndTime:    now.Add(time.Second),
	})
	if err != nil {
		t.Fatalf("ApplyCallback returned error: %v", err)
	}
	if got.Status != JobLogStatusSuccess {
		t.Fatalf("expected original final status, got %q", got.Status)
	}
	if store.written != "" || repo.updated != nil {
		t.Fatalf("expected no write/update, written=%q updated=%+v", store.written, repo.updated)
	}
}

func TestCallbackUsecaseRejectsMismatchedJobID(t *testing.T) {
	now := time.Now()
	repo := &fakeRunJobLogRepo{created: &JobLog{ID: 1, JobID: 10, Status: JobLogStatusRunning, StartTime: now}}
	store := &fakeCallbackLogStore{}
	uc := NewCallbackUsecase(repo, store, CallbackConfig{MaxLogBytes: 1024}, nil, log.DefaultLogger)

	_, err := uc.ApplyCallback(context.Background(), &CallbackInput{
		LogID:      1,
		JobID:      99,
		Status:     JobLogStatusSuccess,
		LogContent: "wrong job",
		EndTime:    now.Add(time.Second),
	})
	if err == nil {
		t.Fatal("expected mismatched job_id error, got nil")
	}
	if store.written != "" || repo.updated != nil {
		t.Fatalf("expected no write/update, written=%q updated=%+v", store.written, repo.updated)
	}
}
