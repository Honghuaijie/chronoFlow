package data

import (
	"context"
	"time"

	"chronoFlow-admin/internal/biz"
	httpErrors "chronoFlow-admin/internal/errors"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type JobLogRepo struct {
	data *Data
	log  *log.Helper
}

func NewJobLogRepo(data *Data, logger log.Logger) *JobLogRepo {
	return &JobLogRepo{data: data, log: log.NewHelper(logger)}
}

func (r *JobLogRepo) Create(ctx context.Context, jobLog *biz.JobLog) (*biz.JobLog, error) {
	model := toJobLogModel(jobLog)
	if err := r.data.DB(ctx).WithContext(ctx).Create(model).Error; err != nil {
		return nil, err
	}
	return toBizJobLog(model), nil
}

func (r *JobLogRepo) CreateRunningIfNoActive(ctx context.Context, jobLog *biz.JobLog) (*biz.JobLog, error) {
	var created *biz.JobLog
	err := r.data.DB(ctx).WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var job Job
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&job, uint64(jobLog.JobID)).Error; err != nil {
			return err
		}
		var active JobLog
		err := tx.Where("job_id = ? AND status IN ?", jobLog.JobID, []string{biz.JobLogStatusRunning, biz.JobLogStatusKilling}).
			Order("id desc").
			First(&active).Error
		if err == nil {
			return httpErrors.EWithMessage(httpErrors.ErrConflict, "任务正在执行中")
		}
		if !isNotFound(err) {
			return err
		}
		model := toJobLogModel(jobLog)
		if err := tx.Create(model).Error; err != nil {
			return err
		}
		created = toBizJobLog(model)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (r *JobLogRepo) GetRunningByJobID(ctx context.Context, jobID int64) (*biz.JobLog, error) {
	var model JobLog
	err := r.data.DB(ctx).WithContext(ctx).
		Where("job_id = ? AND status IN ?", jobID, []string{biz.JobLogStatusRunning, biz.JobLogStatusKilling}).
		Order("id desc").
		First(&model).Error
	if err != nil {
		if isNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return toBizJobLog(&model), nil
}

func (r *JobLogRepo) List(ctx context.Context, filter biz.JobLogFilter) ([]*biz.JobLog, int64, error) {
	db := r.data.DB(ctx).WithContext(ctx).Model(&JobLog{})
	if filter.JobID > 0 {
		db = db.Where("job_id = ?", filter.JobID)
	}
	if filter.ExecutorID > 0 {
		db = db.Where("executor_id = ?", filter.ExecutorID)
	}
	if filter.Status != "" {
		db = db.Where("status = ?", filter.Status)
	}
	if filter.TriggerType != "" {
		db = db.Where("trigger_type = ?", filter.TriggerType)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	page := filter.Page
	if page <= 0 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	var models []*JobLog
	if err := db.Order("id desc").Offset(int((page - 1) * pageSize)).Limit(int(pageSize)).Find(&models).Error; err != nil {
		return nil, 0, err
	}
	items := make([]*biz.JobLog, 0, len(models))
	for _, model := range models {
		items = append(items, toBizJobLog(model))
	}
	return items, total, nil
}

func (r *JobLogRepo) GetByID(ctx context.Context, id int64) (*biz.JobLog, error) {
	var model JobLog
	err := r.data.DB(ctx).WithContext(ctx).First(&model, uint64(id)).Error
	if err != nil {
		if isNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return toBizJobLog(&model), nil
}

func (r *JobLogRepo) Update(ctx context.Context, jobLog *biz.JobLog) (*biz.JobLog, error) {
	db := r.data.DB(ctx).WithContext(ctx)
	var model JobLog
	err := db.First(&model, uint64(jobLog.ID)).Error
	if err != nil {
		if isNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	model.Status = jobLog.Status
	model.EndTime = jobLog.EndTime
	model.DurationMS = jobLog.DurationMS
	model.ExitCode = jobLog.ExitCode
	model.LogPath = jobLog.LogPath
	model.LogSizeBytes = jobLog.LogSizeBytes
	model.LogTruncated = jobLog.LogTruncated
	model.ErrorMessage = jobLog.ErrorMessage
	if err := db.Save(&model).Error; err != nil {
		return nil, err
	}
	return toBizJobLog(&model), nil
}

func (r *JobLogRepo) MarkActiveLogsFailedByExecutorID(ctx context.Context, executorID int64, message string) error {
	now := time.Now()
	return r.data.DB(ctx).WithContext(ctx).Model(&JobLog{}).
		Where("executor_id = ? AND status IN ?", executorID, []string{biz.JobLogStatusRunning, biz.JobLogStatusKilling}).
		Updates(map[string]any{
			"status":        biz.JobLogStatusFailed,
			"end_time":      now,
			"error_message": message,
		}).Error
}

func (r *JobLogRepo) MarkAllActiveLogsFailed(ctx context.Context, message string) error {
	now := time.Now()
	return r.data.DB(ctx).WithContext(ctx).Model(&JobLog{}).
		Where("status IN ?", []string{biz.JobLogStatusRunning, biz.JobLogStatusKilling}).
		Updates(map[string]any{
			"status":        biz.JobLogStatusFailed,
			"end_time":      now,
			"error_message": message,
		}).Error
}

func (r *JobLogRepo) MarkKillingTimeoutLogsFailed(ctx context.Context, timeoutSeconds int32, message string) error {
	if timeoutSeconds <= 0 {
		return nil
	}
	now := time.Now()
	deadline := now.Add(-time.Duration(timeoutSeconds) * time.Second)
	return r.data.DB(ctx).WithContext(ctx).Model(&JobLog{}).
		Where("status = ? AND updated_at < ?", biz.JobLogStatusKilling, deadline).
		Updates(map[string]any{
			"status":        biz.JobLogStatusFailed,
			"end_time":      now,
			"error_message": message,
		}).Error
}

func (r *JobLogRepo) DeleteExpiredLogs(ctx context.Context, retentionDays int32) ([]string, error) {
	if retentionDays <= 0 {
		return nil, nil
	}
	deadline := time.Now().AddDate(0, 0, -int(retentionDays))
	var models []*JobLog
	db := r.data.DB(ctx).WithContext(ctx)
	if err := db.Where("created_at < ?", deadline).Find(&models).Error; err != nil {
		return nil, err
	}
	paths := make([]string, 0, len(models))
	ids := make([]uint64, 0, len(models))
	for _, model := range models {
		if model.LogPath != "" {
			paths = append(paths, model.LogPath)
		}
		ids = append(ids, model.ID)
	}
	if len(ids) == 0 {
		return paths, nil
	}
	if err := db.Delete(&JobLog{}, ids).Error; err != nil {
		return nil, err
	}
	return paths, nil
}

func toJobLogModel(jobLog *biz.JobLog) *JobLog {
	if jobLog == nil {
		return nil
	}
	return &JobLog{
		ID:              uint64(jobLog.ID),
		JobID:           uint64(jobLog.JobID),
		JobName:         jobLog.JobName,
		ExecutorID:      uint64(jobLog.ExecutorID),
		ExecutorName:    jobLog.ExecutorName,
		ExecutorAddress: jobLog.ExecutorAddress,
		CronExpr:        jobLog.CronExpr,
		TimeoutSeconds:  jobLog.TimeoutSeconds,
		GlueSnapshot:    jobLog.GlueSnapshot,
		TriggerType:     jobLog.TriggerType,
		Status:          jobLog.Status,
		StartTime:       jobLog.StartTime,
		EndTime:         jobLog.EndTime,
		DurationMS:      jobLog.DurationMS,
		ExitCode:        jobLog.ExitCode,
		LogPath:         jobLog.LogPath,
		LogSizeBytes:    jobLog.LogSizeBytes,
		LogTruncated:    jobLog.LogTruncated,
		ErrorMessage:    jobLog.ErrorMessage,
	}
}

func toBizJobLog(model *JobLog) *biz.JobLog {
	if model == nil {
		return nil
	}
	return &biz.JobLog{
		ID:              int64(model.ID),
		JobID:           int64(model.JobID),
		JobName:         model.JobName,
		ExecutorID:      int64(model.ExecutorID),
		ExecutorName:    model.ExecutorName,
		ExecutorAddress: model.ExecutorAddress,
		CronExpr:        model.CronExpr,
		TimeoutSeconds:  model.TimeoutSeconds,
		GlueSnapshot:    model.GlueSnapshot,
		TriggerType:     model.TriggerType,
		Status:          model.Status,
		StartTime:       model.StartTime,
		EndTime:         model.EndTime,
		DurationMS:      model.DurationMS,
		ExitCode:        model.ExitCode,
		LogPath:         model.LogPath,
		LogSizeBytes:    model.LogSizeBytes,
		LogTruncated:    model.LogTruncated,
		ErrorMessage:    model.ErrorMessage,
		CreatedAt:       model.CreatedAt,
		UpdatedAt:       model.UpdatedAt,
	}
}
