package data

import (
	"context"
	"testing"

	"chronoFlow-admin/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newSystemSettingRepoForTest(t *testing.T) *SystemSettingRepo {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&SystemSetting{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	return NewSystemSettingRepo(&Data{db: db, log: log.NewHelper(log.DefaultLogger)}, log.DefaultLogger)
}

func TestSystemSettingRepoUpsertAndGetByKey(t *testing.T) {
	repo := newSystemSettingRepoForTest(t)
	ctx := context.Background()

	first, err := repo.Upsert(ctx, &biz.SystemSetting{SettingKey: "alert.feishu.webhook", ValueEncrypted: "cipher-1"})
	if err != nil {
		t.Fatalf("first Upsert returned error: %v", err)
	}
	if first.ValueEncrypted != "cipher-1" {
		t.Fatalf("expected first ciphertext, got %+v", first)
	}

	second, err := repo.Upsert(ctx, &biz.SystemSetting{SettingKey: "alert.feishu.webhook", ValueEncrypted: "cipher-2"})
	if err != nil {
		t.Fatalf("second Upsert returned error: %v", err)
	}
	if second.ID != first.ID {
		t.Fatalf("expected same row id %d, got %d", first.ID, second.ID)
	}
	if second.ValueEncrypted != "cipher-2" {
		t.Fatalf("expected updated ciphertext, got %+v", second)
	}

	got, err := repo.GetByKey(ctx, "alert.feishu.webhook")
	if err != nil {
		t.Fatalf("GetByKey returned error: %v", err)
	}
	if got == nil || got.ValueEncrypted != "cipher-2" {
		t.Fatalf("unexpected setting: %+v", got)
	}
}
