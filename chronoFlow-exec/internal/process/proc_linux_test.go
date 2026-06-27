//go:build linux

package process

import (
	"context"
	"os/exec"
	"syscall"
	"testing"
	"time"
)

func TestWaitProcessGroupGoneUsesProcWithoutPs(t *testing.T) {
	cmd := exec.Command("/bin/bash", "-c", "sleep 30 & wait")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		t.Fatalf("start command: %v", err)
	}
	pgid := cmd.Process.Pid
	defer func() {
		_ = syscall.Kill(-pgid, syscall.SIGKILL)
		_, _ = cmd.Process.Wait()
	}()

	if !ProcessGroupExists(pgid) {
		t.Fatalf("expected process group %d to exist", pgid)
	}

	if err := syscall.Kill(-pgid, syscall.SIGKILL); err != nil {
		t.Fatalf("kill process group: %v", err)
	}
	waitDone := make(chan error, 1)
	go func() {
		_, err := cmd.Process.Wait()
		waitDone <- err
	}()
	select {
	case <-waitDone:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for shell process to exit")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := WaitProcessGroupGone(ctx, pgid, 50*time.Millisecond); err != nil {
		t.Fatalf("WaitProcessGroupGone returned error: %v", err)
	}
}
