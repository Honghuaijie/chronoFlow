package biz

import (
	"context"
	"testing"

	"github.com/go-kratos/kratos/v2/log"
)

type fakeJobRepo struct {
	items map[int64]*Job
}

func (r *fakeJobRepo) Create(_ context.Context, job *Job) (*Job, error) {
	cp := *job
	cp.ID = int64(len(r.items) + 1)
	r.items[cp.ID] = &cp
	return &cp, nil
}

func (r *fakeJobRepo) GetByID(_ context.Context, id int64) (*Job, error) {
	item := r.items[id]
	if item == nil {
		return nil, nil
	}
	cp := *item
	return &cp, nil
}

func (r *fakeJobRepo) List(_ context.Context, executorID int64) ([]*Job, error) {
	items := make([]*Job, 0)
	for _, item := range r.items {
		if executorID == 0 || item.ExecutorID == executorID {
			cp := *item
			items = append(items, &cp)
		}
	}
	return items, nil
}

func (r *fakeJobRepo) Update(_ context.Context, job *Job) (*Job, error) {
	cp := *job
	r.items[cp.ID] = &cp
	return &cp, nil
}

func (r *fakeJobRepo) Delete(_ context.Context, id int64) error {
	delete(r.items, id)
	return nil
}

type fakeGlueExistRepo struct {
	hasGlue bool
}

func (r fakeGlueExistRepo) GetByJobID(context.Context, int64) (*Glue, error) {
	if !r.hasGlue {
		return nil, nil
	}
	return &Glue{ID: 1, JobID: 1, Content: "echo hello"}, nil
}

func (fakeGlueExistRepo) Save(context.Context, *Glue) (*Glue, error) {
	return nil, nil
}

func TestJobUsecaseCreateDefaultsStopped(t *testing.T) {
	jobRepo := &fakeJobRepo{items: map[int64]*Job{}}
	uc := NewJobUsecase(jobRepo, fakeGlueExistRepo{}, log.DefaultLogger)

	got, err := uc.CreateJob(context.Background(), &CreateJobInput{
		ExecutorID:     10,
		Name:           "  daily  ",
		CronExpr:       "0 0 1 * * *",
		TimeoutSeconds: 30,
		Description:    " desc ",
	})
	if err != nil {
		t.Fatalf("CreateJob returned error: %v", err)
	}
	if got.ScheduleStatus != ScheduleStatusStopped {
		t.Fatalf("expected stopped, got %q", got.ScheduleStatus)
	}
	if got.Name != "daily" || got.Description != "desc" {
		t.Fatalf("expected normalized fields, got %+v", got)
	}
}

func TestJobUsecaseRejectsInvalidCron(t *testing.T) {
	uc := NewJobUsecase(&fakeJobRepo{items: map[int64]*Job{}}, fakeGlueExistRepo{}, log.DefaultLogger)

	_, err := uc.CreateJob(context.Background(), &CreateJobInput{
		ExecutorID:     10,
		Name:           "daily",
		CronExpr:       "* * * * *",
		TimeoutSeconds: 30,
	})
	if err == nil {
		t.Fatal("expected invalid cron error, got nil")
	}
}

func TestJobUsecaseStartRequiresGlue(t *testing.T) {
	jobRepo := &fakeJobRepo{items: map[int64]*Job{
		1: {ID: 1, ExecutorID: 10, Name: "daily", CronExpr: "0 0 1 * * *", TimeoutSeconds: 30, ScheduleStatus: ScheduleStatusStopped},
	}}
	uc := NewJobUsecase(jobRepo, fakeGlueExistRepo{hasGlue: false}, log.DefaultLogger)

	_, err := uc.StartJob(context.Background(), 1)
	if err == nil {
		t.Fatal("expected missing glue error, got nil")
	}
}

func TestJobUsecaseStartAndStop(t *testing.T) {
	jobRepo := &fakeJobRepo{items: map[int64]*Job{
		1: {ID: 1, ExecutorID: 10, Name: "daily", CronExpr: "0 0 1 * * *", TimeoutSeconds: 30, ScheduleStatus: ScheduleStatusStopped},
	}}
	uc := NewJobUsecase(jobRepo, fakeGlueExistRepo{hasGlue: true}, log.DefaultLogger)

	started, err := uc.StartJob(context.Background(), 1)
	if err != nil {
		t.Fatalf("StartJob returned error: %v", err)
	}
	if started.ScheduleStatus != ScheduleStatusRunning {
		t.Fatalf("expected running, got %q", started.ScheduleStatus)
	}

	stopped, err := uc.StopJob(context.Background(), 1)
	if err != nil {
		t.Fatalf("StopJob returned error: %v", err)
	}
	if stopped.ScheduleStatus != ScheduleStatusStopped {
		t.Fatalf("expected stopped, got %q", stopped.ScheduleStatus)
	}
}
