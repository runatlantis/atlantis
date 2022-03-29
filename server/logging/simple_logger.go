// Copyright 2017 HootSuite Media Inc.
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
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
	"testing"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
)

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_simple_logging.go SimpleLogging

// SimpleLogging is the interface used for logging throughout the codebase.
type SimpleLogging interface {

	// These basically just fmt.Sprintf() the message and args.
	Debug(format string, a ...interface{})
	Info(format string, a ...interface{})
	Warn(format string, a ...interface{})
	Err(format string, a ...interface{})
	Log(level LogLevel, format string, a ...interface{})
	SetLevel(lvl LogLevel)

	// With adds a variadic number of fields to the logging context. It accepts a
	// mix of strongly-typed Field objects and loosely-typed key-value pairs. When
	// processing pairs, the first element of the pair is used as the field key
	// and the second as the field value.
	With(a ...interface{}) SimpleLogging

	// Flushes anything left in the buffer
	Flush() error
}

type StructuredLogger struct {
	z     *zap.SugaredLogger
	level zap.AtomicLevel
}

func NewStructuredLoggerFromLevel(lvl LogLevel) (SimpleLogging, error) {
	cfg := zap.NewProductionConfig()

	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.Level = zap.NewAtomicLevelAt(lvl.zLevel)
	return newStructuredLogger(cfg)
}

func NewStructuredLogger() (SimpleLogging, error) {
	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
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
		z:     baseLogger.Sugar(),
		level: cfg.Level,
	}, nil
}

func (l *StructuredLogger) With(a ...interface{}) SimpleLogging {
	return &StructuredLogger{
		z:     l.z.With(a...),
		level: l.level,
	}
}

func (l *StructuredLogger) Debug(format string, a ...interface{}) {
	l.z.Debugf(format, a...)
}

func (l *StructuredLogger) Info(format string, a ...interface{}) {
	l.z.Infof(format, a...)
}

func (l *StructuredLogger) Warn(format string, a ...interface{}) {
	l.z.Warnf(format, a...)
}

func (l *StructuredLogger) Err(format string, a ...interface{}) {
	l.z.Errorf(format, a...)
}

func (l *StructuredLogger) Log(level LogLevel, format string, a ...interface{}) {
	switch level {
	case Debug:
		l.Debug(format, a...)
	case Info:
		l.Info(format, a...)
	case Warn:
		l.Warn(format, a...)
	case Error:
		l.Err(format, a...)
	}
}

func (l *StructuredLogger) SetLevel(lvl LogLevel) {
	if l != nil {
		l.level.SetLevel(lvl.zLevel)
	}
}

func (l *StructuredLogger) Flush() error {
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
