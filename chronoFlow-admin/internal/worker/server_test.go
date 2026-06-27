package worker

import (
	"context"
	"errors"
	"testing"

	"chronoFlow-admin/internal/biz"
	"chronoFlow-admin/internal/conf"
	"chronoFlow-admin/internal/scheduler"

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

func (r *fakeExecutorRepo) GetByAddress(context.Context, string) (*biz.Executor, error) {
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

type fakeJobRepo struct {
	items []*biz.Job
}

func (r *fakeJobRepo) Create(context.Context, *biz.Job) (*biz.Job, error) { return nil, nil }
func (r *fakeJobRepo) GetByID(context.Context, int64) (*biz.Job, error)   { return nil, nil }
func (r *fakeJobRepo) List(context.Context, int64) ([]*biz.Job, error)    { return r.items, nil }
func (r *fakeJobRepo) Update(context.Context, *biz.Job) (*biz.Job, error) { return nil, nil }
func (r *fakeJobRepo) Delete(context.Context, int64) error                { return nil }

type fakeJobRunner struct{}

func (fakeJobRunner) RunJob(context.Context, int64, string) (*biz.JobRunResult, error) {
	return &biz.JobRunResult{}, nil
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
		nil,
		nil,
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

func TestWorkerRegistersRunningJobsOnStartup(t *testing.T) {
	manager, err := scheduler.NewManager(&conf.Scheduler{Timezone: "Asia/Shanghai"})
	if err != nil {
		t.Fatalf("new scheduler manager: %v", err)
	}
	server := NewServer(
		&conf.Executor{},
		&conf.Recovery{},
		&conf.Logs{},
		&fakeJobRepo{items: []*biz.Job{
			{ID: 10, CronExpr: "0 */5 * * * *", ScheduleStatus: biz.ScheduleStatusRunning},
			{ID: 11, CronExpr: "0 */5 * * * *", ScheduleStatus: biz.ScheduleStatusStopped},
		}},
		fakeJobRunner{},
		&fakeExecutorRepo{},
		&fakeMaintenanceRepo{},
		fakeTokenCipher{},
		fakeHealthClient{},
		nil,
		manager,
		log.DefaultLogger,
	)

	if err := server.registerRunningJobs(context.Background()); err != nil {
		t.Fatalf("registerRunningJobs returned error: %v", err)
	}
	if !manager.Has(10) {
		t.Fatalf("expected running job 10 to be registered")
	}
	if manager.Has(11) {
		t.Fatalf("expected stopped job 11 to remain unregistered")
	}
}
