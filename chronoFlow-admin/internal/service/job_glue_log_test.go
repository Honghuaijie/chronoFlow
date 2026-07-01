package service

import (
	"context"
	"testing"
	"time"

	v1 "chronoFlow-admin/api/all-pb-go/v1"
	"chronoFlow-admin/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

func TestJobServiceCreateJob(t *testing.T) {
	jobRepo := &serviceJobRepo{items: map[int64]*biz.Job{}}
	uc := biz.NewJobUsecase(jobRepo, &serviceGlueRepo{item: &biz.Glue{JobID: 1, Content: "echo hello"}}, log.DefaultLogger)
	svc := NewJobService(uc, nil, nil)

	reply, err := svc.CreateJob(context.Background(), &v1.CreateJobRequest{
		ExecutorId:          10,
		Name:                "daily",
		CronExpr:            "0 0 1 * * *",
		TimeoutSeconds:      30,
		FailureAlertEnabled: true,
	})
	if err != nil {
		t.Fatalf("CreateJob returned error: %v", err)
	}
	if reply.GetData().GetJob().GetScheduleStatus() != biz.ScheduleStatusStopped {
		t.Fatalf("expected stopped, got %q", reply.GetData().GetJob().GetScheduleStatus())
	}
	if !reply.GetData().GetJob().GetFailureAlertEnabled() {
		t.Fatal("expected failure alert enabled")
	}
}

func TestGlueServiceSaveGlue(t *testing.T) {
	uc := biz.NewGlueUsecase(&serviceGlueRepo{}, log.DefaultLogger)
	svc := NewGlueService(uc)

	reply, err := svc.SaveGlue(context.Background(), &v1.SaveGlueRequest{JobId: 1, Content: " echo hi "})
	if err != nil {
		t.Fatalf("SaveGlue returned error: %v", err)
	}
	if reply.GetData().GetGlue().GetContent() != "echo hi" {
		t.Fatalf("expected trimmed content, got %q", reply.GetData().GetGlue().GetContent())
	}
}

func TestJobLogServiceDetail(t *testing.T) {
	now := time.Now()
	alertSentAt := now.Add(time.Minute)
	uc := biz.NewJobLogUsecase(serviceJobLogRepo{items: []*biz.JobLog{
		{
			ID:                   1,
			JobID:                10,
			JobName:              "daily",
			Status:               biz.JobLogStatusSuccess,
			StartTime:            now,
			LogPath:              "logs/1.log",
			GlueSnapshot:         "echo hi",
			AlertEnabledSnapshot: true,
			AlertStatus:          "sent",
			AlertSentAt:          &alertSentAt,
		},
	}}, serviceLogReader{content: "hello"}, log.DefaultLogger)
	svc := NewJobLogService(uc)

	reply, err := svc.GetJobLogDetail(context.Background(), &v1.GetJobLogDetailRequest{Id: 1})
	if err != nil {
		t.Fatalf("GetJobLogDetail returned error: %v", err)
	}
	if reply.GetData().GetLogContent() != "hello" || reply.GetData().GetGlueSnapshot() != "echo hi" {
		t.Fatalf("unexpected reply: %+v", reply)
	}
	if !reply.GetData().GetLog().GetAlertEnabledSnapshot() {
		t.Fatal("expected alert snapshot enabled")
	}
	if reply.GetData().GetLog().GetAlertStatus() != "sent" {
		t.Fatalf("expected alert sent, got %q", reply.GetData().GetLog().GetAlertStatus())
	}
	if reply.GetData().GetLog().GetAlertSentAt() == "" {
		t.Fatal("expected alert sent time")
	}
}

type serviceJobRepo struct {
	items map[int64]*biz.Job
}

func (r *serviceJobRepo) Create(_ context.Context, job *biz.Job) (*biz.Job, error) {
	cp := *job
	cp.ID = int64(len(r.items) + 1)
	r.items[cp.ID] = &cp
	return &cp, nil
}

func (r *serviceJobRepo) GetByID(_ context.Context, id int64) (*biz.Job, error) {
	item := r.items[id]
	if item == nil {
		return nil, nil
	}
	cp := *item
	return &cp, nil
}

func (r *serviceJobRepo) List(_ context.Context, executorID int64) ([]*biz.Job, error) {
	items := make([]*biz.Job, 0)
	for _, item := range r.items {
		if executorID == 0 || item.ExecutorID == executorID {
			cp := *item
			items = append(items, &cp)
		}
	}
	return items, nil
}

func (r *serviceJobRepo) Update(_ context.Context, job *biz.Job) (*biz.Job, error) {
	cp := *job
	r.items[cp.ID] = &cp
	return &cp, nil
}

func (r *serviceJobRepo) Delete(_ context.Context, id int64) error {
	delete(r.items, id)
	return nil
}

type serviceGlueRepo struct {
	item *biz.Glue
}

func (r serviceGlueRepo) GetByJobID(_ context.Context, jobID int64) (*biz.Glue, error) {
	if r.item == nil || r.item.JobID != jobID {
		return nil, nil
	}
	cp := *r.item
	return &cp, nil
}

func (r *serviceGlueRepo) Save(_ context.Context, glue *biz.Glue) (*biz.Glue, error) {
	cp := *glue
	cp.ID = 1
	r.item = &cp
	return &cp, nil
}

type serviceJobLogRepo struct {
	items []*biz.JobLog
}

func (r serviceJobLogRepo) List(_ context.Context, _ biz.JobLogFilter) ([]*biz.JobLog, int64, error) {
	return r.items, int64(len(r.items)), nil
}

func (r serviceJobLogRepo) GetByID(_ context.Context, id int64) (*biz.JobLog, error) {
	for _, item := range r.items {
		if item.ID == id {
			cp := *item
			return &cp, nil
		}
	}
	return nil, nil
}

type serviceLogReader struct {
	content string
}

func (r serviceLogReader) Read(context.Context, string) (string, error) {
	return r.content, nil
}
