package worker

import (
	"context"
	"errors"
	"testing"

	"chronoFlow-admin/internal/biz"
	"chronoFlow-admin/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
)

type fakeExecutorRepo struct {
	items   []*biz.Executor
	updated *biz.Executor
}

func (r *fakeExecutorRepo) List(context.Context) ([]*biz.Executor, error) {
	return r.items, nil
}

func (r *fakeExecutorRepo) Create(context.Context, *biz.Executor) (*biz.Executor, error) {
	return nil, nil
}

func (r *fakeExecutorRepo) GetByID(context.Context, int64) (*biz.Executor, error) {
	return nil, nil
}

func (r *fakeExecutorRepo) Update(_ context.Context, executor *biz.Executor) (*biz.Executor, error) {
	cp := *executor
	r.updated = &cp
	return &cp, nil
}

func (r *fakeExecutorRepo) Delete(context.Context, int64) error {
	return nil
}

type fakeTokenCipher struct{}

func (fakeTokenCipher) Encrypt(string) (string, error) { return "", nil }
func (fakeTokenCipher) Decrypt(string) (string, error) { return "token", nil }

type fakeHealthClient struct{}

func (fakeHealthClient) Health(context.Context, string, string) error {
	return errors.New("offline")
}

type fakeMaintenanceRepo struct {
	failedExecutorID int64
}

func (r *fakeMaintenanceRepo) MarkActiveLogsFailedByExecutorID(_ context.Context, executorID int64, _ string) error {
	r.failedExecutorID = executorID
	return nil
}

func (r *fakeMaintenanceRepo) MarkAllActiveLogsFailed(context.Context, string) error { return nil }
func (r *fakeMaintenanceRepo) MarkKillingTimeoutLogsFailed(context.Context, int32, string) error {
	return nil
}
func (r *fakeMaintenanceRepo) DeleteExpiredLogs(context.Context, int32) ([]string, error) {
	return nil, nil
}

func TestWorkerCheckExecutorsMarksOfflineAfterThreshold(t *testing.T) {
	executorRepo := &fakeExecutorRepo{items: []*biz.Executor{{
		ID:                 7,
		Address:            "http://exec",
		TokenCiphertext:    "cipher",
		Status:             biz.ExecutorStatusOnline,
		HeartbeatFailCount: 2,
	}}}
	maintenanceRepo := &fakeMaintenanceRepo{}
	server := NewServer(
		&conf.Executor{HealthCheckFailThreshold: 3},
		&conf.Recovery{},
		&conf.Logs{},
		executorRepo,
		maintenanceRepo,
		fakeTokenCipher{},
		fakeHealthClient{},
		nil,
		nil,
		log.DefaultLogger,
	)

	if err := server.checkExecutorsOnce(context.Background()); err != nil {
		t.Fatalf("checkExecutorsOnce returned error: %v", err)
	}
	if executorRepo.updated.Status != biz.ExecutorStatusOffline {
		t.Fatalf("expected offline, got %+v", executorRepo.updated)
	}
	if maintenanceRepo.failedExecutorID != 7 {
		t.Fatalf("expected active logs failed for executor 7, got %d", maintenanceRepo.failedExecutorID)
	}
}
