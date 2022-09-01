package logger

import (
	"github.com/runatlantis/atlantis/server/neptune/context"
	"go.temporal.io/sdk/workflow"
)

func Info(ctx workflow.Context, msg string) {
	logger := workflow.GetLogger(ctx)
	kvs := context.ExtractFieldsAsList(ctx)

	logger.Info(msg, kvs...)
}

func Warn(ctx workflow.Context, msg string) {
	logger := workflow.GetLogger(ctx)
	kvs := context.ExtractFieldsAsList(ctx)

	logger.Warn(msg, kvs...)
}

func Error(ctx workflow.Context, msg string, additionalKVs ...interface{}) {
	logger := workflow.GetLogger(ctx)
	kvs := context.ExtractFieldsAsList(ctx)

	logger.Error(msg, append(kvs, additionalKVs)...)
}

func Debug(ctx workflow.Context, msg string) {
	logger := workflow.GetLogger(ctx)
	kvs := context.ExtractFieldsAsList(ctx)

	logger.Debug(msg, kvs...)
}
