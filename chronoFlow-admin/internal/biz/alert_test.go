package biz

import (
	"context"
	"errors"
	"io"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

type fakeAlertLogRepo struct {
	item *JobLog
}

func (r *fakeAlertLogRepo) GetByID(_ context.Context, id int64) (*JobLog, error) {
	if r.item == nil || r.item.ID != id {
		return nil, nil
	}
	cp := *r.item
	return &cp, nil
}

func (r *fakeAlertLogRepo) MarkAlertPending(context.Context, int64) error {
	r.item.AlertStatus = AlertStatusPending
	r.item.AlertError = ""
	r.item.AlertSentAt = nil
	return nil
}

func (r *fakeAlertLogRepo) MarkAlertSent(_ context.Context, _ int64, sentAt time.Time) error {
	r.item.AlertStatus = AlertStatusSent
	r.item.AlertError = ""
	r.item.AlertSentAt = &sentAt
	return nil
}

func (r *fakeAlertLogRepo) MarkAlertFailed(_ context.Context, _ int64, message string) error {
	r.item.AlertStatus = AlertStatusFailed
	r.item.AlertError = message
	return nil
}

func (r *fakeAlertLogRepo) MarkAlertSkipped(_ context.Context, _ int64, message string) error {
	r.item.AlertStatus = AlertStatusSkipped
	r.item.AlertError = message
	return nil
}

func (r *fakeAlertLogRepo) MarkPendingAlertsFailed(_ context.Context, message string) error {
	if r.item.AlertStatus == AlertStatusPending {
		r.item.AlertStatus = AlertStatusFailed
		r.item.AlertError = message
	}
	return nil
}

type fakeAlertSettings struct {
	webhook    string
	configured bool
}

func (s fakeAlertSettings) GetFeishuWebhook(context.Context) (string, bool, error) {
	return s.webhook, s.configured, nil
}

type flakyAlertSender struct {
	count int32
	fail  int32
}

func (s *flakyAlertSender) SendCard(context.Context, string, any) error {
	count := atomic.AddInt32(&s.count, 1)
	if count <= s.fail {
		return errors.New("temporary feishu error")
	}
	return nil
}

func TestAlertUsecaseRetriesAndMarksSent(t *testing.T) {
	repo := &fakeAlertLogRepo{item: &JobLog{
		ID:                   11,
		JobID:                7,
		JobName:              "alert-job",
		ExecutorName:         "exec",
		Status:               JobLogStatusFailed,
		TriggerType:          TriggerTypeCron,
		StartTime:            time.Now(),
		AlertEnabledSnapshot: true,
		AlertStatus:          AlertStatusNone,
	}}
	sender := &flakyAlertSender{fail: 2}
	uc := newAlertUsecase(repo, fakeAlertSettings{webhook: "http://example.com", configured: true}, sender, 3, 0, log.NewStdLogger(io.Discard))

	if err := uc.sendJobLogAlert(context.Background(), 11); err != nil {
		t.Fatalf("sendJobLogAlert returned error: %v", err)
	}
	if got := atomic.LoadInt32(&sender.count); got != 3 {
		t.Fatalf("expected 3 attempts, got %d", got)
	}
	if repo.item.AlertStatus != AlertStatusSent {
		t.Fatalf("expected alert sent, got %s", repo.item.AlertStatus)
	}
	if repo.item.AlertSentAt == nil {
		t.Fatal("expected alert sent time")
	}
}

func TestAlertUsecaseSkipsWhenWebhookMissing(t *testing.T) {
	repo := &fakeAlertLogRepo{item: &JobLog{
		ID:                   12,
		JobName:              "alert-job",
		Status:               JobLogStatusFailed,
		AlertEnabledSnapshot: true,
		AlertStatus:          AlertStatusNone,
	}}
	uc := newAlertUsecase(repo, fakeAlertSettings{}, &flakyAlertSender{}, 3, 0, log.NewStdLogger(io.Discard))

	if err := uc.sendJobLogAlert(context.Background(), 12); err != nil {
		t.Fatalf("sendJobLogAlert returned error: %v", err)
	}
	if repo.item.AlertStatus != AlertStatusSkipped {
		t.Fatalf("expected alert skipped, got %s", repo.item.AlertStatus)
	}
}

func TestShouldTriggerFailureAlert(t *testing.T) {
	if !ShouldTriggerFailureAlert(JobLogStatusFailed) || !ShouldTriggerFailureAlert(JobLogStatusTimeout) {
		t.Fatal("expected failed and timeout to trigger")
	}
	if ShouldTriggerFailureAlert(JobLogStatusSuccess) {
		t.Fatal("did not expect success to trigger")
	}
}
