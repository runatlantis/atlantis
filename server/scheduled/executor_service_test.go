package scheduled

import (
	"testing"
	"time"

	pegomock "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/scheduled/mocks"
)

func TestExecutorService_Run(t *testing.T) {
	pegomock.RegisterMockTestingT(t)
	mockJob := mocks.NewMockJob()
	type fields struct {
		log  logging.SimpleLogging
		jobs []JobDefinition
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "test",
			fields: fields{
				log: logging.NewNoopLogger(t),
				jobs: []JobDefinition{
					{
						Job:    mockJob,
						Period: 1 * time.Second,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ExecutorService{
				log:  tt.fields.log,
				jobs: make([]JobDefinition, 0),
			}
			s.AddJob(tt.fields.jobs[0])
			go s.Run()
			time.Sleep(1050 * time.Millisecond)
			mockJob.VerifyWasCalledOnce().Run()
		})
	}
}
