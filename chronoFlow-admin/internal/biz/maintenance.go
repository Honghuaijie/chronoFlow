package biz

import "context"

type ExecutorHealthClient interface {
	Health(context.Context, string, string) error
}

type JobLogMaintenanceRepo interface {
	MarkActiveLogsFailedByExecutorID(context.Context, int64, string) error
	MarkAllActiveLogsFailed(context.Context, string) error
	MarkKillingTimeoutLogsFailed(context.Context, int32, string) error
	DeleteExpiredLogs(context.Context, int32) ([]string, error)
}
