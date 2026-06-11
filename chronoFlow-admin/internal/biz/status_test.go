package biz

import "testing"

func TestJobLogStatusFinality(t *testing.T) {
	finalStatuses := []string{
		JobLogStatusSuccess,
		JobLogStatusFailed,
		JobLogStatusTimeout,
		JobLogStatusSkipped,
		JobLogStatusKilled,
	}
	for _, status := range finalStatuses {
		if !IsFinalJobLogStatus(status) {
			t.Fatalf("IsFinalJobLogStatus(%q) = false, want true", status)
		}
	}

	runningStatuses := []string{
		JobLogStatusRunning,
		JobLogStatusKilling,
	}
	for _, status := range runningStatuses {
		if IsFinalJobLogStatus(status) {
			t.Fatalf("IsFinalJobLogStatus(%q) = true, want false", status)
		}
		if !CanCallbackUpdateJobLogStatus(status) {
			t.Fatalf("CanCallbackUpdateJobLogStatus(%q) = false, want true", status)
		}
	}
}

func TestCanCallbackUpdateJobLogStatusRejectsFinalStatus(t *testing.T) {
	if CanCallbackUpdateJobLogStatus(JobLogStatusSuccess) {
		t.Fatal("CanCallbackUpdateJobLogStatus(success) = true, want false")
	}
}
