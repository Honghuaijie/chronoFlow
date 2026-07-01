package data

import (
	"context"

	"chronoFlow-admin/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm/clause"
)

type SystemSettingRepo struct {
	data *Data
	log  *log.Helper
}

func NewSystemSettingRepo(data *Data, logger log.Logger) *SystemSettingRepo {
	return &SystemSettingRepo{data: data, log: log.NewHelper(logger)}
}

func (r *SystemSettingRepo) GetByKey(ctx context.Context, key string) (*biz.SystemSetting, error) {
	var model SystemSetting
	err := r.data.DB(ctx).WithContext(ctx).Where("setting_key = ?", key).First(&model).Error
	if err != nil {
		if isNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return toBizSystemSetting(&model), nil
}

func (r *SystemSettingRepo) Upsert(ctx context.Context, setting *biz.SystemSetting) (*biz.SystemSetting, error) {
	model := toSystemSettingModel(setting)
	if err := r.data.DB(ctx).WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "setting_key"}},
			DoUpdates: clause.AssignmentColumns([]string{"value_encrypted", "updated_at"}),
		}).
		Create(model).Error; err != nil {
		return nil, err
	}
	return r.GetByKey(ctx, setting.SettingKey)
}

func toSystemSettingModel(setting *biz.SystemSetting) *SystemSetting {
	if setting == nil {
		return nil
	}
	return &SystemSetting{
		ID:             uint64(setting.ID),
		SettingKey:     setting.SettingKey,
		ValueEncrypted: setting.ValueEncrypted,
	}
}

func toBizSystemSetting(model *SystemSetting) *biz.SystemSetting {
	if model == nil {
		return nil
	}
	return &biz.SystemSetting{
		ID:             int64(model.ID),
		SettingKey:     model.SettingKey,
		ValueEncrypted: model.ValueEncrypted,
		CreatedAt:      model.CreatedAt,
		UpdatedAt:      model.UpdatedAt,
	}
}
