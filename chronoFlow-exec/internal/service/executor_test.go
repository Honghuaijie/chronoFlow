package service

import (
	"context"
	"testing"
	"time"

	v1 "chronoFlow-exec/api/all-pb-go/v1"
	"chronoFlow-exec/internal/conf"
	"chronoFlow-exec/internal/process"
	"chronoFlow-exec/internal/store"
)

func TestExecutorServiceRunRejectsMissingFields(t *testing.T) {
	svc := NewExecutorService(&conf.Executor{Name: "exec", Token: "token"}, nil, nil, nil)

	_, err := svc.Run(context.Background(), &v1.RunRequest{JobId: 1})
	if err == nil {
		t.Fatal("expected missing fields error, got nil")
	}
}

func TestExecutorServiceRunSurvivesRequestContextCancel(t *testing.T) {
	manager := process.NewManager(process.Config{
		ShellPath:        "/bin/bash",
		TempDir:          t.TempDir(),
		MaxLogBytes:      1024,
		KillGraceSeconds: 1,
	})
	pendingStore := store.NewPendingStore(t.TempDir())
	svc := NewExecutorService(&conf.Executor{Name: "exec", Token: "token"}, manager, pendingStore, nil)
	ctx, cancel := context.WithCancel(context.Background())

	reply, err := svc.Run(ctx, &v1.RunRequest{
		JobId:          1,
		LogId:          10,
		Script:         "echo survives-cancel",
		TimeoutSeconds: 5,
		CallbackUrl:    "http://admin/internal/job-runs/callback",
		CallbackToken:  "callback",
	})
	cancel()
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if reply.GetData().GetStatus() != "accepted" {
		t.Fatalf("expected accepted, got %+v", reply)
	}
	time.Sleep(200 * time.Millisecond)
	if manager.IsRunning(1) {
		t.Fatal("expected job to finish instead of being killed by request context")
	}
	items, err := pendingStore.ListPending()
	if err != nil {
		t.Fatalf("ListPending returned error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 pending callback, got %+v", items)
	}
	if items[0].Status != process.StatusSuccess {
		t.Fatalf("expected success callback, got %+v", items[0])
	}
}
