package data

import (
	"time"

	"gorm.io/plugin/soft_delete"
)

type ModelOpt struct {
	CreatedAt time.Time             `json:"createdAt" gorm:"not null;autoCreateTime"`
	CreatedBy string                `json:"createdBy" gorm:"not null;size:64;default:''"`
	UpdatedAt time.Time             `json:"updatedAt" gorm:"not null;autoUpdateTime"`
	UpdatedBy string                `json:"updatedBy" gorm:"not null;size:64;default:''"`
	Deleted   soft_delete.DeletedAt `json:"deleted" gorm:"softDelete:flag;not null;default:0"`
}

type User struct {
	ID uint64 `json:"id" gorm:"primaryKey;autoIncrement"`
	ModelOpt
	Name  string `json:"name" gorm:"size:128;not null"`
	Email string `json:"email" gorm:"size:128;not null;index"`
	Phone string `json:"phone" gorm:"size:32;default:''"`
}

func (User) TableName() string {
	return "users"
}

type Executor struct {
	ID uint64 `json:"id" gorm:"primaryKey;autoIncrement"`
	ModelOpt
	Name               string     `json:"name" gorm:"size:100;not null"`
	Address            string     `json:"address" gorm:"size:255;not null"`
	TokenCiphertext    string     `json:"tokenCiphertext" gorm:"column:token_ciphertext;size:1000;not null"`
	Description        string     `json:"description" gorm:"size:500;default:''"`
	Status             string     `json:"status" gorm:"size:32;not null;index"`
	HeartbeatFailCount int32      `json:"heartbeatFailCount" gorm:"column:heartbeat_fail_count;not null;default:0"`
	LastHeartbeatTime  *time.Time `json:"lastHeartbeatTime" gorm:"column:last_heartbeat_time"`
}

func (Executor) TableName() string {
	return "executors"
}

type Job struct {
	ID uint64 `json:"id" gorm:"primaryKey;autoIncrement"`
	ModelOpt
	ExecutorID          uint64 `json:"executorId" gorm:"column:executor_id;not null;index"`
	Name                string `json:"name" gorm:"size:100;not null"`
	CronExpr            string `json:"cronExpr" gorm:"column:cron_expr;size:100;not null"`
	TimeoutSeconds      int32  `json:"timeoutSeconds" gorm:"column:timeout_seconds;not null;default:600"`
	ScheduleStatus      string `json:"scheduleStatus" gorm:"column:schedule_status;size:32;not null;index"`
	Description         string `json:"description" gorm:"size:500;default:''"`
	FailureAlertEnabled bool   `json:"failureAlertEnabled" gorm:"column:failure_alert_enabled;not null;default:false"`
}

func (Job) TableName() string {
	return "jobs"
}

type JobGlue struct {
	ID uint64 `json:"id" gorm:"primaryKey;autoIncrement"`
	ModelOpt
	JobID   uint64 `json:"jobId" gorm:"column:job_id;not null;uniqueIndex"`
	Content string `json:"content" gorm:"type:text;not null"`
}

func (JobGlue) TableName() string {
	return "job_glues"
}

type JobLog struct {
	ID uint64 `json:"id" gorm:"primaryKey;autoIncrement"`
	ModelOpt
	JobID                uint64     `json:"jobId" gorm:"column:job_id;not null;index"`
	JobName              string     `json:"jobName" gorm:"column:job_name;size:100;not null"`
	ExecutorID           uint64     `json:"executorId" gorm:"column:executor_id;not null;index"`
	ExecutorName         string     `json:"executorName" gorm:"column:executor_name;size:100;not null"`
	ExecutorAddress      string     `json:"executorAddress" gorm:"column:executor_address;size:255;not null"`
	CronExpr             string     `json:"cronExpr" gorm:"column:cron_expr;size:100;not null"`
	TimeoutSeconds       int32      `json:"timeoutSeconds" gorm:"column:timeout_seconds;not null"`
	GlueSnapshot         string     `json:"glueSnapshot" gorm:"column:glue_snapshot;type:mediumtext;not null"`
	TriggerType          string     `json:"triggerType" gorm:"column:trigger_type;size:32;not null;index"`
	Status               string     `json:"status" gorm:"size:32;not null;index"`
	StartTime            time.Time  `json:"startTime" gorm:"column:start_time;not null"`
	EndTime              *time.Time `json:"endTime" gorm:"column:end_time"`
	DurationMS           int64      `json:"durationMs" gorm:"column:duration_ms;not null;default:0"`
	ExitCode             *int32     `json:"exitCode" gorm:"column:exit_code"`
	LogPath              string     `json:"logPath" gorm:"column:log_path;size:500;default:''"`
	LogSizeBytes         int64      `json:"logSizeBytes" gorm:"column:log_size_bytes;not null;default:0"`
	LogTruncated         bool       `json:"logTruncated" gorm:"column:log_truncated;not null;default:false"`
	ErrorMessage         string     `json:"errorMessage" gorm:"column:error_message;type:text"`
	AlertEnabledSnapshot bool       `json:"alertEnabledSnapshot" gorm:"column:alert_enabled_snapshot;not null;default:false"`
	AlertStatus          string     `json:"alertStatus" gorm:"column:alert_status;size:32;not null;default:'none';index"`
	AlertError           string     `json:"alertError" gorm:"column:alert_error;type:text"`
	AlertSentAt          *time.Time `json:"alertSentAt" gorm:"column:alert_sent_at"`
}

func (JobLog) TableName() string {
	return "job_logs"
}
