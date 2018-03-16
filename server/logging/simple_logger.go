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
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"unicode"
)

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_simple_logging.go SimpleLogging

// SimpleLogging is the interface that our SimpleLogger implements.
// It's really only used for mocking when we need to test what's being logged.
type SimpleLogging interface {
	Debug(format string, a ...interface{})
	Info(format string, a ...interface{})
	Warn(format string, a ...interface{})
	Err(format string, a ...interface{})
	Log(level LogLevel, format string, a ...interface{})
	// Underlying returns the underlying logger.
	Underlying() *log.Logger
	// GetLevel returns the current log level.
	GetLevel() LogLevel
}

// SimpleLogger wraps the standard logger with leveled logging
// and the ability to store log history for later adding it
// to a VCS comment.
type SimpleLogger struct {
	// Source is added as a prefix to each log entry.
	// It's useful if you want to trace a log entry back to a
	// context, for example a pull request id.
	Source string
	// History stores all log entries ever written using
	// this logger. This is safe for short-lived loggers
	// like those used during plan/apply commands.
	History     bytes.Buffer
	Logger      *log.Logger
	KeepHistory bool
	Level       LogLevel
}

type LogLevel int

const (
	Debug LogLevel = iota
	Info
	Warn
	Error
)

// NewSimpleLogger creates a new logger.
// source is added as a prefix to each log entry. It's useful if you want to
// trace a log entry back to a specific context, for example a pull request id.
// logger is the underlying logger. If nil will create a logger from stdlib.
// keepHistory set to true will store all log entries written using this logger.
// level will set the level at which logs >= than that level will be written.
// If keepHistory is set to true, we'll store logs at all levels, regardless of
// what level is set to.
func NewSimpleLogger(source string, logger *log.Logger, keepHistory bool, level LogLevel) *SimpleLogger {
	if logger == nil {
		flags := log.LstdFlags
		if level == Debug {
			// If we're using debug logging, we also have the logger print the
			// filename the log comes from with log.Lshortfile.
			flags = flags | log.Lshortfile
		}
		logger = log.New(os.Stderr, "", flags)
	}
	return &SimpleLogger{
		Source:      source,
		Logger:      logger,
		Level:       level,
		KeepHistory: keepHistory,
	}
}

// NewNoopLogger creates a logger instance that discards all logs and never
// writes them. Used for testing.
func NewNoopLogger() *SimpleLogger {
	logger := log.New(os.Stderr, "", 0)
	logger.SetOutput(ioutil.Discard)
	return &SimpleLogger{
		Source:      "",
		Logger:      logger,
		Level:       Info,
		KeepHistory: false,
	}
}

// ToLogLevel converts a log level string to a valid LogLevel object.
// If the string doesn't match a level, it will return Info.
func ToLogLevel(levelStr string) LogLevel {
	switch levelStr {
	case "debug":
		return Debug
	case "info":
		return Info
	case "warn":
		return Warn
	case "error":
		return Error
	}
	return Info
}

// Debug logs at debug level.
func (l *SimpleLogger) Debug(format string, a ...interface{}) {
	l.Log(Debug, format, a...)
}

// Info logs at info level.
func (l *SimpleLogger) Info(format string, a ...interface{}) {
	l.Log(Info, format, a...)
}

// Warn logs at warn level.
func (l *SimpleLogger) Warn(format string, a ...interface{}) {
	l.Log(Warn, format, a...)
}

// Err logs at error level.
func (l *SimpleLogger) Err(format string, a ...interface{}) {
	l.Log(Error, format, a...)
}

// Log writes the log at level.
func (l *SimpleLogger) Log(level LogLevel, format string, a ...interface{}) {
	levelStr := l.levelToString(level)
	msg := l.capitalizeFirstLetter(fmt.Sprintf(format, a...))

	// Only log this message if configured to log at this level.
	if l.Level <= level {
		// Calling .Output instead of Printf so we can change the calldepth
		// param to 3. The default is 2 which would identify the log as coming
		// from this file and line every time instead of our caller's.
		l.Logger.Output(3, fmt.Sprintf("[%s] %s: %s\n", levelStr, l.Source, msg)) // nolint: errcheck
	}

	// Keep history at all log levels.
	if l.KeepHistory {
		l.saveToHistory(levelStr, msg)
	}
}

// Underlying returns the underlying logger.
func (l *SimpleLogger) Underlying() *log.Logger {
	return l.Logger
}

// GetLevel returns the current log level of the logger.
func (l *SimpleLogger) GetLevel() LogLevel {
	return l.Level
}

func (l *SimpleLogger) saveToHistory(level string, msg string) {
	l.History.WriteString(fmt.Sprintf("[%s] %s\n", level, msg))
}

func (l *SimpleLogger) capitalizeFirstLetter(s string) string {
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

func (l *SimpleLogger) levelToString(level LogLevel) string {
	switch level {
	case Debug:
		return "DEBUG"
	case Info:
		return "INFO"
	case Warn:
		return "WARN"
	case Error:
		return "ERROR"
	}
	return "NOLEVEL"
}
