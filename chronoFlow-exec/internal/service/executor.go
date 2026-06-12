package service

import (
	"context"
	"strings"
	"time"

	v1 "chronoFlow-exec/api/all-pb-go/v1"
	"chronoFlow-exec/internal/callback"
	"chronoFlow-exec/internal/conf"
	httpErrors "chronoFlow-exec/internal/errors"
	"chronoFlow-exec/internal/process"
	"chronoFlow-exec/internal/store"
)

type ExecutorService struct {
	v1.UnimplementedExecutorServer

	executorConf *conf.Executor
	manager      *process.Manager
	pendingStore *store.PendingStore
	callback     *callback.Client
}

func NewExecutorService(executorConf *conf.Executor, manager *process.Manager, pendingStore *store.PendingStore, callbackClient *callback.Client) *ExecutorService {
	return &ExecutorService{
		executorConf: executorConf,
		manager:      manager,
		pendingStore: pendingStore,
		callback:     callbackClient,
	}
}

func (s *ExecutorService) Health(ctx context.Context, _ *v1.HealthRequest) (*v1.HealthReply, error) {
	return &v1.HealthReply{
		Code: 0,
		Msg:  "Health success",
		Data: &v1.HealthReply_Data{Status: "online", ExecutorName: s.executorName()},
	}, nil
}

func (s *ExecutorService) Run(ctx context.Context, req *v1.RunRequest) (*v1.RunReply, error) {
	if req == nil {
		return nil, httpErrors.E(httpErrors.ErrInvalidRequestBody)
	}
	if req.GetJobId() <= 0 || req.GetLogId() <= 0 {
		return nil, httpErrors.E(httpErrors.ErrInvalidID)
	}
	if strings.TrimSpace(req.GetScript()) == "" || strings.TrimSpace(req.GetCallbackUrl()) == "" || strings.TrimSpace(req.GetCallbackToken()) == "" {
		return nil, httpErrors.EWithMessage(httpErrors.ErrMissingRequiredField, "script、callback_url 和 callback_token 不能为空")
	}
	if s.manager == nil {
		return nil, httpErrors.EWithMessage(httpErrors.ErrConfigInvalid, "process manager 未初始化")
	}
	// /run 是异步接口，不能沿用 HTTP request context；请求返回后该 context 会被取消。
	err := s.manager.Run(context.Background(), process.RunRequest{
		JobID:          req.GetJobId(),
		LogID:          req.GetLogId(),
		Script:         req.GetScript(),
		TimeoutSeconds: req.GetTimeoutSeconds(),
		CallbackURL:    req.GetCallbackUrl(),
		CallbackToken:  req.GetCallbackToken(),
	}, s.handleResult)
	if err != nil {
		return nil, httpErrors.EWithMessage(httpErrors.ErrConflict, "任务正在执行中")
	}
	return &v1.RunReply{Code: 0, Msg: "Run accepted", Data: &v1.RunReply_Data{Status: "accepted", LogId: req.GetLogId()}}, nil
}

func (s *ExecutorService) Kill(ctx context.Context, req *v1.KillRequest) (*v1.KillReply, error) {
	if req == nil {
		return nil, httpErrors.E(httpErrors.ErrInvalidRequestBody)
	}
	if req.GetJobId() <= 0 || req.GetLogId() <= 0 {
		return nil, httpErrors.E(httpErrors.ErrInvalidID)
	}
	if s.manager == nil {
		return nil, httpErrors.EWithMessage(httpErrors.ErrConfigInvalid, "process manager 未初始化")
	}
	if err := s.manager.Kill(req.GetJobId(), req.GetLogId()); err != nil {
		return nil, httpErrors.EWithMessage(httpErrors.ErrConflict, "任务未在执行中")
	}
	return &v1.KillReply{Code: 0, Msg: "Kill accepted", Data: &v1.KillReply_Data{Status: "killing", LogId: req.GetLogId()}}, nil
}

func (s *ExecutorService) handleResult(result *process.Result) {
	item := &store.CallbackItem{
		LogID:         result.LogID,
		JobID:         result.JobID,
		CallbackURL:   result.CallbackURL,
		CallbackToken: result.CallbackToken,
		Status:        result.Status,
		ExitCode:      result.ExitCode,
		LogContent:    result.LogContent,
		LogTruncated:  result.LogTruncated,
		StartTime:     result.StartTime,
		EndTime:       result.EndTime,
		DurationMS:    result.DurationMS,
		ErrorMessage:  result.ErrorMessage,
		CreatedAt:     time.Now(),
	}
	if s.pendingStore == nil {
		return
	}
	if err := s.pendingStore.Save(item); err != nil {
		return
	}
	if s.callback == nil {
		return
	}
	if err := s.callback.Send(item); err == nil {
		_ = s.pendingStore.DeletePending(item.LogID)
	}
}

func (s *ExecutorService) executorName() string {
	if s.executorConf != nil && s.executorConf.Name != "" {
		return s.executorConf.Name
	}
	return "executor"
}
