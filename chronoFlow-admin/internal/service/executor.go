package service

import (
	"context"
	"strings"
	"time"

	v1 "chronoFlow-admin/api/all-pb-go/v1"
	"chronoFlow-admin/internal/biz"
	httpErrors "chronoFlow-admin/internal/errors"
)

type ExecutorService struct {
	v1.UnimplementedExecutorServer

	uc *biz.ExecutorUsecase
}

func NewExecutorService(uc *biz.ExecutorUsecase) *ExecutorService {
	return &ExecutorService{uc: uc}
}

func (s *ExecutorService) CreateExecutor(ctx context.Context, req *v1.CreateExecutorRequest) (*v1.CreateExecutorReply, error) {
	input, err := validateCreateExecutorRequest(req)
	if err != nil {
		return nil, err
	}
	executor, err := s.uc.CreateExecutor(ctx, input)
	if err != nil {
		return nil, err
	}
	return &v1.CreateExecutorReply{
		Code:    0,
		Message: successMessage("CreateExecutor"),
		Data:    &v1.CreateExecutorReply_Data{Executor: toExecutorInfo(executor)},
	}, nil
}

func (s *ExecutorService) UpdateExecutor(ctx context.Context, req *v1.UpdateExecutorRequest) (*v1.UpdateExecutorReply, error) {
	input, err := validateUpdateExecutorRequest(req)
	if err != nil {
		return nil, err
	}
	executor, err := s.uc.UpdateExecutor(ctx, input)
	if err != nil {
		return nil, err
	}
	return &v1.UpdateExecutorReply{
		Code:    0,
		Message: successMessage("UpdateExecutor"),
		Data:    &v1.UpdateExecutorReply_Data{Executor: toExecutorInfo(executor)},
	}, nil
}

func (s *ExecutorService) DeleteExecutor(ctx context.Context, req *v1.DeleteExecutorRequest) (*v1.DeleteExecutorReply, error) {
	id, err := validateInt64ID(req.GetId())
	if err != nil {
		return nil, err
	}
	if err := s.uc.DeleteExecutor(ctx, id); err != nil {
		return nil, err
	}
	return &v1.DeleteExecutorReply{
		Code:    0,
		Message: successMessage("DeleteExecutor"),
		Data:    &v1.DeleteExecutorReply_Data{Id: id},
	}, nil
}

func (s *ExecutorService) GetExecutor(ctx context.Context, req *v1.GetExecutorRequest) (*v1.GetExecutorReply, error) {
	id, err := validateInt64ID(req.GetId())
	if err != nil {
		return nil, err
	}
	executor, err := s.uc.GetExecutor(ctx, id)
	if err != nil {
		return nil, err
	}
	return &v1.GetExecutorReply{
		Code:    0,
		Message: successMessage("GetExecutor"),
		Data:    &v1.GetExecutorReply_Data{Executor: toExecutorInfo(executor)},
	}, nil
}

func (s *ExecutorService) ListExecutors(ctx context.Context, _ *v1.ListExecutorsRequest) (*v1.ListExecutorsReply, error) {
	executors, err := s.uc.ListExecutors(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]*v1.ExecutorInfo, 0, len(executors))
	for _, executor := range executors {
		items = append(items, toExecutorInfo(executor))
	}
	return &v1.ListExecutorsReply{
		Code:    0,
		Message: successMessage("ListExecutors"),
		Data: &v1.ListExecutorsReply_Data{
			Items: items,
			Total: int32(len(items)),
		},
	}, nil
}

func validateCreateExecutorRequest(req *v1.CreateExecutorRequest) (*biz.CreateExecutorInput, error) {
	if req == nil {
		return nil, httpErrors.E(httpErrors.ErrInvalidRequestBody)
	}
	name := strings.TrimSpace(req.GetName())
	address := strings.TrimSpace(req.GetAddress())
	token := strings.TrimSpace(req.GetToken())
	if name == "" || address == "" || token == "" {
		return nil, httpErrors.EWithMessage(httpErrors.ErrMissingRequiredField, "name、address 和 token 不能为空")
	}
	return &biz.CreateExecutorInput{
		Name:        name,
		Address:     address,
		Token:       token,
		Description: strings.TrimSpace(req.GetDescription()),
	}, nil
}

func validateUpdateExecutorRequest(req *v1.UpdateExecutorRequest) (*biz.UpdateExecutorInput, error) {
	if req == nil {
		return nil, httpErrors.E(httpErrors.ErrInvalidRequestBody)
	}
	id, err := validateInt64ID(req.GetId())
	if err != nil {
		return nil, err
	}
	name := strings.TrimSpace(req.GetName())
	address := strings.TrimSpace(req.GetAddress())
	if name == "" || address == "" {
		return nil, httpErrors.EWithMessage(httpErrors.ErrMissingRequiredField, "name 和 address 不能为空")
	}
	return &biz.UpdateExecutorInput{
		ID:          id,
		Name:        name,
		Address:     address,
		Token:       strings.TrimSpace(req.GetToken()),
		Description: strings.TrimSpace(req.GetDescription()),
	}, nil
}

func validateInt64ID(id int64) (int64, error) {
	if id <= 0 {
		return 0, httpErrors.E(httpErrors.ErrInvalidID)
	}
	return id, nil
}

func toExecutorInfo(executor *biz.Executor) *v1.ExecutorInfo {
	if executor == nil {
		return nil
	}
	lastHeartbeatTime := ""
	if executor.LastHeartbeatTime != nil {
		lastHeartbeatTime = formatServiceTime(*executor.LastHeartbeatTime)
	}
	return &v1.ExecutorInfo{
		Id:                 executor.ID,
		Name:               executor.Name,
		Address:            executor.Address,
		Description:        executor.Description,
		Status:             executor.Status,
		HeartbeatFailCount: executor.HeartbeatFailCount,
		LastHeartbeatTime:  lastHeartbeatTime,
		CreatedAt:          formatServiceTime(executor.CreatedAt),
		UpdatedAt:          formatServiceTime(executor.UpdatedAt),
	}
}

func formatServiceTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.In(time.Local).Format(time.RFC3339)
}
