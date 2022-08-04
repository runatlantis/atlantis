package logger

import (
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/config"
	"go.temporal.io/sdk/workflow"
)

func Info(ctx workflow.Context, msg string) {
	logger := workflow.GetLogger(ctx)
	kvs := config.ExtractLogKeyFields(ctx)

	logger.Info(msg, kvs...)
}

func Warn(ctx workflow.Context, msg string) {
	logger := workflow.GetLogger(ctx)
	kvs := config.ExtractLogKeyFields(ctx)

	logger.Warn(msg, kvs...)
}

func Error(ctx workflow.Context, msg string) {
	logger := workflow.GetLogger(ctx)
	kvs := config.ExtractLogKeyFields(ctx)

	logger.Error(msg, kvs...)
}

func Debug(ctx workflow.Context, msg string) {
	logger := workflow.GetLogger(ctx)
	kvs := config.ExtractLogKeyFields(ctx)

	logger.Debug(msg, kvs...)
}
