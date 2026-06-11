package data

import (
	"context"
	"testing"

	"chronoFlow-admin/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newExecutorRepoForTest(t *testing.T) biz.ExecutorRepo {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&Executor{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	return NewExecutorRepo(&Data{db: db, log: log.NewHelper(log.DefaultLogger)}, log.DefaultLogger)
}

func TestExecutorRepoCreateGetListUpdateDelete(t *testing.T) {
	repo := newExecutorRepoForTest(t)
	ctx := context.Background()

	created, err := repo.Create(ctx, &biz.Executor{
		Name:            "exec-a",
		Address:         "http://127.0.0.1:19090",
		TokenCiphertext: "cipher",
		Description:     "desc",
		Status:          biz.ExecutorStatusOffline,
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if created.ID <= 0 {
		t.Fatalf("expected positive id, got %d", created.ID)
	}

	got, err := repo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}
	if got == nil || got.Name != "exec-a" || got.TokenCiphertext != "cipher" {
		t.Fatalf("unexpected executor: %+v", got)
	}

	got.Name = "exec-b"
	updated, err := repo.Update(ctx, got)
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if updated.Name != "exec-b" {
		t.Fatalf("expected updated name, got %q", updated.Name)
	}

	items, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	if err := repo.Delete(ctx, created.ID); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	deleted, err := repo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetByID after delete returned error: %v", err)
	}
	if deleted != nil {
		t.Fatalf("expected nil after delete, got %+v", deleted)
	}
}
