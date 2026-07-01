package biz

import (
	"context"
	"strings"
	"time"

	httpErrors "chronoFlow-admin/internal/errors"

	"github.com/go-kratos/kratos/v2/log"
)

type JobRunConfig struct {
	PublicBaseURL string
	CallbackToken string
}

type JobRunResult struct {
	LogID  int64
	Status string
}

type ExecutorRunRequest struct {
	JobID          int64
	LogID          int64
	Script         string
	TimeoutSeconds int32
	CallbackURL    string
	CallbackToken  string
}

type ExecutorKillRequest struct {
	JobID int64
	LogID int64
}

type ExecutorRunner interface {
	Run(context.Context, string, string, ExecutorRunRequest) error
	Kill(context.Context, string, string, ExecutorKillRequest) error
}

type JobRunLogRepo interface {
	GetRunningByJobID(context.Context, int64) (*JobLog, error)
	Create(context.Context, *JobLog) (*JobLog, error)
	CreateRunningIfNoActive(context.Context, *JobLog) (*JobLog, error)
	GetByID(context.Context, int64) (*JobLog, error)
	Update(context.Context, *JobLog) (*JobLog, error)
}

type JobRunUsecase struct {
	jobRepo      JobLookupRepo
	glueRepo     GlueRepo
	executorRepo ExecutorLookupRepo
	logRepo      JobRunLogRepo
	cipher       TokenCipher
	runner       ExecutorRunner
	config       JobRunConfig
	log          *log.Helper
}

type JobLookupRepo interface {
	GetByID(context.Context, int64) (*Job, error)
}

type ExecutorLookupRepo interface {
	GetByID(context.Context, int64) (*Executor, error)
}

func NewJobRunUsecase(
	jobRepo JobLookupRepo,
	glueRepo GlueRepo,
	executorRepo ExecutorLookupRepo,
	logRepo JobRunLogRepo,
	cipher TokenCipher,
	runner ExecutorRunner,
	config JobRunConfig,
	logger log.Logger,
) *JobRunUsecase {
	return &JobRunUsecase{
		jobRepo:      jobRepo,
		glueRepo:     glueRepo,
		executorRepo: executorRepo,
		logRepo:      logRepo,
		cipher:       cipher,
		runner:       runner,
		config:       config,
		log:          log.NewHelper(logger),
	}
}

func (uc *JobRunUsecase) RunJob(ctx context.Context, jobID int64, triggerType string) (*JobRunResult, error) {
	if jobID <= 0 {
		return nil, httpErrors.E(httpErrors.ErrInvalidID)
	}
	if triggerType == "" {
		triggerType = TriggerTypeManual
	}
	running, err := uc.logRepo.GetRunningByJobID(ctx, jobID)
	if err != nil {
		return nil, err
	}
	if running != nil {
		return nil, httpErrors.EWithMessage(httpErrors.ErrConflict, "任务正在执行中")
	}
	job, err := uc.mustGetRunJob(ctx, jobID)
	if err != nil {
		return nil, err
	}
	glue, err := uc.glueRepo.GetByJobID(ctx, jobID)
	if err != nil {
		return nil, err
	}
	if glue == nil || strings.TrimSpace(glue.Content) == "" {
		return nil, httpErrors.EWithMessage(httpErrors.ErrConflict, "任务 Glue 不能为空")
	}
	executor, err := uc.mustGetRunExecutor(ctx, job.ExecutorID)
	if err != nil {
		return nil, err
	}
	created, err := uc.logRepo.CreateRunningIfNoActive(ctx, &JobLog{
		JobID:                job.ID,
		JobName:              job.Name,
		ExecutorID:           executor.ID,
		ExecutorName:         executor.Name,
		ExecutorAddress:      executor.Address,
		CronExpr:             job.CronExpr,
		TimeoutSeconds:       job.TimeoutSeconds,
		GlueSnapshot:         glue.Content,
		TriggerType:          triggerType,
		Status:               JobLogStatusRunning,
		StartTime:            time.Now(),
		AlertEnabledSnapshot: job.FailureAlertEnabled,
		AlertStatus:          AlertStatusNone,
	})
	if err != nil {
		return nil, err
	}
	token, err := uc.cipher.Decrypt(executor.TokenCiphertext)
	if err != nil {
		return nil, err
	}
	if err := uc.runner.Run(ctx, executor.Address, token, ExecutorRunRequest{
		JobID:          job.ID,
		LogID:          created.ID,
		Script:         glue.Content,
		TimeoutSeconds: job.TimeoutSeconds,
		CallbackURL:    uc.callbackURL(),
		CallbackToken:  uc.config.CallbackToken,
	}); err != nil {
		uc.markLogFailed(ctx, created, err.Error())
		return nil, err
	}
	return &JobRunResult{LogID: created.ID, Status: created.Status}, nil
}

func (uc *JobRunUsecase) KillJob(ctx context.Context, jobID int64) (*JobRunResult, error) {
	if jobID <= 0 {
		return nil, httpErrors.E(httpErrors.ErrInvalidID)
	}
	jobLog, err := uc.logRepo.GetRunningByJobID(ctx, jobID)
	if err != nil {
		return nil, err
	}
	if jobLog == nil {
		return nil, httpErrors.EWithMessage(httpErrors.ErrConflict, "任务未在执行中")
	}
	executor, err := uc.mustGetRunExecutor(ctx, jobLog.ExecutorID)
	if err != nil {
		return nil, err
	}
	token, err := uc.cipher.Decrypt(executor.TokenCiphertext)
	if err != nil {
		return nil, err
	}
	jobLog.Status = JobLogStatusKilling
	updated, err := uc.logRepo.Update(ctx, jobLog)
	if err != nil {
		return nil, err
	}
	if err := uc.runner.Kill(ctx, executor.Address, token, ExecutorKillRequest{JobID: jobID, LogID: jobLog.ID}); err != nil {
		uc.markLogFailed(ctx, updated, err.Error())
		return nil, err
	}
	return &JobRunResult{LogID: updated.ID, Status: updated.Status}, nil
}

func (uc *JobRunUsecase) markLogFailed(ctx context.Context, jobLog *JobLog, message string) {
	if jobLog == nil {
		return
	}
	now := time.Now()
	jobLog.Status = JobLogStatusFailed
	jobLog.EndTime = &now
	jobLog.ErrorMessage = message
	_, _ = uc.logRepo.Update(ctx, jobLog)
}

func (uc *JobRunUsecase) mustGetRunJob(ctx context.Context, id int64) (*Job, error) {
	job, err := uc.jobRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if job == nil {
		return nil, httpErrors.EWithMessage(httpErrors.ErrNotFound, "任务不存在")
	}
	return job, nil
}

func (uc *JobRunUsecase) mustGetRunExecutor(ctx context.Context, id int64) (*Executor, error) {
	executor, err := uc.executorRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if executor == nil {
		return nil, httpErrors.EWithMessage(httpErrors.ErrNotFound, "执行器不存在")
	}
	return executor, nil
}

func (uc *JobRunUsecase) callbackURL() string {
	return strings.TrimRight(uc.config.PublicBaseURL, "/") + "/internal/job-runs/callback"
}
