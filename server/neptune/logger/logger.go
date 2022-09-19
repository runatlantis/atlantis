package logger

import (
	"context"

	internalContext "github.com/runatlantis/atlantis/server/neptune/context"
	"go.temporal.io/sdk/activity"
)

func Info(ctx context.Context, msg string) {
	logger := activity.GetLogger(ctx)
	kvs := internalContext.ExtractFieldsAsList(ctx)

	logger.Info(msg, kvs...)
}

func Warn(ctx context.Context, msg string) {
	logger := activity.GetLogger(ctx)
	kvs := internalContext.ExtractFieldsAsList(ctx)

	logger.Warn(msg, kvs...)
}

func Error(ctx context.Context, msg string, additionalKVs ...interface{}) {
	logger := activity.GetLogger(ctx)
	kvs := internalContext.ExtractFieldsAsList(ctx)

	logger.Error(msg, append(kvs, additionalKVs)...)
}

func Debug(ctx context.Context, msg string) {
	logger := activity.GetLogger(ctx)
	kvs := internalContext.ExtractFieldsAsList(ctx)

	logger.Debug(msg, kvs...)
}
