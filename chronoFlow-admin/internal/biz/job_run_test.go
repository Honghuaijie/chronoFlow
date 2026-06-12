package biz

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

type fakeRunJobRepo struct {
	job *Job
}

func (r fakeRunJobRepo) GetByID(context.Context, int64) (*Job, error) {
	cp := *r.job
	return &cp, nil
}

type fakeRunGlueRepo struct {
	glue *Glue
}

func (r fakeRunGlueRepo) GetByJobID(context.Context, int64) (*Glue, error) {
	cp := *r.glue
	return &cp, nil
}

func (r fakeRunGlueRepo) Save(context.Context, *Glue) (*Glue, error) {
	return nil, nil
}

type fakeRunExecutorRepo struct {
	executor *Executor
}

func (r fakeRunExecutorRepo) GetByID(context.Context, int64) (*Executor, error) {
	cp := *r.executor
	return &cp, nil
}

type fakeRunJobLogRepo struct {
	running *JobLog
	created *JobLog
	updated *JobLog
}

func (r *fakeRunJobLogRepo) GetRunningByJobID(context.Context, int64) (*JobLog, error) {
	if r.running == nil {
		return nil, nil
	}
	cp := *r.running
	return &cp, nil
}

func (r *fakeRunJobLogRepo) Create(_ context.Context, jobLog *JobLog) (*JobLog, error) {
	cp := *jobLog
	if cp.ID == 0 {
		cp.ID = 1
	}
	r.created = &cp
	return &cp, nil
}

func (r *fakeRunJobLogRepo) CreateRunningIfNoActive(ctx context.Context, jobLog *JobLog) (*JobLog, error) {
	return r.Create(ctx, jobLog)
}

func (r *fakeRunJobLogRepo) GetByID(context.Context, int64) (*JobLog, error) {
	if r.created != nil {
		cp := *r.created
		return &cp, nil
	}
	if r.running != nil {
		cp := *r.running
		return &cp, nil
	}
	return nil, nil
}

func (r *fakeRunJobLogRepo) Update(_ context.Context, jobLog *JobLog) (*JobLog, error) {
	cp := *jobLog
	r.updated = &cp
	return &cp, nil
}

type fakeExecutorRunner struct {
	runReq  *ExecutorRunRequest
	killReq *ExecutorKillRequest
	runErr  error
	killErr error
}

func (r *fakeExecutorRunner) Run(_ context.Context, _ string, _ string, req ExecutorRunRequest) error {
	cp := req
	r.runReq = &cp
	return r.runErr
}

func (r *fakeExecutorRunner) Kill(_ context.Context, _ string, _ string, req ExecutorKillRequest) error {
	cp := req
	r.killReq = &cp
	return r.killErr
}

type fakeCallbackLogStore struct {
	written string
}

func (s *fakeCallbackLogStore) Write(_ context.Context, _ int64, _ int64, content string) (string, int64, error) {
	s.written = content
	return "logs/1.log", int64(len(content)), nil
}

func TestJobRunUsecaseRunCreatesLogAndCallsExecutor(t *testing.T) {
	jobLogRepo := &fakeRunJobLogRepo{}
	runner := &fakeExecutorRunner{}
	uc := NewJobRunUsecase(
		fakeRunJobRepo{job: &Job{ID: 1, ExecutorID: 2, Name: "daily", CronExpr: "0 0 1 * * *", TimeoutSeconds: 30}},
		fakeRunGlueRepo{glue: &Glue{JobID: 1, Content: "echo hello"}},
		fakeRunExecutorRepo{executor: &Executor{ID: 2, Name: "exec", Address: "http://exec", TokenCiphertext: "cipher"}},
		jobLogRepo,
		fakeTokenCipher{},
		runner,
		JobRunConfig{PublicBaseURL: "http://admin", CallbackToken: "callback"},
		log.DefaultLogger,
	)

	got, err := uc.RunJob(context.Background(), 1, TriggerTypeManual)
	if err != nil {
		t.Fatalf("RunJob returned error: %v", err)
	}
	if got.Status != JobLogStatusRunning || got.LogID != 1 {
		t.Fatalf("unexpected result: %+v", got)
	}
	if jobLogRepo.created.Status != JobLogStatusRunning || jobLogRepo.created.GlueSnapshot != "echo hello" {
		t.Fatalf("unexpected created log: %+v", jobLogRepo.created)
	}
	if runner.runReq == nil || runner.runReq.CallbackURL != "http://admin/internal/job-runs/callback" {
		t.Fatalf("executor was not called with callback url: %+v", runner.runReq)
	}
}

func TestJobRunUsecaseRunRejectsWhenSameJobRunning(t *testing.T) {
	uc := NewJobRunUsecase(
		fakeRunJobRepo{job: &Job{ID: 1, ExecutorID: 2, Name: "daily", CronExpr: "0 0 1 * * *", TimeoutSeconds: 30}},
		fakeRunGlueRepo{glue: &Glue{JobID: 1, Content: "echo hello"}},
		fakeRunExecutorRepo{executor: &Executor{ID: 2, Name: "exec", Address: "http://exec", TokenCiphertext: "cipher"}},
		&fakeRunJobLogRepo{running: &JobLog{ID: 9, JobID: 1, Status: JobLogStatusRunning}},
		fakeTokenCipher{},
		&fakeExecutorRunner{},
		JobRunConfig{PublicBaseURL: "http://admin", CallbackToken: "callback"},
		log.DefaultLogger,
	)

	_, err := uc.RunJob(context.Background(), 1, TriggerTypeManual)
	if err == nil {
		t.Fatal("expected running conflict, got nil")
	}
}

func TestJobRunUsecaseRunMarksLogFailedWhenDispatchFails(t *testing.T) {
	jobLogRepo := &fakeRunJobLogRepo{}
	runner := &fakeExecutorRunner{runErr: errors.New("executor unavailable")}
	uc := NewJobRunUsecase(
		fakeRunJobRepo{job: &Job{ID: 1, ExecutorID: 2, Name: "daily", CronExpr: "0 0 1 * * *", TimeoutSeconds: 30}},
		fakeRunGlueRepo{glue: &Glue{JobID: 1, Content: "echo hello"}},
		fakeRunExecutorRepo{executor: &Executor{ID: 2, Name: "exec", Address: "http://exec", TokenCiphertext: "cipher"}},
		jobLogRepo,
		fakeTokenCipher{},
		runner,
		JobRunConfig{PublicBaseURL: "http://admin", CallbackToken: "callback"},
		log.DefaultLogger,
	)

	_, err := uc.RunJob(context.Background(), 1, TriggerTypeManual)
	if err == nil {
		t.Fatal("expected dispatch error, got nil")
	}
	if jobLogRepo.updated == nil || jobLogRepo.updated.Status != JobLogStatusFailed {
		t.Fatalf("expected created log marked failed, got %+v", jobLogRepo.updated)
	}
	if jobLogRepo.updated.EndTime == nil || jobLogRepo.updated.ErrorMessage == "" {
		t.Fatalf("expected failed log end time and error message, got %+v", jobLogRepo.updated)
	}
}

func TestJobRunUsecaseKillMarksKillingAndCallsExecutor(t *testing.T) {
	now := time.Now()
	jobLogRepo := &fakeRunJobLogRepo{running: &JobLog{ID: 9, JobID: 1, ExecutorID: 2, ExecutorAddress: "http://exec", Status: JobLogStatusRunning, StartTime: now}}
	runner := &fakeExecutorRunner{}
	uc := NewJobRunUsecase(
		fakeRunJobRepo{job: &Job{ID: 1, ExecutorID: 2, Name: "daily", CronExpr: "0 0 1 * * *", TimeoutSeconds: 30}},
		fakeRunGlueRepo{glue: &Glue{JobID: 1, Content: "echo hello"}},
		fakeRunExecutorRepo{executor: &Executor{ID: 2, Name: "exec", Address: "http://exec", TokenCiphertext: "cipher"}},
		jobLogRepo,
		fakeTokenCipher{},
		runner,
		JobRunConfig{PublicBaseURL: "http://admin", CallbackToken: "callback"},
		log.DefaultLogger,
	)

	got, err := uc.KillJob(context.Background(), 1)
	if err != nil {
		t.Fatalf("KillJob returned error: %v", err)
	}
	if got.Status != JobLogStatusKilling || runner.killReq == nil || runner.killReq.LogID != 9 {
		t.Fatalf("unexpected kill result=%+v req=%+v", got, runner.killReq)
	}
}

func TestJobRunUsecaseKillMarksFailedWhenExecutorKillFails(t *testing.T) {
	now := time.Now()
	jobLogRepo := &fakeRunJobLogRepo{running: &JobLog{ID: 9, JobID: 1, ExecutorID: 2, ExecutorAddress: "http://exec", Status: JobLogStatusRunning, StartTime: now}}
	runner := &fakeExecutorRunner{killErr: errors.New("kill failed")}
	uc := NewJobRunUsecase(
		fakeRunJobRepo{job: &Job{ID: 1, ExecutorID: 2, Name: "daily", CronExpr: "0 0 1 * * *", TimeoutSeconds: 30}},
		fakeRunGlueRepo{glue: &Glue{JobID: 1, Content: "echo hello"}},
		fakeRunExecutorRepo{executor: &Executor{ID: 2, Name: "exec", Address: "http://exec", TokenCiphertext: "cipher"}},
		jobLogRepo,
		fakeTokenCipher{},
		runner,
		JobRunConfig{PublicBaseURL: "http://admin", CallbackToken: "callback"},
		log.DefaultLogger,
	)

	_, err := uc.KillJob(context.Background(), 1)
	if err == nil {
		t.Fatal("expected kill error, got nil")
	}
	if jobLogRepo.updated == nil || jobLogRepo.updated.Status != JobLogStatusFailed {
		t.Fatalf("expected killing log marked failed, got %+v", jobLogRepo.updated)
	}
	if jobLogRepo.updated.EndTime == nil || jobLogRepo.updated.ErrorMessage == "" {
		t.Fatalf("expected failed kill end time and error message, got %+v", jobLogRepo.updated)
	}
}
