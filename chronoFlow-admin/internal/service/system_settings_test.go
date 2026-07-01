package service

import (
	"context"
	"encoding/base64"
	"strings"
	"testing"
	"time"

	v1 "chronoFlow-admin/api/all-pb-go/v1"
	"chronoFlow-admin/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

type fakeSystemSettingRepo struct {
	item *biz.SystemSetting
}

func (r *fakeSystemSettingRepo) GetByKey(_ context.Context, key string) (*biz.SystemSetting, error) {
	if r.item == nil || r.item.SettingKey != key {
		return nil, nil
	}
	cp := *r.item
	return &cp, nil
}

func (r *fakeSystemSettingRepo) Upsert(_ context.Context, setting *biz.SystemSetting) (*biz.SystemSetting, error) {
	cp := *setting
	if cp.ID == 0 {
		cp.ID = 1
	}
	cp.UpdatedAt = time.Now()
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

func TestSystemSettingsServiceSaveAndClearWebhook(t *testing.T) {
	uc := biz.NewSystemSettingUsecase(&fakeSystemSettingRepo{}, fakeSettingCipher{}, log.DefaultLogger)
	svc := NewSystemSettingsService(uc, nil)

	saveReply, err := svc.SaveFeishuWebhook(context.Background(), &v1.SaveFeishuWebhookRequest{
		Webhook: "https://open.feishu.cn/open-apis/bot/v2/hook/abc",
	})
	if err != nil {
		t.Fatalf("SaveFeishuWebhook returned error: %v", err)
	}
	if !saveReply.GetData().GetSettings().GetFeishuWebhookConfigured() {
		t.Fatal("expected webhook configured")
	}
	if saveReply.GetData().GetSettings().GetFeishuWebhookUpdatedAt() == "" {
		t.Fatal("expected updated_at")
	}

	getReply, err := svc.GetAlertSettings(context.Background(), &v1.GetAlertSettingsRequest{})
	if err != nil {
		t.Fatalf("GetAlertSettings returned error: %v", err)
	}
	if !getReply.GetData().GetSettings().GetFeishuWebhookConfigured() {
		t.Fatal("expected webhook configured in get")
	}

	clearReply, err := svc.ClearFeishuWebhook(context.Background(), &v1.ClearFeishuWebhookRequest{})
	if err != nil {
		t.Fatalf("ClearFeishuWebhook returned error: %v", err)
	}
	if clearReply.GetData().GetSettings().GetFeishuWebhookConfigured() {
		t.Fatal("expected webhook unconfigured")
	}
}
