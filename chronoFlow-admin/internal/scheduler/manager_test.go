package scheduler

import (
	"context"
	"testing"

	"chronoFlow-admin/internal/conf"
)

func TestManagerRegisterAndRemove(t *testing.T) {
	manager, err := NewManager(&conf.Scheduler{Timezone: "Asia/Shanghai"})
	if err != nil {
		t.Fatalf("NewManager returned error: %v", err)
	}

	if err := manager.Register(1, "0 0 1 * * *", func(context.Context) error { return nil }); err != nil {
		t.Fatalf("Register returned error: %v", err)
	}
	if !manager.Has(1) {
		t.Fatal("expected job 1 registered")
	}
	manager.Remove(1)
	if manager.Has(1) {
		t.Fatal("expected job 1 removed")
	}
}

func TestManagerRejectsInvalidCron(t *testing.T) {
	manager, err := NewManager(&conf.Scheduler{Timezone: "Asia/Shanghai"})
	if err != nil {
		t.Fatalf("NewManager returned error: %v", err)
	}

	err = manager.Register(1, "* * * * *", func(context.Context) error { return nil })
	if err == nil {
		t.Fatal("expected invalid cron error, got nil")
	}
}
