package biz

import (
	"context"
	"testing"

	"github.com/go-kratos/kratos/v2/log"
)

type fakeGlueRepo struct {
	item *Glue
}

func (r *fakeGlueRepo) GetByJobID(_ context.Context, jobID int64) (*Glue, error) {
	if r.item == nil || r.item.JobID != jobID {
		return nil, nil
	}
	cp := *r.item
	return &cp, nil
}

func (r *fakeGlueRepo) Save(_ context.Context, glue *Glue) (*Glue, error) {
	cp := *glue
	if r.item != nil && r.item.JobID == glue.JobID {
		cp.ID = r.item.ID
	} else {
		cp.ID = 1
	}
	r.item = &cp
	return &cp, nil
}

func TestGlueUsecaseSaveTrimsAndUpserts(t *testing.T) {
	repo := &fakeGlueRepo{item: &Glue{ID: 9, JobID: 1, Content: "old"}}
	uc := NewGlueUsecase(repo, log.DefaultLogger)

	got, err := uc.SaveGlue(context.Background(), 1, "  echo hello  ")
	if err != nil {
		t.Fatalf("SaveGlue returned error: %v", err)
	}
	if got.ID != 9 {
		t.Fatalf("expected existing glue id to be kept, got %d", got.ID)
	}
	if got.Content != "echo hello" {
		t.Fatalf("expected trimmed content, got %q", got.Content)
	}
}

func TestGlueUsecaseRejectsEmptyContent(t *testing.T) {
	uc := NewGlueUsecase(&fakeGlueRepo{}, log.DefaultLogger)

	_, err := uc.SaveGlue(context.Background(), 1, "  ")
	if err == nil {
		t.Fatal("expected empty content error, got nil")
	}
}
