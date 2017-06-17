package logging

import (
	"bytes"
	"fmt"
	"log"
	"unicode"
)

type SimpleLogger struct {
	Source      string
	History     bytes.Buffer
	Log         *log.Logger
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

func NewSimpleLogger(source string, log *log.Logger, keepHistory bool, level LogLevel) *SimpleLogger {
	return &SimpleLogger{
		Source:      source,
		Log:         log,
		Level:       level,
		KeepHistory: keepHistory,
	}
}

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

func (l *SimpleLogger) Debug(format string, a ...interface{}) {
	l.log(Debug, format, a...)
}

func (l *SimpleLogger) Info(format string, a ...interface{}) {
	l.log(Info, format, a...)
}

func (l *SimpleLogger) Warn(format string, a ...interface{}) {
	l.log(Warn, format, a...)
}

func (l *SimpleLogger) Err(format string, a ...interface{}) {
	l.log(Error, format, a...)
}

func (l *SimpleLogger) log(level LogLevel, format string, a ...interface{}) {
	levelStr := l.levelToString(level)
	msg := l.capitalizeFirstLetter(fmt.Sprintf(format, a...))

	// only log this message if configured to log at this level
	if l.Level <= level {
		l.Log.Printf("[%s] %s: %s\n", levelStr, l.Source, msg)
	}

	// keep history at all log levels
	if l.KeepHistory {
		l.saveToHistory(levelStr, msg)
	}
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
