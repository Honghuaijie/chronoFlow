package biz

import (
	"context"
	"strings"
	"time"

	httpErrors "chronoFlow-admin/internal/errors"

	"github.com/go-kratos/kratos/v2/log"
)

type Glue struct {
	ID        int64
	JobID     int64
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type GlueRepo interface {
	GetByJobID(context.Context, int64) (*Glue, error)
	Save(context.Context, *Glue) (*Glue, error)
}

type GlueUsecase struct {
	repo GlueRepo
	log  *log.Helper
}

func NewGlueUsecase(repo GlueRepo, logger log.Logger) *GlueUsecase {
	return &GlueUsecase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

func (uc *GlueUsecase) GetGlue(ctx context.Context, jobID int64) (*Glue, error) {
	if jobID <= 0 {
		return nil, httpErrors.E(httpErrors.ErrInvalidID)
	}
	return uc.repo.GetByJobID(ctx, jobID)
}

func (uc *GlueUsecase) SaveGlue(ctx context.Context, jobID int64, content string) (*Glue, error) {
	if jobID <= 0 {
		return nil, httpErrors.E(httpErrors.ErrInvalidID)
	}
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, httpErrors.EWithMessage(httpErrors.ErrMissingRequiredField, "content 不能为空")
	}
	return uc.repo.Save(ctx, &Glue{
		JobID:   jobID,
		Content: content,
	})
}
