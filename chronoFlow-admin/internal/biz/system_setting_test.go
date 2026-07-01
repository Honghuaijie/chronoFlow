package biz

import (
	"context"
	"encoding/base64"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

type fakeSystemSettingRepo struct {
	item            *SystemSetting
	savedCiphertext string
}

func (r *fakeSystemSettingRepo) GetByKey(_ context.Context, key string) (*SystemSetting, error) {
	if r.item == nil || r.item.SettingKey != key {
		return nil, nil
	}
	cp := *r.item
	return &cp, nil
}

func (r *fakeSystemSettingRepo) Upsert(_ context.Context, setting *SystemSetting) (*SystemSetting, error) {
	cp := *setting
	if cp.ID == 0 {
		cp.ID = 1
	}
	cp.UpdatedAt = time.Now()
	r.savedCiphertext = cp.ValueEncrypted
	r.item = &cp
	return &cp, nil
}

type fakeSettingCipher struct{}

func (fakeSettingCipher) Encrypt(plaintext string) (string, error) {
	return "enc:" + base64.StdEncoding.EncodeToString([]byte(plaintext)), nil
}

func (fakeSettingCipher) Decrypt(ciphertext string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(ciphertext, "enc:"))
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

func TestSystemSettingUsecaseSaveGetAndClearFeishuWebhook(t *testing.T) {
	ctx := context.Background()
	repo := &fakeSystemSettingRepo{}
	uc := NewSystemSettingUsecase(repo, fakeSettingCipher{}, log.NewStdLogger(io.Discard))

	settings, err := uc.SaveFeishuWebhook(ctx, "https://open.feishu.cn/open-apis/bot/v2/hook/abc")
	if err != nil {
		t.Fatalf("SaveFeishuWebhook returned error: %v", err)
	}
	if !settings.FeishuWebhookConfigured {
		t.Fatal("expected webhook configured")
	}
	if strings.Contains(repo.savedCiphertext, "open.feishu.cn") {
		t.Fatalf("expected encrypted value to avoid plaintext, got %q", repo.savedCiphertext)
	}

	webhook, ok, err := uc.GetFeishuWebhook(ctx)
	if err != nil {
		t.Fatalf("GetFeishuWebhook returned error: %v", err)
	}
	if !ok || webhook != "https://open.feishu.cn/open-apis/bot/v2/hook/abc" {
		t.Fatalf("unexpected webhook ok=%v value=%q", ok, webhook)
	}

	settings, err = uc.ClearFeishuWebhook(ctx)
	if err != nil {
		t.Fatalf("ClearFeishuWebhook returned error: %v", err)
	}
	if settings.FeishuWebhookConfigured {
		t.Fatal("expected webhook unconfigured")
	}
}

func TestSystemSettingUsecaseSaveFeishuWebhookRejectsInvalidURL(t *testing.T) {
	uc := NewSystemSettingUsecase(&fakeSystemSettingRepo{}, fakeSettingCipher{}, log.NewStdLogger(io.Discard))

	if _, err := uc.SaveFeishuWebhook(context.Background(), "not-url"); err == nil {
		t.Fatal("expected invalid url error")
	}
	if _, err := uc.SaveFeishuWebhook(context.Background(), "ftp://example.com/hook"); err == nil {
		t.Fatal("expected unsupported scheme error")
	}
}
