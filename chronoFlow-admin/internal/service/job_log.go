package service

import (
	"context"

	v1 "chronoFlow-admin/api/all-pb-go/v1"
	"chronoFlow-admin/internal/biz"
)

type JobLogService struct {
	v1.UnimplementedJobLogServer

	uc *biz.JobLogUsecase
}

func NewJobLogService(uc *biz.JobLogUsecase) *JobLogService {
	return &JobLogService{uc: uc}
}

func (s *JobLogService) ListJobLogs(ctx context.Context, req *v1.ListJobLogsRequest) (*v1.ListJobLogsReply, error) {
	items, total, err := s.uc.ListJobLogs(ctx, biz.JobLogFilter{
		JobID:       req.GetJobId(),
		ExecutorID:  req.GetExecutorId(),
		Status:      req.GetStatus(),
		TriggerType: req.GetTriggerType(),
		Page:        req.GetPage(),
		PageSize:    req.GetPageSize(),
	})
	if err != nil {
		return nil, err
	}
	respItems := make([]*v1.JobLogInfo, 0, len(items))
	for _, item := range items {
		respItems = append(respItems, toJobLogInfo(item))
	}
	return &v1.ListJobLogsReply{
		Code:    0,
		Message: successMessage("ListJobLogs"),
		Data:    &v1.ListJobLogsReply_Data{Items: respItems, Total: int32(total)},
	}, nil
}

func (s *JobLogService) GetJobLogDetail(ctx context.Context, req *v1.GetJobLogDetailRequest) (*v1.GetJobLogDetailReply, error) {
	id, err := validateInt64ID(req.GetId())
	if err != nil {
		return nil, err
	}
	detail, err := s.uc.GetJobLogDetail(ctx, id)
	if err != nil {
		return nil, err
	}
	return &v1.GetJobLogDetailReply{
		Code:    0,
		Message: successMessage("GetJobLogDetail"),
		Data: &v1.GetJobLogDetailReply_Data{
			Log:          toJobLogInfo(detail.Log),
			GlueSnapshot: detail.GlueSnapshot,
			LogContent:   detail.LogContent,
		},
	}, nil
}

func toJobLogInfo(jobLog *biz.JobLog) *v1.JobLogInfo {
	if jobLog == nil {
		return nil
	}
	endTime := ""
	if jobLog.EndTime != nil {
		endTime = formatServiceTime(*jobLog.EndTime)
	}
	exitCode := int32(0)
	if jobLog.ExitCode != nil {
		exitCode = *jobLog.ExitCode
	}
	alertSentAt := ""
	if jobLog.AlertSentAt != nil {
		alertSentAt = formatServiceTime(*jobLog.AlertSentAt)
	}
	return &v1.JobLogInfo{
		Id:                   jobLog.ID,
		JobId:                jobLog.JobID,
		JobName:              jobLog.JobName,
		ExecutorId:           jobLog.ExecutorID,
		ExecutorName:         jobLog.ExecutorName,
		ExecutorAddress:      jobLog.ExecutorAddress,
		CronExpr:             jobLog.CronExpr,
		TimeoutSeconds:       jobLog.TimeoutSeconds,
		TriggerType:          jobLog.TriggerType,
		Status:               jobLog.Status,
		StartTime:            formatServiceTime(jobLog.StartTime),
		EndTime:              endTime,
		DurationMs:           jobLog.DurationMS,
		ExitCode:             exitCode,
		LogPath:              jobLog.LogPath,
		LogSizeBytes:         jobLog.LogSizeBytes,
		LogTruncated:         jobLog.LogTruncated,
		ErrorMessage:         jobLog.ErrorMessage,
		CreatedAt:            formatServiceTime(jobLog.CreatedAt),
		UpdatedAt:            formatServiceTime(jobLog.UpdatedAt),
		AlertEnabledSnapshot: jobLog.AlertEnabledSnapshot,
		AlertStatus:          jobLog.AlertStatus,
		AlertError:           jobLog.AlertError,
		AlertSentAt:          alertSentAt,
	}
}
