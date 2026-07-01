package biz

import (
	"context"
	"net/url"
	"strings"
	"time"

	httpErrors "chronoFlow-admin/internal/errors"

	"github.com/go-kratos/kratos/v2/log"
)

const FeishuWebhookSettingKey = "alert.feishu.webhook"

type SystemSetting struct {
	ID             int64
	SettingKey     string
	ValueEncrypted string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type AlertSettings struct {
	FeishuWebhookConfigured bool
	FeishuWebhookUpdatedAt  time.Time
}

type SystemSettingRepo interface {
	GetByKey(context.Context, string) (*SystemSetting, error)
	Upsert(context.Context, *SystemSetting) (*SystemSetting, error)
}

type SystemSettingUsecase struct {
	repo   SystemSettingRepo
	cipher TokenCipher
	log    *log.Helper
}

func NewSystemSettingUsecase(repo SystemSettingRepo, cipher TokenCipher, logger log.Logger) *SystemSettingUsecase {
	return &SystemSettingUsecase{
		repo:   repo,
		cipher: cipher,
		log:    log.NewHelper(logger),
	}
}

func (uc *SystemSettingUsecase) GetAlertSettings(ctx context.Context) (*AlertSettings, error) {
	setting, err := uc.repo.GetByKey(ctx, FeishuWebhookSettingKey)
	if err != nil {
		return nil, err
	}
	return toAlertSettings(setting), nil
}

func (uc *SystemSettingUsecase) SaveFeishuWebhook(ctx context.Context, webhook string) (*AlertSettings, error) {
	webhook = strings.TrimSpace(webhook)
	if err := validateWebhookURL(webhook); err != nil {
		return nil, err
	}
	ciphertext, err := uc.cipher.Encrypt(webhook)
	if err != nil {
		return nil, err
	}
	setting, err := uc.repo.Upsert(ctx, &SystemSetting{
		SettingKey:     FeishuWebhookSettingKey,
		ValueEncrypted: ciphertext,
	})
	if err != nil {
		return nil, err
	}
	return toAlertSettings(setting), nil
}

func (uc *SystemSettingUsecase) ClearFeishuWebhook(ctx context.Context) (*AlertSettings, error) {
	setting, err := uc.repo.Upsert(ctx, &SystemSetting{
		SettingKey:     FeishuWebhookSettingKey,
		ValueEncrypted: "",
	})
	if err != nil {
		return nil, err
	}
	return toAlertSettings(setting), nil
}

func (uc *SystemSettingUsecase) GetFeishuWebhook(ctx context.Context) (string, bool, error) {
	setting, err := uc.repo.GetByKey(ctx, FeishuWebhookSettingKey)
	if err != nil {
		return "", false, err
	}
	if setting == nil || strings.TrimSpace(setting.ValueEncrypted) == "" {
		return "", false, nil
	}
	plaintext, err := uc.cipher.Decrypt(setting.ValueEncrypted)
	if err != nil {
		return "", false, err
	}
	return plaintext, true, nil
}

func toAlertSettings(setting *SystemSetting) *AlertSettings {
	if setting == nil || strings.TrimSpace(setting.ValueEncrypted) == "" {
		return &AlertSettings{}
	}
	return &AlertSettings{
		FeishuWebhookConfigured: true,
		FeishuWebhookUpdatedAt:  setting.UpdatedAt,
	}
}

func validateWebhookURL(value string) error {
	if value == "" {
		return httpErrors.EWithMessage(httpErrors.ErrMissingRequiredField, "飞书 Webhook 不能为空")
	}
	parsed, err := url.ParseRequestURI(value)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return httpErrors.EWithMessage(httpErrors.ErrInvalidParam, "飞书 Webhook URL 格式错误")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return httpErrors.EWithMessage(httpErrors.ErrInvalidParam, "飞书 Webhook 仅支持 http 或 https")
	}
	return nil
}
