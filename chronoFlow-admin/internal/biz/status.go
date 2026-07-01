package biz

const (
	ScheduleStatusStopped = "stopped"
	ScheduleStatusRunning = "running"

	ExecutorStatusOnline  = "online"
	ExecutorStatusOffline = "offline"

	TriggerTypeManual = "manual"
	TriggerTypeCron   = "cron"

	JobLogStatusRunning = "running"
	JobLogStatusKilling = "killing"
	JobLogStatusSuccess = "success"
	JobLogStatusFailed  = "failed"
	JobLogStatusTimeout = "timeout"
	JobLogStatusSkipped = "skipped"
	JobLogStatusKilled  = "killed"

	AlertStatusNone    = "none"
	AlertStatusPending = "pending"
	AlertStatusSent    = "sent"
	AlertStatusFailed  = "failed"
	AlertStatusSkipped = "skipped"
)

func IsFinalJobLogStatus(status string) bool {
	switch status {
	case JobLogStatusSuccess, JobLogStatusFailed, JobLogStatusTimeout, JobLogStatusSkipped, JobLogStatusKilled:
		return true
	default:
		return false
	}
}

func CanCallbackUpdateJobLogStatus(status string) bool {
	return status == JobLogStatusRunning || status == JobLogStatusKilling
}
