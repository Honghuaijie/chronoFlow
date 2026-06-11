package biz

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

type fakeJobLogRepo struct {
	items []*JobLog
}

func (r fakeJobLogRepo) List(_ context.Context, filter JobLogFilter) ([]*JobLog, int64, error) {
	items := make([]*JobLog, 0)
	for _, item := range r.items {
		if filter.JobID > 0 && item.JobID != filter.JobID {
			continue
		}
		cp := *item
		items = append(items, &cp)
	}
	return items, int64(len(items)), nil
}

func (r fakeJobLogRepo) GetByID(_ context.Context, id int64) (*JobLog, error) {
	for _, item := range r.items {
		if item.ID == id {
			cp := *item
			return &cp, nil
		}
	}
	return nil, nil
}

type fakeLogReader struct {
	content string
	err     error
}

func (r fakeLogReader) Read(context.Context, string) (string, error) {
	return r.content, r.err
}

func TestJobLogUsecaseDetailReadsLogContent(t *testing.T) {
	now := time.Now()
	uc := NewJobLogUsecase(fakeJobLogRepo{items: []*JobLog{
		{ID: 1, JobID: 10, JobName: "daily", Status: JobLogStatusSuccess, LogPath: "logs/1.log", GlueSnapshot: "echo hello", StartTime: now},
	}}, fakeLogReader{content: "hello\n"}, log.DefaultLogger)

	got, err := uc.GetJobLogDetail(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetJobLogDetail returned error: %v", err)
	}
	if got.LogContent != "hello\n" {
		t.Fatalf("expected log content, got %q", got.LogContent)
	}
	if got.GlueSnapshot != "echo hello" {
		t.Fatalf("expected glue snapshot, got %q", got.GlueSnapshot)
	}
}

func TestJobLogUsecaseDetailHandlesMissingLogFile(t *testing.T) {
	now := time.Now()
	uc := NewJobLogUsecase(fakeJobLogRepo{items: []*JobLog{
		{ID: 1, JobID: 10, JobName: "daily", Status: JobLogStatusSuccess, LogPath: "logs/missing.log", StartTime: now},
	}}, fakeLogReader{err: os.ErrNotExist}, log.DefaultLogger)

	got, err := uc.GetJobLogDetail(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetJobLogDetail returned error: %v", err)
	}
	if got.LogContent != MissingLogContentMessage {
		t.Fatalf("expected missing log message, got %q", got.LogContent)
	}
}

func TestJobLogUsecaseDetailReturnsReadError(t *testing.T) {
	now := time.Now()
	uc := NewJobLogUsecase(fakeJobLogRepo{items: []*JobLog{
		{ID: 1, JobID: 10, JobName: "daily", Status: JobLogStatusSuccess, LogPath: "logs/1.log", StartTime: now},
	}}, fakeLogReader{err: errors.New("permission denied")}, log.DefaultLogger)

	_, err := uc.GetJobLogDetail(context.Background(), 1)
	if err == nil {
		t.Fatal("expected read error, got nil")
	}
}
