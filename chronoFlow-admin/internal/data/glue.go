package data

import (
	"context"

	"chronoFlow-admin/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

type GlueRepo struct {
	data *Data
	log  *log.Helper
}

func NewGlueRepo(data *Data, logger log.Logger) *GlueRepo {
	return &GlueRepo{data: data, log: log.NewHelper(logger)}
}

func (r *GlueRepo) GetByJobID(ctx context.Context, jobID int64) (*biz.Glue, error) {
	var model JobGlue
	err := r.data.DB(ctx).WithContext(ctx).Where("job_id = ?", jobID).First(&model).Error
	if err != nil {
		if isNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return toBizGlue(&model), nil
}

func (r *GlueRepo) Save(ctx context.Context, glue *biz.Glue) (*biz.Glue, error) {
	db := r.data.DB(ctx).WithContext(ctx)
	var model JobGlue
	err := db.Where("job_id = ?", glue.JobID).First(&model).Error
	if err != nil && !isNotFound(err) {
		return nil, err
	}
	if isNotFound(err) {
		model = JobGlue{
			JobID:   uint64(glue.JobID),
			Content: glue.Content,
		}
		if err := db.Create(&model).Error; err != nil {
			return nil, err
		}
		return toBizGlue(&model), nil
	}
	model.Content = glue.Content
	if err := db.Save(&model).Error; err != nil {
		return nil, err
	}
	return toBizGlue(&model), nil
}

func toBizGlue(model *JobGlue) *biz.Glue {
	if model == nil {
		return nil
	}
	return &biz.Glue{
		ID:        int64(model.ID),
		JobID:     int64(model.JobID),
		Content:   model.Content,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}
}
