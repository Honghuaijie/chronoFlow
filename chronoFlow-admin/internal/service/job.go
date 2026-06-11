package service

import (
	"context"
	"strings"

	v1 "chronoFlow-admin/api/all-pb-go/v1"
	"chronoFlow-admin/internal/biz"
	httpErrors "chronoFlow-admin/internal/errors"
	"chronoFlow-admin/internal/scheduler"
)

type JobService struct {
	v1.UnimplementedJobServer

	uc    *biz.JobUsecase
	runUC *biz.JobRunUsecase
	sched *scheduler.Manager
}

func NewJobService(uc *biz.JobUsecase, runUC *biz.JobRunUsecase, sched *scheduler.Manager) *JobService {
	return &JobService{uc: uc, runUC: runUC, sched: sched}
}

func (s *JobService) CreateJob(ctx context.Context, req *v1.CreateJobRequest) (*v1.CreateJobReply, error) {
	input, err := validateCreateJobRequest(req)
	if err != nil {
		return nil, err
	}
	job, err := s.uc.CreateJob(ctx, input)
	if err != nil {
		return nil, err
	}
	return &v1.CreateJobReply{Code: 0, Message: successMessage("CreateJob"), Data: &v1.CreateJobReply_Data{Job: toJobInfo(job)}}, nil
}

func (s *JobService) UpdateJob(ctx context.Context, req *v1.UpdateJobRequest) (*v1.UpdateJobReply, error) {
	input, err := validateUpdateJobRequest(req)
	if err != nil {
		return nil, err
	}
	job, err := s.uc.UpdateJob(ctx, input)
	if err != nil {
		return nil, err
	}
	return &v1.UpdateJobReply{Code: 0, Message: successMessage("UpdateJob"), Data: &v1.UpdateJobReply_Data{Job: toJobInfo(job)}}, nil
}

func (s *JobService) DeleteJob(ctx context.Context, req *v1.DeleteJobRequest) (*v1.DeleteJobReply, error) {
	id, err := validateInt64ID(req.GetId())
	if err != nil {
		return nil, err
	}
	if err := s.uc.DeleteJob(ctx, id); err != nil {
		return nil, err
	}
	return &v1.DeleteJobReply{Code: 0, Message: successMessage("DeleteJob"), Data: &v1.DeleteJobReply_Data{Id: id}}, nil
}

func (s *JobService) GetJob(ctx context.Context, req *v1.GetJobRequest) (*v1.GetJobReply, error) {
	id, err := validateInt64ID(req.GetId())
	if err != nil {
		return nil, err
	}
	job, err := s.uc.GetJob(ctx, id)
	if err != nil {
		return nil, err
	}
	return &v1.GetJobReply{Code: 0, Message: successMessage("GetJob"), Data: &v1.GetJobReply_Data{Job: toJobInfo(job)}}, nil
}

func (s *JobService) ListJobs(ctx context.Context, req *v1.ListJobsRequest) (*v1.ListJobsReply, error) {
	executorID := req.GetExecutorId()
	if executorID < 0 {
		return nil, httpErrors.E(httpErrors.ErrInvalidID)
	}
	jobs, err := s.uc.ListJobs(ctx, executorID)
	if err != nil {
		return nil, err
	}
	items := make([]*v1.JobInfo, 0, len(jobs))
	for _, job := range jobs {
		items = append(items, toJobInfo(job))
	}
	return &v1.ListJobsReply{
		Code:    0,
		Message: successMessage("ListJobs"),
		Data:    &v1.ListJobsReply_Data{Items: items, Total: int32(len(items))},
	}, nil
}

func (s *JobService) StartJob(ctx context.Context, req *v1.StartJobRequest) (*v1.StartJobReply, error) {
	id, err := validateInt64ID(req.GetId())
	if err != nil {
		return nil, err
	}
	job, err := s.uc.StartJob(ctx, id)
	if err != nil {
		return nil, err
	}
	if s.sched != nil && s.runUC != nil {
		if err := s.sched.Register(job.ID, job.CronExpr, func(runCtx context.Context) error {
			_, err := s.runUC.RunJob(runCtx, job.ID, biz.TriggerTypeCron)
			return err
		}); err != nil {
			return nil, err
		}
	}
	return &v1.StartJobReply{Code: 0, Message: successMessage("StartJob"), Data: &v1.StartJobReply_Data{Job: toJobInfo(job)}}, nil
}

func (s *JobService) StopJob(ctx context.Context, req *v1.StopJobRequest) (*v1.StopJobReply, error) {
	id, err := validateInt64ID(req.GetId())
	if err != nil {
		return nil, err
	}
	job, err := s.uc.StopJob(ctx, id)
	if err != nil {
		return nil, err
	}
	if s.sched != nil {
		s.sched.Remove(id)
	}
	return &v1.StopJobReply{Code: 0, Message: successMessage("StopJob"), Data: &v1.StopJobReply_Data{Job: toJobInfo(job)}}, nil
}

func (s *JobService) RunJob(ctx context.Context, req *v1.RunJobRequest) (*v1.RunJobReply, error) {
	id, err := validateInt64ID(req.GetId())
	if err != nil {
		return nil, err
	}
	result, err := s.runUC.RunJob(ctx, id, biz.TriggerTypeManual)
	if err != nil {
		return nil, err
	}
	return &v1.RunJobReply{
		Code:    0,
		Message: successMessage("RunJob"),
		Data:    &v1.RunJobReply_Data{LogId: result.LogID, Status: result.Status},
	}, nil
}

func (s *JobService) KillJob(ctx context.Context, req *v1.KillJobRequest) (*v1.KillJobReply, error) {
	id, err := validateInt64ID(req.GetId())
	if err != nil {
		return nil, err
	}
	result, err := s.runUC.KillJob(ctx, id)
	if err != nil {
		return nil, err
	}
	return &v1.KillJobReply{
		Code:    0,
		Message: successMessage("KillJob"),
		Data:    &v1.KillJobReply_Data{LogId: result.LogID, Status: result.Status},
	}, nil
}

func validateCreateJobRequest(req *v1.CreateJobRequest) (*biz.CreateJobInput, error) {
	if req == nil {
		return nil, httpErrors.E(httpErrors.ErrInvalidRequestBody)
	}
	return &biz.CreateJobInput{
		ExecutorID:     req.GetExecutorId(),
		Name:           strings.TrimSpace(req.GetName()),
		CronExpr:       strings.TrimSpace(req.GetCronExpr()),
		TimeoutSeconds: req.GetTimeoutSeconds(),
		Description:    strings.TrimSpace(req.GetDescription()),
	}, nil
}

func validateUpdateJobRequest(req *v1.UpdateJobRequest) (*biz.UpdateJobInput, error) {
	if req == nil {
		return nil, httpErrors.E(httpErrors.ErrInvalidRequestBody)
	}
	return &biz.UpdateJobInput{
		ID:             req.GetId(),
		ExecutorID:     req.GetExecutorId(),
		Name:           strings.TrimSpace(req.GetName()),
		CronExpr:       strings.TrimSpace(req.GetCronExpr()),
		TimeoutSeconds: req.GetTimeoutSeconds(),
		Description:    strings.TrimSpace(req.GetDescription()),
	}, nil
}

func toJobInfo(job *biz.Job) *v1.JobInfo {
	if job == nil {
		return nil
	}
	return &v1.JobInfo{
		Id:             job.ID,
		ExecutorId:     job.ExecutorID,
		Name:           job.Name,
		CronExpr:       job.CronExpr,
		TimeoutSeconds: job.TimeoutSeconds,
		ScheduleStatus: job.ScheduleStatus,
		Description:    job.Description,
		CreatedAt:      formatServiceTime(job.CreatedAt),
		UpdatedAt:      formatServiceTime(job.UpdatedAt),
	}
}
