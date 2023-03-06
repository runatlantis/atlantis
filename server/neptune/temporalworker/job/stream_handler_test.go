package job_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events/terraform/filter"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/neptune/temporalworker/job"
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
		logs := []string{outputMsg, fmt.Sprintf("[%s] New Line", regexString)}
		testJobStore := &strictTestStore{
			t: t,
			write: struct {
				runners []*testStore
				count   int
			}{
				runners: []*testStore{
					{
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
					{
						t: t,
						Msg: job.OutputLine{
							JobID: jobID,
							Line:  outputMsg,
						},
					},
				},
			},
		}

		streamHandler := job.NewStreamHandler(
			testJobStore,
			testReceiverRegistry,
			valid.TerraformLogFilters(logFilter),
			logging.NewNoopCtxLogger(t),
		)

		ch := streamHandler.RegisterJob(jobID)

		for _, line := range logs {
			ch <- line
		}
		close(ch)
	})
}

func TestStreamHandler_Close(t *testing.T) {
	jobID := "1234"

	t.Run("closes receiver registry", func(t *testing.T) {
		testReceiverRegistry := &strictTestReceiverRegistry{
			t: t,
			close: struct {
				runners []*testReceiverRegistry
				count   int
			}{
				runners: []*testReceiverRegistry{
					{
						t:     t,
						JobID: jobID,
					},
				},
			},
		}

		testStore := &strictTestStore{
			t: t,
			close: struct {
				runners []*testStore
				count   int
			}{
				runners: []*testStore{
					{
						t:      t,
						JobID:  jobID,
						Status: job.Complete,
					},
				},
			},
		}
		streamHandler := job.NewStreamHandler(
			testStore,
			testReceiverRegistry,
			valid.TerraformLogFilters{},
			logging.NewNoopCtxLogger(t),
		)
		err := streamHandler.CloseJob(context.Background(), jobID)
		assert.NoError(t, err)
	})
}

// Should clean up josb store and receiver registry
func TestStreamHandler_Cleanup(t *testing.T) {
	t.Run("cleans up store and receiver registry", func(t *testing.T) {
		testReceiverRegistry := &strictTestReceiverRegistry{
			t: t,
			cleanup: struct {
				runners []*testReceiverRegistry
				count   int
			}{
				runners: []*testReceiverRegistry{
					{
						t: t,
					},
				},
			},
		}

		testStore := &strictTestStore{
			t: t,
			cleanup: struct {
				runners []*testStore
				count   int
			}{
				runners: []*testStore{
					{
						t: t,
					},
				},
			},
		}
		streamHandler := job.NewStreamHandler(
			testStore,
			testReceiverRegistry,
			valid.TerraformLogFilters{},
			logging.NewNoopCtxLogger(t),
		)
		err := streamHandler.CleanUp(context.Background())
		assert.NoError(t, err)
	})
}
