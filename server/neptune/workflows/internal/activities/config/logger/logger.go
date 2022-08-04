package logger

import (
	"context"

	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/config"
	"go.temporal.io/sdk/activity"
)

func Info(ctx context.Context, msg string) {
	logger := activity.GetLogger(ctx)
	kvs := config.ExtractLogKeyFields(ctx)

	logger.Info(msg, kvs...)
}

func Warn(ctx context.Context, msg string) {
	logger := activity.GetLogger(ctx)
	kvs := config.ExtractLogKeyFields(ctx)

	logger.Warn(msg, kvs...)
}

func Error(ctx context.Context, msg string) {
	logger := activity.GetLogger(ctx)
	kvs := config.ExtractLogKeyFields(ctx)

	logger.Error(msg, kvs...)
}

func Debug(ctx context.Context, msg string) {
	logger := activity.GetLogger(ctx)
	kvs := config.ExtractLogKeyFields(ctx)

	logger.Debug(msg, kvs...)
}
