package service

import (
	"context"
	"strings"
	"time"

	v1 "chronoFlow-admin/api/all-pb-go/v1"
	"chronoFlow-admin/internal/biz"
	"chronoFlow-admin/internal/conf"
	httpErrors "chronoFlow-admin/internal/errors"

	httpCtx "github.com/go-kratos/kratos/v2/transport/http"
)

const callbackTokenHeader = "X-Callback-Token"

type CallbackService struct {
	v1.UnimplementedJobRunCallbackServer

	uc            *biz.CallbackUsecase
	callbackToken string
}

func NewCallbackService(uc *biz.CallbackUsecase, security *conf.Security) *CallbackService {
	token := ""
	if security != nil {
		token = security.CallbackToken
	}
	return &CallbackService{uc: uc, callbackToken: token}
}

func (s *CallbackService) CallbackJobRun(ctx context.Context, req *v1.CallbackJobRunRequest) (*v1.CallbackJobRunReply, error) {
	if err := s.validateCallbackToken(ctx); err != nil {
		return nil, err
	}
	input, err := validateCallbackJobRunRequest(req)
	if err != nil {
		return nil, err
	}
	result, err := s.uc.ApplyCallback(ctx, input)
	if err != nil {
		return nil, err
	}
	return &v1.CallbackJobRunReply{
		Code:    0,
		Message: successMessage("CallbackJobRun"),
		Data:    &v1.CallbackJobRunReply_Data{LogId: result.LogID, Status: result.Status},
	}, nil
}

func (s *CallbackService) validateCallbackToken(ctx context.Context) error {
	tr, ok := httpCtx.RequestFromServerContext(ctx)
	if !ok {
		return httpErrors.E(httpErrors.ErrInvalidToken)
	}
	if strings.TrimSpace(tr.Header.Get(callbackTokenHeader)) != s.callbackToken {
		return httpErrors.E(httpErrors.ErrInvalidToken)
	}
	return nil
}

func validateCallbackJobRunRequest(req *v1.CallbackJobRunRequest) (*biz.CallbackInput, error) {
	if req == nil {
		return nil, httpErrors.E(httpErrors.ErrInvalidRequestBody)
	}
	endTime, err := parseCallbackTime(req.GetEndTime())
	if err != nil {
		return nil, err
	}
	return &biz.CallbackInput{
		LogID:        req.GetLogId(),
		JobID:        req.GetJobId(),
		Status:       strings.TrimSpace(req.GetStatus()),
		ExitCode:     req.GetExitCode(),
		LogContent:   req.GetLogContent(),
		LogTruncated: req.GetLogTruncated(),
		EndTime:      endTime,
		DurationMS:   req.GetDurationMs(),
		ErrorMessage: strings.TrimSpace(req.GetErrorMessage()),
	}, nil
}

func parseCallbackTime(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, nil
	}
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t, nil
	}
	t, err := time.ParseInLocation("2006-01-02 15:04:05", value, time.Local)
	if err != nil {
		return time.Time{}, httpErrors.EWithMessage(httpErrors.ErrInvalidParam, "end_time 格式错误")
	}
	return t, nil
}
