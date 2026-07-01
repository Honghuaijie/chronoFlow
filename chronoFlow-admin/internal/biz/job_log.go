package biz

import (
	"context"
	"errors"
	"os"
	"time"

	httpErrors "chronoFlow-admin/internal/errors"

	"github.com/go-kratos/kratos/v2/log"
)

const MissingLogContentMessage = "日志文件不存在或已被清理"

type JobLog struct {
	ID                   int64
	JobID                int64
	JobName              string
	ExecutorID           int64
	ExecutorName         string
	ExecutorAddress      string
	CronExpr             string
	TimeoutSeconds       int32
	GlueSnapshot         string
	TriggerType          string
	Status               string
	StartTime            time.Time
	EndTime              *time.Time
	DurationMS           int64
	ExitCode             *int32
	LogPath              string
	LogSizeBytes         int64
	LogTruncated         bool
	ErrorMessage         string
	AlertEnabledSnapshot bool
	AlertStatus          string
	AlertError           string
	AlertSentAt          *time.Time
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type JobLogFilter struct {
	JobID       int64
	ExecutorID  int64
	Status      string
	TriggerType string
	Page        int32
	PageSize    int32
}

type JobLogDetail struct {
	Log          *JobLog
	GlueSnapshot string
	LogContent   string
}

type JobLogRepo interface {
	List(context.Context, JobLogFilter) ([]*JobLog, int64, error)
	GetByID(context.Context, int64) (*JobLog, error)
}

type LogReader interface {
	Read(context.Context, string) (string, error)
}

type JobLogUsecase struct {
	repo      JobLogRepo
	logReader LogReader
	log       *log.Helper
}

func NewJobLogUsecase(repo JobLogRepo, logReader LogReader, logger log.Logger) *JobLogUsecase {
	return &JobLogUsecase{
		repo:      repo,
		logReader: logReader,
		log:       log.NewHelper(logger),
	}
}

func (uc *JobLogUsecase) ListJobLogs(ctx context.Context, filter JobLogFilter) ([]*JobLog, int64, error) {
	if filter.JobID < 0 || filter.ExecutorID < 0 {
		return nil, 0, httpErrors.E(httpErrors.ErrInvalidID)
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}
	return uc.repo.List(ctx, filter)
}

func (uc *JobLogUsecase) GetJobLogDetail(ctx context.Context, id int64) (*JobLogDetail, error) {
	if id <= 0 {
		return nil, httpErrors.E(httpErrors.ErrInvalidID)
	}
	jobLog, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if jobLog == nil {
		return nil, httpErrors.EWithMessage(httpErrors.ErrNotFound, "执行日志不存在")
	}
	content := ""
	if jobLog.LogPath != "" {
		content, err = uc.logReader.Read(ctx, jobLog.LogPath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				content = MissingLogContentMessage
			} else {
				return nil, err
			}
		}
	}
	return &JobLogDetail{
		Log:          jobLog,
		GlueSnapshot: jobLog.GlueSnapshot,
		LogContent:   content,
	}, nil
}
