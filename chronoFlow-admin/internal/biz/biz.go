package biz

import (
	"context"

	"github.com/google/wire"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(
	NewUserUsecase,
	NewExecutorUsecase,
	NewJobUsecase,
	NewGlueUsecase,
	NewJobLogUsecase,
	NewCallbackUsecase,
	NewSystemSettingUsecase,
	NewFeishuAlertSender,
	NewAlertUsecase,
	NewJobRunConfig,
	NewCallbackConfig,
	wire.Bind(new(AlertDispatcher), new(*AlertUsecase)),
)

type Transaction interface {
	ExecTx(context.Context, func(ctx context.Context) error) error
}
