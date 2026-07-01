package biz

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestBuildFeishuAlertCardIncludesFailureDetails(t *testing.T) {
	exitCode := int32(1)
	endTime := time.Date(2026, 6, 30, 10, 1, 0, 0, time.Local)
	card := buildFeishuAlertCard(FeishuAlertCardInput{
		JobLogID:     123,
		JobID:        9,
		JobName:      "daily-report",
		ExecutorName: "exec-1",
		Status:       JobLogStatusFailed,
		TriggerType:  TriggerTypeCron,
		StartTime:    time.Date(2026, 6, 30, 10, 0, 0, 0, time.Local),
		EndTime:      &endTime,
		DurationMS:   1000,
		ExitCode:     &exitCode,
		ErrorMessage: strings.Repeat("x", 510),
	})

	raw, err := json.Marshal(card)
	if err != nil {
		t.Fatalf("marshal card: %v", err)
	}
	content := string(raw)
	for _, want := range []string{
		"ChronoFlow 任务执行失败",
		"daily-report",
		"123",
		"exec-1",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("expected card to contain %q, got %s", want, content)
		}
	}
	if !strings.Contains(content, strings.Repeat("x", maxAlertErrorLength)+"...") {
		t.Fatalf("expected truncated error content, got %s", content)
	}
}

func TestBuildFailureAlertTitleUsesTimeoutTitle(t *testing.T) {
	if got := buildFailureAlertTitle(JobLogStatusTimeout); got != "ChronoFlow 任务执行超时" {
		t.Fatalf("unexpected timeout title: %s", got)
	}
}
