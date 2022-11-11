// Copyright 2017 HootSuite Media Inc.
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// Modified hereafter by contributors to runatlantis/atlantis.
//
// Package logging handles logging throughout Atlantis.
package logging

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	ctxInternal "github.com/runatlantis/atlantis/server/neptune/context"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
	logurzap "logur.dev/adapter/zap"
	"logur.dev/logur"
)

// Logger is the logging interface used throughout the code.
// All new code should leverage this over SimpleLogging
type Logger interface {
	logur.Logger
	logur.LoggerContext
	Closer
}

// Closer is responsible for closing an underlying logger.  This may mean,
// flushing buffer contents on exit, but is up to the implementation.
type Closer interface {
	Close() error
}

type logger struct {
	logur.LoggerFacade
	Closer
}

func NewLoggerFromLevel(lvl LogLevel) (*logger, error) { //nolint:revive // avoiding refactor while adding linter action
	structuredLogger, err := NewStructuredLoggerFromLevel(lvl)
	if err != nil {
		return &logger{}, err

	}

	ctxLogger := logur.WithContextExtractor(
		structuredLogger,
		func(ctx context.Context) map[string]interface{} {
			return ctxInternal.ExtractFields(ctx)
		},
	)

	return &logger{
		LoggerFacade: ctxLogger,
		Closer:       structuredLogger,
	}, nil

}

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_simple_logging.go SimpleLogging

// Deprecated: Use Logger instead
// SimpleLogging is the interface used for logging throughout the codebase.
type SimpleLogging interface {
	// These basically just fmt.Sprintf() the message and args.

	Debugf(format string, a ...interface{})
	Infof(format string, a ...interface{})
	Warnf(format string, a ...interface{})
	Errorf(format string, a ...interface{})

	Log(level LogLevel, format string, a ...interface{})
	SetLevel(lvl LogLevel)

	// With adds a variadic number of fields to the logging context. It accepts a
	// mix of strongly-typed Field objects and loosely-typed key-value pairs. When
	// processing pairs, the first element of the pair is used as the field key
	// and the second as the field value.
	With(a ...interface{}) SimpleLogging

	// Closes the underlying log writer, flushing anything from the buffer
	Close() error
}

type StructuredLogger struct {
	z     *zap.SugaredLogger
	level zap.AtomicLevel
	logur.Logger
}

func NewStructuredLoggerFromLevel(lvl LogLevel) (*StructuredLogger, error) {
	cfg := zap.NewProductionConfig()

	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.Level = zap.NewAtomicLevelAt(lvl.zLevel)
	return newStructuredLogger(cfg)
}

func newStructuredLogger(cfg zap.Config) (*StructuredLogger, error) {
	baseLogger, err := cfg.Build()

	baseLogger = baseLogger.
		// ensures that the caller doesn't just say logging/simple_logger each time
		WithOptions(zap.AddCallerSkip(1)).
		WithOptions(zap.AddStacktrace(zapcore.WarnLevel)).
		// creates isolated context for all future kv pairs, name can be flexible as needed
		With(zap.Namespace("json"))

	if err != nil {
		return nil, errors.Wrap(err, " initializing structured logger")
	}

	return &StructuredLogger{
		z:      baseLogger.Sugar(),
		level:  cfg.Level,
		Logger: logurzap.New(baseLogger),
	}, nil
}

func (l *StructuredLogger) With(a ...interface{}) SimpleLogging {
	return &StructuredLogger{
		z:     l.z.With(a...),
		level: l.level,
	}
}

func (l *StructuredLogger) Debugf(format string, a ...interface{}) {
	l.z.Debugf(format, a...)
}

func (l *StructuredLogger) Infof(format string, a ...interface{}) {
	l.z.Infof(format, a...)
}

func (l *StructuredLogger) Warnf(format string, a ...interface{}) {
	l.z.Warnf(format, a...)
}

func (l *StructuredLogger) Errorf(format string, a ...interface{}) {
	l.z.Errorf(format, a...)
}

func (l *StructuredLogger) Log(level LogLevel, format string, a ...interface{}) {
	switch level {
	case Debug:
		l.Debugf(format, a...)
	case Info:
		l.Infof(format, a...)
	case Warn:
		l.Warnf(format, a...)
	case Error:
		l.Errorf(format, a...)
	}
}

func (l *StructuredLogger) SetLevel(lvl LogLevel) {
	if l != nil {
		l.level.SetLevel(lvl.zLevel)
	}
}

func (l *StructuredLogger) Close() error {
	return l.z.Sync()
}

// NewNoopLogger creates a logger instance that discards all logs and never
// writes them. Used for testing.
func NewNoopLogger(t *testing.T) SimpleLogging {
	level := zap.DebugLevel
	return &StructuredLogger{
		z:     zaptest.NewLogger(t, zaptest.Level(level)).Sugar(),
		level: zap.NewAtomicLevelAt(level),
	}
}

// NewNoopLogger creates a logger instance that discards all logs and never
// writes them. Used for testing.
func NewNoopCtxLogger(t *testing.T) Logger {
	level := zap.DebugLevel
	zapLogger := zaptest.NewLogger(t, zaptest.Level(level))
	sLogger := &StructuredLogger{
		z:      zapLogger.Sugar(),
		level:  zap.NewAtomicLevelAt(level),
		Logger: logurzap.New(zapLogger),
	}

	return &logger{
		LoggerFacade: logur.WithContextExtractor(
			sLogger,
			func(ctx context.Context) map[string]interface{} {
				return ctxInternal.ExtractFields(ctx)
			},
		),
	}
}

type LogLevel struct {
	zLevel   zapcore.Level
	shortStr string
}

var (
	Debug = LogLevel{
		zLevel:   zapcore.DebugLevel,
		shortStr: "DBUG",
	}
	Info = LogLevel{
		zLevel:   zapcore.InfoLevel,
		shortStr: "INFO",
	}
	Warn = LogLevel{
		zLevel:   zapcore.WarnLevel,
		shortStr: "WARN",
	}
	Error = LogLevel{
		zLevel:   zapcore.ErrorLevel,
		shortStr: "EROR",
	}
)
