package service

import (
	"context"
	"strings"

	v1 "chronoFlow-admin/api/all-pb-go/v1"
	"chronoFlow-admin/internal/biz"
	httpErrors "chronoFlow-admin/internal/errors"
)

type SystemSettingsService struct {
	v1.UnimplementedSystemSettingsServer

	uc *biz.SystemSettingUsecase
}

func NewSystemSettingsService(uc *biz.SystemSettingUsecase) *SystemSettingsService {
	return &SystemSettingsService{uc: uc}
}

func (s *SystemSettingsService) GetAlertSettings(ctx context.Context, _ *v1.GetAlertSettingsRequest) (*v1.GetAlertSettingsReply, error) {
	settings, err := s.uc.GetAlertSettings(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.GetAlertSettingsReply{
		Code:    0,
		Message: successMessage("GetAlertSettings"),
		Data:    &v1.GetAlertSettingsReply_Data{Settings: toAlertSettingsInfo(settings)},
	}, nil
}

func (s *SystemSettingsService) SaveFeishuWebhook(ctx context.Context, req *v1.SaveFeishuWebhookRequest) (*v1.SaveFeishuWebhookReply, error) {
	if req == nil {
		return nil, httpErrors.E(httpErrors.ErrInvalidRequestBody)
	}
	settings, err := s.uc.SaveFeishuWebhook(ctx, strings.TrimSpace(req.GetWebhook()))
	if err != nil {
		return nil, err
	}
	return &v1.SaveFeishuWebhookReply{
		Code:    0,
		Message: successMessage("SaveFeishuWebhook"),
		Data:    &v1.SaveFeishuWebhookReply_Data{Settings: toAlertSettingsInfo(settings)},
	}, nil
}

func (s *SystemSettingsService) TestFeishuWebhook(context.Context, *v1.TestFeishuWebhookRequest) (*v1.TestFeishuWebhookReply, error) {
	return nil, httpErrors.EWithMessage(httpErrors.ErrConflict, "飞书测试发送器尚未初始化")
}

func (s *SystemSettingsService) ClearFeishuWebhook(ctx context.Context, _ *v1.ClearFeishuWebhookRequest) (*v1.ClearFeishuWebhookReply, error) {
	settings, err := s.uc.ClearFeishuWebhook(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.ClearFeishuWebhookReply{
		Code:    0,
		Message: successMessage("ClearFeishuWebhook"),
		Data:    &v1.ClearFeishuWebhookReply_Data{Settings: toAlertSettingsInfo(settings)},
	}, nil
}

func toAlertSettingsInfo(settings *biz.AlertSettings) *v1.AlertSettingsInfo {
	if settings == nil {
		return &v1.AlertSettingsInfo{}
	}
	updatedAt := ""
	if !settings.FeishuWebhookUpdatedAt.IsZero() {
		updatedAt = formatServiceTime(settings.FeishuWebhookUpdatedAt)
	}
	return &v1.AlertSettingsInfo{
		FeishuWebhookConfigured: settings.FeishuWebhookConfigured,
		FeishuWebhookUpdatedAt:  updatedAt,
	}
}
