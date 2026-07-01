package biz

import "context"

type ExecutorHealthClient interface {
	Health(context.Context, string, string) error
}

type JobLogMaintenanceRepo interface {
	MarkActiveLogsFailedByExecutorID(context.Context, int64, string) error
	MarkActiveLogsFailedByExecutorIDReturningIDs(context.Context, int64, string) ([]int64, error)
	MarkAllActiveLogsFailed(context.Context, string) error
	MarkAllActiveLogsFailedReturningIDs(context.Context, string) ([]int64, error)
	MarkKillingTimeoutLogsFailed(context.Context, int32, string) error
	MarkKillingTimeoutLogsFailedReturningIDs(context.Context, int32, string) ([]int64, error)
	DeleteExpiredLogs(context.Context, int32) ([]string, error)
}
