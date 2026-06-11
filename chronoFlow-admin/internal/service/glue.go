package service

import (
	"context"

	v1 "chronoFlow-admin/api/all-pb-go/v1"
	"chronoFlow-admin/internal/biz"
	httpErrors "chronoFlow-admin/internal/errors"
)

type GlueService struct {
	v1.UnimplementedGlueServer

	uc *biz.GlueUsecase
}

func NewGlueService(uc *biz.GlueUsecase) *GlueService {
	return &GlueService{uc: uc}
}

func (s *GlueService) GetGlue(ctx context.Context, req *v1.GetGlueRequest) (*v1.GetGlueReply, error) {
	jobID, err := validateInt64ID(req.GetJobId())
	if err != nil {
		return nil, err
	}
	glue, err := s.uc.GetGlue(ctx, jobID)
	if err != nil {
		return nil, err
	}
	return &v1.GetGlueReply{Code: 0, Message: successMessage("GetGlue"), Data: &v1.GetGlueReply_Data{Glue: toGlueInfo(glue)}}, nil
}

func (s *GlueService) SaveGlue(ctx context.Context, req *v1.SaveGlueRequest) (*v1.SaveGlueReply, error) {
	if req == nil {
		return nil, httpErrors.E(httpErrors.ErrInvalidRequestBody)
	}
	glue, err := s.uc.SaveGlue(ctx, req.GetJobId(), req.GetContent())
	if err != nil {
		return nil, err
	}
	return &v1.SaveGlueReply{Code: 0, Message: successMessage("SaveGlue"), Data: &v1.SaveGlueReply_Data{Glue: toGlueInfo(glue)}}, nil
}

func toGlueInfo(glue *biz.Glue) *v1.GlueInfo {
	if glue == nil {
		return nil
	}
	return &v1.GlueInfo{
		Id:        glue.ID,
		JobId:     glue.JobID,
		Content:   glue.Content,
		CreatedAt: formatServiceTime(glue.CreatedAt),
		UpdatedAt: formatServiceTime(glue.UpdatedAt),
	}
}
