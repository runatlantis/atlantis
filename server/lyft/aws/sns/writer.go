package sns

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	awsSns "github.com/aws/aws-sdk-go/service/sns"
	snsApi "github.com/aws/aws-sdk-go/service/sns/snsiface"
	"github.com/runatlantis/atlantis/server/events/metrics"
	"github.com/uber-go/tally"
)

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_writer.go Writer

type Writer interface {
	// Write a message to an SNS topic with the specified string payload
	Write([]byte) error
}

func NewNoopWriter() Writer {
	return &noopWriter{}
}

// NewWriterWithStats returns a new instance of Writer that will connect to the specifed
// sns topic using the specified session
func NewWriterWithStats(
	session client.ConfigProvider,
	topicArn string,
	scope tally.Scope,
) Writer {
	return &writerWithStats{
		scope: scope,
		Writer: &writer{
			client:   awsSns.New(session),
			topicArn: aws.String(topicArn),
		},
	}
}

type writer struct {
	client   snsApi.SNSAPI
	topicArn *string
}

func (w *writer) Write(payload []byte) error {
	_, err := w.client.Publish(&awsSns.PublishInput{
		Message:  aws.String(string(payload)),
		TopicArn: w.topicArn,
	})
	return err
}

// writerWithStats decorator to track writing to sns topic
type writerWithStats struct {
	Writer
	scope tally.Scope
}

func (w *writerWithStats) Write(payload []byte) error {
	executionTime := w.scope.Timer(metrics.ExecutionTimeMetric).Start()
	defer executionTime.Stop()

	if err := w.Writer.Write(payload); err != nil {
		w.scope.Counter(metrics.ExecutionErrorMetric).Inc(1)
		return err
	}

	w.scope.Counter(metrics.ExecutionSuccessMetric).Inc(1)
	return nil
}

type noopWriter struct{}

func (n *noopWriter) Write(payload []byte) error {
	return nil
}
