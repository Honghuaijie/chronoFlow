package data

import (
	"context"

	"chronoFlow-admin/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

type ExecutorRepo struct {
	data *Data
	log  *log.Helper
}

func NewExecutorRepo(data *Data, logger log.Logger) *ExecutorRepo {
	return &ExecutorRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *ExecutorRepo) Create(ctx context.Context, executor *biz.Executor) (*biz.Executor, error) {
	model := toExecutorModel(executor)
	if err := r.data.DB(ctx).WithContext(ctx).Create(model).Error; err != nil {
		return nil, err
	}
	return toBizExecutor(model), nil
}

func (r *ExecutorRepo) GetByID(ctx context.Context, id int64) (*biz.Executor, error) {
	var model Executor
	err := r.data.DB(ctx).WithContext(ctx).First(&model, uint64(id)).Error
	if err != nil {
		if isNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return toBizExecutor(&model), nil
}

func (r *ExecutorRepo) List(ctx context.Context) ([]*biz.Executor, error) {
	var models []*Executor
	if err := r.data.DB(ctx).WithContext(ctx).Order("id asc").Find(&models).Error; err != nil {
		return nil, err
	}
	items := make([]*biz.Executor, 0, len(models))
	for _, model := range models {
		items = append(items, toBizExecutor(model))
	}
	return items, nil
}

func (r *ExecutorRepo) Update(ctx context.Context, executor *biz.Executor) (*biz.Executor, error) {
	db := r.data.DB(ctx).WithContext(ctx)
	var model Executor
	err := db.First(&model, uint64(executor.ID)).Error
	if err != nil {
		if isNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	model.Name = executor.Name
	model.Address = executor.Address
	model.TokenCiphertext = executor.TokenCiphertext
	model.Description = executor.Description
	model.Status = executor.Status
	model.HeartbeatFailCount = executor.HeartbeatFailCount
	model.LastHeartbeatTime = executor.LastHeartbeatTime
	if err := db.Save(&model).Error; err != nil {
		return nil, err
	}
	return toBizExecutor(&model), nil
}

func (r *ExecutorRepo) Delete(ctx context.Context, id int64) error {
	return r.data.DB(ctx).WithContext(ctx).Delete(&Executor{}, uint64(id)).Error
}

func toExecutorModel(executor *biz.Executor) *Executor {
	if executor == nil {
		return nil
	}
	return &Executor{
		ID:                 uint64(executor.ID),
		Name:               executor.Name,
		Address:            executor.Address,
		TokenCiphertext:    executor.TokenCiphertext,
		Description:        executor.Description,
		Status:             executor.Status,
		HeartbeatFailCount: executor.HeartbeatFailCount,
		LastHeartbeatTime:  executor.LastHeartbeatTime,
	}
}

func toBizExecutor(model *Executor) *biz.Executor {
	if model == nil {
		return nil
	}
	return &biz.Executor{
		ID:                 int64(model.ID),
		Name:               model.Name,
		Address:            model.Address,
		TokenCiphertext:    model.TokenCiphertext,
		Description:        model.Description,
		Status:             model.Status,
		HeartbeatFailCount: model.HeartbeatFailCount,
		LastHeartbeatTime:  model.LastHeartbeatTime,
		CreatedAt:          model.CreatedAt,
		UpdatedAt:          model.UpdatedAt,
	}
}
