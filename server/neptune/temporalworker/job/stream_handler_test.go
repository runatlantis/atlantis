package job_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/terraform/filter"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/neptune/temporalworker/job"
	"github.com/runatlantis/atlantis/server/neptune/terraform"
	"github.com/stretchr/testify/assert"
)

func TestStreamHandler_Handle(t *testing.T) {
	regexString := "filter"
	regex, err := regexp.Compile(regexString)
	assert.Nil(t, err)

	logFilter := filter.LogFilter{
		Regexes: []*regexp.Regexp{
			regex,
		},
	}
	jobID := "1234"
	outputMsg := "a"

	t.Run("broadcasts and stores filtered logs", func(t *testing.T) {
		outputCh := make(chan *job.OutputLine)
		logs := []string{outputMsg, fmt.Sprintf("[%s] New Line", regexString)}
		testJobStore := &strictTestStore{
			t: t,
			write: struct {
				runners []*testStore
				count   int
			}{
				runners: []*testStore{
					&testStore{
						t:      t,
						JobID:  jobID,
						Output: outputMsg,
					},
				},
			},
		}
		testReceiverRegistry := &strictTestReceiverRegistry{
			t: t,
			broadcast: struct {
				runners []*testReceiverRegistry
				count   int
			}{
				runners: []*testReceiverRegistry{
					&testReceiverRegistry{
						t: t,
						Msg: job.OutputLine{
							JobID: jobID,
							Line:  outputMsg,
						},
					},
				},
			},
		}

		streamHandler := job.StreamHandler{
			JobOutput:        outputCh,
			Store:            testJobStore,
			ReceiverRegistry: testReceiverRegistry,
			LogFilter:        logFilter,
			Logger:           logging.NewNoopCtxLogger(t),
		}

		go streamHandler.Handle()

		for _, line := range logs {
			outputCh <- &job.OutputLine{
				JobID: jobID,
				Line:  line,
			}
		}
		close(outputCh)
	})
}

func TestStreamHandler_Stream(t *testing.T) {
	jobID := "1234"
	outputMsg := "a"

	t.Run("streams to main terraform channel", func(t *testing.T) {
		logs := []string{outputMsg, outputMsg}

		// Buffered channel to simplify testing since it's not blocking
		mainTfCh := make(chan *job.OutputLine, len(logs))
		streamHandler := job.StreamHandler{
			JobOutput:        mainTfCh,
			Store:            &testStore{},
			ReceiverRegistry: &testReceiverRegistry{},
			Logger:           logging.NewNoopCtxLogger(t),
		}

		buffer := make(chan terraform.Line, 2)
		go streamHandler.Stream(context.TODO(), jobID, buffer)

		for _, line := range logs {
			buffer <- terraform.Line{
				Line: line,
			}
		}

		gotLogs := []string{}

		// Read main terraform output channel
	outside:
		for {
			select {
			case line := <-mainTfCh:
				gotLogs = append(gotLogs, line.Line)

			// give buffer time for logs to be streamed to main terraform channe;
			case <-time.After(2 * time.Second):
				break outside
			}
		}
		assert.Equal(t, logs, gotLogs)
	})

	t.Run("error when line error", func(t *testing.T) {
		logError := errors.New("error")
		expectedError := errors.Wrap(logError, "executing command")
		logs := []terraform.Line{
			terraform.Line{
				Err: errors.New("error"),
			},
		}

		// Buffered channel to simplify testing since it's not blocking
		mainTfCh := make(chan *job.OutputLine, len(logs))
		streamHandler := job.StreamHandler{
			JobOutput:        mainTfCh,
			Store:            &testStore{},
			ReceiverRegistry: &testReceiverRegistry{},
			Logger:           logging.NewNoopCtxLogger(t),
		}

		buffer := make(chan terraform.Line, 2)

		go func() {
			for _, line := range logs {
				buffer <- line
			}
		}()

		gotErr := streamHandler.Stream(context.TODO(), jobID, buffer)
		assert.Equal(t, expectedError.Error(), gotErr.Error())
	})
}

func TestStreamHandler_Close(t *testing.T) {
	jobID := "1234"

	t.Run("closes receiver registry", func(t *testing.T) {
		testReceiverRegistry := strictTestReceiverRegistry{
			t: t,
			close: struct {
				runners []*testReceiverRegistry
				count   int
			}{
				runners: []*testReceiverRegistry{
					&testReceiverRegistry{
						t:     t,
						JobID: jobID,
					},
				},
			},
		}

		testStore := strictTestStore{
			t: t,
			close: struct {
				runners []*testStore
				count   int
			}{
				runners: []*testStore{
					&testStore{
						t:      t,
						JobID:  jobID,
						Status: job.Complete,
					},
				},
			},
		}
		streamHandler := job.StreamHandler{
			Store:            testStore,
			ReceiverRegistry: testReceiverRegistry,
			Logger:           logging.NewNoopCtxLogger(t),
		}
		streamHandler.Close(context.Background(), jobID)
	})
}
