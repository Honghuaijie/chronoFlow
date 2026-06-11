package biz

import (
	"context"
	"time"

	httpErrors "chronoFlow-admin/internal/errors"

	"github.com/go-kratos/kratos/v2/log"
)

type CallbackConfig struct {
	MaxLogBytes int64
}

type CallbackInput struct {
	LogID        int64
	JobID        int64
	Status       string
	ExitCode     int32
	LogContent   string
	LogTruncated bool
	EndTime      time.Time
	DurationMS   int64
	ErrorMessage string
}

type LogWriter interface {
	Write(context.Context, int64, int64, string) (string, int64, error)
}

type CallbackUsecase struct {
	logRepo JobRunLogRepo
	store   LogWriter
	config  CallbackConfig
	log     *log.Helper
}

func NewCallbackUsecase(logRepo JobRunLogRepo, store LogWriter, config CallbackConfig, logger log.Logger) *CallbackUsecase {
	return &CallbackUsecase{
		logRepo: logRepo,
		store:   store,
		config:  config,
		log:     log.NewHelper(logger),
	}
}

func (uc *CallbackUsecase) ApplyCallback(ctx context.Context, input *CallbackInput) (*JobRunResult, error) {
	if input == nil {
		return nil, httpErrors.E(httpErrors.ErrInvalidRequestBody)
	}
	if input.LogID <= 0 || input.JobID <= 0 {
		return nil, httpErrors.E(httpErrors.ErrInvalidID)
	}
	if !IsFinalJobLogStatus(input.Status) || input.Status == JobLogStatusSkipped {
		return nil, httpErrors.EWithMessage(httpErrors.ErrInvalidParam, "回调状态无效")
	}
	jobLog, err := uc.logRepo.GetByID(ctx, input.LogID)
	if err != nil {
		return nil, err
	}
	if jobLog == nil {
		return nil, httpErrors.EWithMessage(httpErrors.ErrNotFound, "执行日志不存在")
	}
	if !CanCallbackUpdateJobLogStatus(jobLog.Status) {
		return &JobRunResult{LogID: jobLog.ID, Status: jobLog.Status}, nil
	}
	content, truncated := truncateLogContent(input.LogContent, uc.config.MaxLogBytes)
	logPath, logSize, err := uc.store.Write(ctx, jobLog.ID, jobLog.JobID, content)
	if err != nil {
		return nil, err
	}
	endTime := input.EndTime
	if endTime.IsZero() {
		endTime = time.Now()
	}
	exitCode := input.ExitCode
	jobLog.Status = input.Status
	jobLog.EndTime = &endTime
	jobLog.DurationMS = input.DurationMS
	jobLog.ExitCode = &exitCode
	jobLog.LogPath = logPath
	jobLog.LogSizeBytes = logSize
	jobLog.LogTruncated = input.LogTruncated || truncated
	jobLog.ErrorMessage = input.ErrorMessage
	updated, err := uc.logRepo.Update(ctx, jobLog)
	if err != nil {
		return nil, err
	}
	return &JobRunResult{LogID: updated.ID, Status: updated.Status}, nil
}

func truncateLogContent(content string, maxBytes int64) (string, bool) {
	if maxBytes <= 0 || int64(len(content)) <= maxBytes {
		return content, false
	}
	if maxBytes <= 32 {
		return content[:maxBytes], true
	}
	headSize := int(maxBytes / 2)
	tailSize := int(maxBytes) - headSize
	return content[:headSize] + content[len(content)-tailSize:], true
}
