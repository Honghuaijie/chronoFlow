package data

import (
	"context"
	"testing"
	"time"

	"chronoFlow-admin/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newChronoReposForTest(t *testing.T) (*Data, biz.JobRepo, biz.GlueRepo, biz.JobLogRepo) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&Job{}, &JobGlue{}, &JobLog{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	data := &Data{db: db, log: log.NewHelper(log.DefaultLogger)}
	return data, NewJobRepo(data, log.DefaultLogger), NewGlueRepo(data, log.DefaultLogger), NewJobLogRepo(data, log.DefaultLogger)
}

func TestJobRepoCreateListUpdateDelete(t *testing.T) {
	_, jobRepo, _, _ := newChronoReposForTest(t)
	ctx := context.Background()

	created, err := jobRepo.Create(ctx, &biz.Job{
		ExecutorID:     10,
		Name:           "daily",
		CronExpr:       "0 0 1 * * *",
		TimeoutSeconds: 30,
		ScheduleStatus: biz.ScheduleStatusStopped,
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	items, err := jobRepo.List(ctx, 10)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(items) != 1 || items[0].ID != created.ID {
		t.Fatalf("unexpected list items: %+v", items)
	}

	created.ScheduleStatus = biz.ScheduleStatusRunning
	updated, err := jobRepo.Update(ctx, created)
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if updated.ScheduleStatus != biz.ScheduleStatusRunning {
		t.Fatalf("expected running, got %q", updated.ScheduleStatus)
	}

	if err := jobRepo.Delete(ctx, created.ID); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	deleted, err := jobRepo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}
	if deleted != nil {
		t.Fatalf("expected nil after delete, got %+v", deleted)
	}
}

func TestGlueRepoSaveUpsertsByJobID(t *testing.T) {
	_, _, glueRepo, _ := newChronoReposForTest(t)
	ctx := context.Background()

	first, err := glueRepo.Save(ctx, &biz.Glue{JobID: 1, Content: "echo one"})
	if err != nil {
		t.Fatalf("first Save returned error: %v", err)
	}
	second, err := glueRepo.Save(ctx, &biz.Glue{JobID: 1, Content: "echo two"})
	if err != nil {
		t.Fatalf("second Save returned error: %v", err)
	}
	if first.ID != second.ID {
		t.Fatalf("expected upsert to keep id %d, got %d", first.ID, second.ID)
	}
	got, err := glueRepo.GetByJobID(ctx, 1)
	if err != nil {
		t.Fatalf("GetByJobID returned error: %v", err)
	}
	if got.Content != "echo two" {
		t.Fatalf("expected updated content, got %q", got.Content)
	}
}

func TestJobLogRepoListAndGet(t *testing.T) {
	data, _, _, jobLogRepo := newChronoReposForTest(t)
	ctx := context.Background()
	now := time.Now()
	model := &JobLog{
		JobID:           1,
		JobName:         "daily",
		ExecutorID:      2,
		ExecutorName:    "exec",
		ExecutorAddress: "http://exec",
		CronExpr:        "0 0 1 * * *",
		TimeoutSeconds:  30,
		GlueSnapshot:    "echo hello",
		TriggerType:     biz.TriggerTypeManual,
		Status:          biz.JobLogStatusSuccess,
		StartTime:       now,
		LogPath:         "logs/1.log",
	}
	if err := data.DB(ctx).Create(model).Error; err != nil {
		t.Fatalf("create job log: %v", err)
	}

	items, total, err := jobLogRepo.List(ctx, biz.JobLogFilter{JobID: 1, Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Fatalf("expected 1 item, total=%d items=%d", total, len(items))
	}
	got, err := jobLogRepo.GetByID(ctx, items[0].ID)
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}
	if got.GlueSnapshot != "echo hello" || got.LogPath != "logs/1.log" {
		t.Fatalf("unexpected job log: %+v", got)
	}
}

func TestJobLogRepoCreateGetRunningAndUpdate(t *testing.T) {
	_, _, _, jobLogRepo := newChronoReposForTest(t)
	ctx := context.Background()
	now := time.Now()

	created, err := jobLogRepo.(interface {
		Create(context.Context, *biz.JobLog) (*biz.JobLog, error)
	}).Create(ctx, &biz.JobLog{
		JobID:           1,
		JobName:         "daily",
		ExecutorID:      2,
		ExecutorName:    "exec",
		ExecutorAddress: "http://exec",
		CronExpr:        "0 0 1 * * *",
		TimeoutSeconds:  30,
		GlueSnapshot:    "echo hello",
		TriggerType:     biz.TriggerTypeManual,
		Status:          biz.JobLogStatusRunning,
		StartTime:       now,
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	running, err := jobLogRepo.(interface {
		GetRunningByJobID(context.Context, int64) (*biz.JobLog, error)
	}).GetRunningByJobID(ctx, 1)
	if err != nil {
		t.Fatalf("GetRunningByJobID returned error: %v", err)
	}
	if running == nil || running.ID != created.ID {
		t.Fatalf("expected running log %d, got %+v", created.ID, running)
	}

	endTime := now.Add(time.Second)
	created.Status = biz.JobLogStatusSuccess
	created.EndTime = &endTime
	created.LogPath = "logs/1.log"
	updated, err := jobLogRepo.(interface {
		Update(context.Context, *biz.JobLog) (*biz.JobLog, error)
	}).Update(ctx, created)
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if updated.Status != biz.JobLogStatusSuccess || updated.LogPath != "logs/1.log" {
		t.Fatalf("unexpected updated log: %+v", updated)
	}
}
