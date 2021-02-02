package slack

import (
	"fmt"
	"sync"
)

// SetLogger let's library users supply a logger, so that api debugging
// can be logged along with the application's debugging info.
func SetLogger(l logProvider) {
	loggerMutex.Lock()
	logger = ilogger{logProvider: l}
	loggerMutex.Unlock()
}

var (
	loggerMutex = new(sync.Mutex)
	logger      logInternal // A logger that can be set by consumers
)

// logProvider is a logger interface compatible with both stdlib and some
// 3rd party loggers such as logrus.
type logProvider interface {
	Output(int, string) error
}

// logInternal represents the internal logging api we use.
type logInternal interface {
	Print(...interface{})
	Printf(string, ...interface{})
	Println(...interface{})
	Output(int, string) error
}

// ilogger implements the additional methods used by our internal logging.
type ilogger struct {
	logProvider
}

// Println replicates the behaviour of the standard logger.
func (t ilogger) Println(v ...interface{}) {
	t.Output(2, fmt.Sprintln(v...))
}

// Printf replicates the behaviour of the standard logger.
func (t ilogger) Printf(format string, v ...interface{}) {
	t.Output(2, fmt.Sprintf(format, v...))
}

// Print replicates the behaviour of the standard logger.
func (t ilogger) Print(v ...interface{}) {
	t.Output(2, fmt.Sprint(v...))
}
