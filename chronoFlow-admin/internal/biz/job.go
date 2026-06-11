package biz

import (
	"context"
	"strings"
	"time"

	httpErrors "chronoFlow-admin/internal/errors"

	"github.com/go-kratos/kratos/v2/log"
)

type Job struct {
	ID             int64
	ExecutorID     int64
	Name           string
	CronExpr       string
	TimeoutSeconds int32
	ScheduleStatus string
	Description    string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type CreateJobInput struct {
	ExecutorID     int64
	Name           string
	CronExpr       string
	TimeoutSeconds int32
	Description    string
}

type UpdateJobInput struct {
	ID             int64
	ExecutorID     int64
	Name           string
	CronExpr       string
	TimeoutSeconds int32
	Description    string
}

type JobRepo interface {
	Create(context.Context, *Job) (*Job, error)
	GetByID(context.Context, int64) (*Job, error)
	List(context.Context, int64) ([]*Job, error)
	Update(context.Context, *Job) (*Job, error)
	Delete(context.Context, int64) error
}

type JobUsecase struct {
	repo     JobRepo
	glueRepo GlueRepo
	log      *log.Helper
}

func NewJobUsecase(repo JobRepo, glueRepo GlueRepo, logger log.Logger) *JobUsecase {
	return &JobUsecase{
		repo:     repo,
		glueRepo: glueRepo,
		log:      log.NewHelper(logger),
	}
}

func (uc *JobUsecase) CreateJob(ctx context.Context, input *CreateJobInput) (*Job, error) {
	job, err := normalizeJobInput(input.ExecutorID, input.Name, input.CronExpr, input.TimeoutSeconds, input.Description)
	if err != nil {
		return nil, err
	}
	job.ScheduleStatus = ScheduleStatusStopped
	return uc.repo.Create(ctx, job)
}

func (uc *JobUsecase) UpdateJob(ctx context.Context, input *UpdateJobInput) (*Job, error) {
	if input.ID <= 0 {
		return nil, httpErrors.E(httpErrors.ErrInvalidID)
	}
	existing, err := uc.mustGetJob(ctx, input.ID)
	if err != nil {
		return nil, err
	}
	normalized, err := normalizeJobInput(input.ExecutorID, input.Name, input.CronExpr, input.TimeoutSeconds, input.Description)
	if err != nil {
		return nil, err
	}
	existing.ExecutorID = normalized.ExecutorID
	existing.Name = normalized.Name
	existing.CronExpr = normalized.CronExpr
	existing.TimeoutSeconds = normalized.TimeoutSeconds
	existing.Description = normalized.Description
	return uc.repo.Update(ctx, existing)
}

func (uc *JobUsecase) GetJob(ctx context.Context, id int64) (*Job, error) {
	return uc.mustGetJob(ctx, id)
}

func (uc *JobUsecase) ListJobs(ctx context.Context, executorID int64) ([]*Job, error) {
	if executorID < 0 {
		return nil, httpErrors.E(httpErrors.ErrInvalidID)
	}
	return uc.repo.List(ctx, executorID)
}

func (uc *JobUsecase) DeleteJob(ctx context.Context, id int64) error {
	if id <= 0 {
		return httpErrors.E(httpErrors.ErrInvalidID)
	}
	return uc.repo.Delete(ctx, id)
}

func (uc *JobUsecase) StartJob(ctx context.Context, id int64) (*Job, error) {
	job, err := uc.mustGetJob(ctx, id)
	if err != nil {
		return nil, err
	}
	glue, err := uc.glueRepo.GetByJobID(ctx, id)
	if err != nil {
		return nil, err
	}
	if glue == nil || strings.TrimSpace(glue.Content) == "" {
		return nil, httpErrors.EWithMessage(httpErrors.ErrConflict, "任务 Glue 不能为空")
	}
	job.ScheduleStatus = ScheduleStatusRunning
	return uc.repo.Update(ctx, job)
}

func (uc *JobUsecase) StopJob(ctx context.Context, id int64) (*Job, error) {
	job, err := uc.mustGetJob(ctx, id)
	if err != nil {
		return nil, err
	}
	job.ScheduleStatus = ScheduleStatusStopped
	return uc.repo.Update(ctx, job)
}

func (uc *JobUsecase) mustGetJob(ctx context.Context, id int64) (*Job, error) {
	if id <= 0 {
		return nil, httpErrors.E(httpErrors.ErrInvalidID)
	}
	job, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if job == nil {
		return nil, httpErrors.EWithMessage(httpErrors.ErrNotFound, "任务不存在")
	}
	return job, nil
}

func normalizeJobInput(executorID int64, name string, cronExpr string, timeoutSeconds int32, description string) (*Job, error) {
	if executorID <= 0 {
		return nil, httpErrors.EWithMessage(httpErrors.ErrInvalidID, "executor_id 无效")
	}
	name = strings.TrimSpace(name)
	cronExpr = strings.Join(strings.Fields(cronExpr), " ")
	if name == "" || cronExpr == "" {
		return nil, httpErrors.EWithMessage(httpErrors.ErrMissingRequiredField, "name 和 cron_expr 不能为空")
	}
	if !isSixFieldCron(cronExpr) {
		return nil, httpErrors.EWithMessage(httpErrors.ErrInvalidParam, "cron_expr 必须是 6 位 Cron 表达式")
	}
	if timeoutSeconds <= 0 {
		return nil, httpErrors.EWithMessage(httpErrors.ErrInvalidParam, "timeout_seconds 必须大于 0")
	}
	return &Job{
		ExecutorID:     executorID,
		Name:           name,
		CronExpr:       cronExpr,
		TimeoutSeconds: timeoutSeconds,
		Description:    strings.TrimSpace(description),
	}, nil
}

func isSixFieldCron(expr string) bool {
	return len(strings.Fields(expr)) == 6
}
