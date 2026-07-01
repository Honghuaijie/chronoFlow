package biz

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

const (
	defaultAlertMaxAttempts = 3
	defaultAlertRetryDelay  = 2 * time.Second
)

type AlertJobLogRepo interface {
	GetByID(context.Context, int64) (*JobLog, error)
	MarkAlertPending(context.Context, int64) error
	MarkAlertSent(context.Context, int64, time.Time) error
	MarkAlertFailed(context.Context, int64, string) error
	MarkAlertSkipped(context.Context, int64, string) error
	MarkPendingAlertsFailed(context.Context, string) error
}

type AlertSettingsProvider interface {
	GetFeishuWebhook(context.Context) (string, bool, error)
}

type AlertCardSender interface {
	SendCard(context.Context, string, any) error
}

type AlertUsecase struct {
	logRepo     AlertJobLogRepo
	settings    AlertSettingsProvider
	sender      AlertCardSender
	maxAttempts int
	retryDelay  time.Duration
	log         *log.Helper
}

func NewAlertUsecase(logRepo AlertJobLogRepo, settings *SystemSettingUsecase, sender *FeishuAlertSender, logger log.Logger) *AlertUsecase {
	return newAlertUsecase(logRepo, settings, sender, defaultAlertMaxAttempts, defaultAlertRetryDelay, logger)
}

func newAlertUsecase(logRepo AlertJobLogRepo, settings AlertSettingsProvider, sender AlertCardSender, maxAttempts int, retryDelay time.Duration, logger log.Logger) *AlertUsecase {
	if maxAttempts <= 0 {
		maxAttempts = defaultAlertMaxAttempts
	}
	return &AlertUsecase{
		logRepo:     logRepo,
		settings:    settings,
		sender:      sender,
		maxAttempts: maxAttempts,
		retryDelay:  retryDelay,
		log:         log.NewHelper(logger),
	}
}

func (uc *AlertUsecase) DispatchJobLogAlert(_ context.Context, logID int64) {
	go func() {
		if err := uc.sendJobLogAlert(context.Background(), logID); err != nil {
			uc.log.Warnf("dispatch job log alert failed: log_id=%d err=%v", logID, err)
		}
	}()
}

func (uc *AlertUsecase) SendTestFeishuAlert(ctx context.Context) error {
	webhook, configured, err := uc.settings.GetFeishuWebhook(ctx)
	if err != nil {
		return err
	}
	if !configured {
		return fmt.Errorf("飞书 Webhook 未配置")
	}
	now := time.Now()
	payload := buildFeishuAlertCard(FeishuAlertCardInput{
		JobLogID:     0,
		JobID:        0,
		JobName:      "ChronoFlow 测试告警",
		ExecutorName: "-",
		Status:       JobLogStatusFailed,
		TriggerType:  "test",
		StartTime:    now,
		EndTime:      &now,
		ErrorMessage: "这是一条测试告警，如果你收到这条消息，说明飞书 Webhook 配置可用。",
	})
	return uc.sendWithRetry(ctx, webhook, payload)
}

func (uc *AlertUsecase) MarkPendingAlertsFailedOnStartup(ctx context.Context) error {
	return uc.logRepo.MarkPendingAlertsFailed(ctx, "Admin 重启，未完成的告警发送状态未知")
}

func (uc *AlertUsecase) sendJobLogAlert(ctx context.Context, logID int64) error {
	jobLog, err := uc.logRepo.GetByID(ctx, logID)
	if err != nil {
		return err
	}
	if jobLog == nil {
		return nil
	}
	if !jobLog.AlertEnabledSnapshot {
		return uc.logRepo.MarkAlertSkipped(ctx, logID, "任务未开启失败告警")
	}
	if !ShouldTriggerFailureAlert(jobLog.Status) {
		return uc.logRepo.MarkAlertSkipped(ctx, logID, "执行状态不需要告警")
	}
	webhook, configured, err := uc.settings.GetFeishuWebhook(ctx)
	if err != nil {
		return err
	}
	if !configured {
		return uc.logRepo.MarkAlertSkipped(ctx, logID, "飞书 Webhook 未配置")
	}
	if err := uc.logRepo.MarkAlertPending(ctx, logID); err != nil {
		return err
	}
	payload := buildFeishuAlertCard(FeishuAlertCardInput{
		JobLogID:     jobLog.ID,
		JobID:        jobLog.JobID,
		JobName:      jobLog.JobName,
		ExecutorName: jobLog.ExecutorName,
		Status:       jobLog.Status,
		TriggerType:  jobLog.TriggerType,
		StartTime:    jobLog.StartTime,
		EndTime:      jobLog.EndTime,
		DurationMS:   jobLog.DurationMS,
		ExitCode:     jobLog.ExitCode,
		ErrorMessage: jobLog.ErrorMessage,
	})
	if err := uc.sendWithRetry(ctx, webhook, payload); err != nil {
		_ = uc.logRepo.MarkAlertFailed(ctx, logID, err.Error())
		return err
	}
	return uc.logRepo.MarkAlertSent(ctx, logID, time.Now())
}

func (uc *AlertUsecase) sendWithRetry(ctx context.Context, webhook string, payload any) error {
	var lastErr error
	for attempt := 1; attempt <= uc.maxAttempts; attempt++ {
		lastErr = uc.sender.SendCard(ctx, webhook, payload)
		if lastErr == nil {
			return nil
		}
		if attempt < uc.maxAttempts && uc.retryDelay > 0 {
			select {
			case <-time.After(uc.retryDelay):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
	return lastErr
}

func ShouldTriggerFailureAlert(status string) bool {
	return status == JobLogStatusFailed || status == JobLogStatusTimeout
}
