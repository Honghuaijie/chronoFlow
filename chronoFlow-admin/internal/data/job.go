package data

import (
	"context"

	"chronoFlow-admin/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

type JobRepo struct {
	data *Data
	log  *log.Helper
}

func NewJobRepo(data *Data, logger log.Logger) *JobRepo {
	return &JobRepo{data: data, log: log.NewHelper(logger)}
}

func (r *JobRepo) Create(ctx context.Context, job *biz.Job) (*biz.Job, error) {
	model := toJobModel(job)
	if err := r.data.DB(ctx).WithContext(ctx).Create(model).Error; err != nil {
		return nil, err
	}
	return toBizJob(model), nil
}

func (r *JobRepo) GetByID(ctx context.Context, id int64) (*biz.Job, error) {
	var model Job
	err := r.data.DB(ctx).WithContext(ctx).First(&model, uint64(id)).Error
	if err != nil {
		if isNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return toBizJob(&model), nil
}

func (r *JobRepo) List(ctx context.Context, executorID int64) ([]*biz.Job, error) {
	var models []*Job
	db := r.data.DB(ctx).WithContext(ctx)
	if executorID > 0 {
		db = db.Where("executor_id = ?", executorID)
	}
	if err := db.Order("id desc").Find(&models).Error; err != nil {
		return nil, err
	}
	items := make([]*biz.Job, 0, len(models))
	for _, model := range models {
		items = append(items, toBizJob(model))
	}
	return items, nil
}

func (r *JobRepo) Update(ctx context.Context, job *biz.Job) (*biz.Job, error) {
	db := r.data.DB(ctx).WithContext(ctx)
	var model Job
	err := db.First(&model, uint64(job.ID)).Error
	if err != nil {
		if isNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	model.ExecutorID = uint64(job.ExecutorID)
	model.Name = job.Name
	model.CronExpr = job.CronExpr
	model.TimeoutSeconds = job.TimeoutSeconds
	model.ScheduleStatus = job.ScheduleStatus
	model.Description = job.Description
	if err := db.Save(&model).Error; err != nil {
		return nil, err
	}
	return toBizJob(&model), nil
}

func (r *JobRepo) Delete(ctx context.Context, id int64) error {
	return r.data.DB(ctx).WithContext(ctx).Delete(&Job{}, uint64(id)).Error
}

func toJobModel(job *biz.Job) *Job {
	if job == nil {
		return nil
	}
	return &Job{
		ID:             uint64(job.ID),
		ExecutorID:     uint64(job.ExecutorID),
		Name:           job.Name,
		CronExpr:       job.CronExpr,
		TimeoutSeconds: job.TimeoutSeconds,
		ScheduleStatus: job.ScheduleStatus,
		Description:    job.Description,
	}
}

func toBizJob(model *Job) *biz.Job {
	if model == nil {
		return nil
	}
	return &biz.Job{
		ID:             int64(model.ID),
		ExecutorID:     int64(model.ExecutorID),
		Name:           model.Name,
		CronExpr:       model.CronExpr,
		TimeoutSeconds: model.TimeoutSeconds,
		ScheduleStatus: model.ScheduleStatus,
		Description:    model.Description,
		CreatedAt:      model.CreatedAt,
		UpdatedAt:      model.UpdatedAt,
	}
}
