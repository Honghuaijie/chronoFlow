package process

import (
	"context"
	"testing"
	"time"
)

func TestManagerRejectsSameJobRunning(t *testing.T) {
	manager := NewManager(Config{ShellPath: "/bin/bash", TempDir: t.TempDir(), MaxLogBytes: 1024, KillGraceSeconds: 1})
	manager.running[1] = &RunState{JobID: 1, LogID: 10}

	err := manager.Run(context.Background(), RunRequest{JobID: 1, LogID: 11, Script: "echo hi"}, func(*Result) {})
	if err == nil {
		t.Fatal("expected running conflict, got nil")
	}
}

func TestManagerRunCompletes(t *testing.T) {
	manager := NewManager(Config{ShellPath: "/bin/bash", TempDir: t.TempDir(), MaxLogBytes: 1024, KillGraceSeconds: 1})
	done := make(chan *Result, 1)

	err := manager.Run(context.Background(), RunRequest{JobID: 1, LogID: 10, Script: "echo hi", TimeoutSeconds: 5}, func(result *Result) {
		done <- result
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	select {
	case result := <-done:
		if result.Status != StatusSuccess || result.ExitCode != 0 || result.LogContent == "" {
			t.Fatalf("unexpected result: %+v", result)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for result")
	}
}
