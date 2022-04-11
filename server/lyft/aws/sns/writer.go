package sns

import (
	"context"

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

	// WriteWithContext writes a message to an SNS topic with the specific
	// string payload and supports context propagation.
	// TODO: Actually add ctx propagation support
	WriteWithContext(ctx context.Context, payload []byte) error
}

// IOWriterAdapter allows us to use Writer in place of io.Writer
// Eventually we should just remove Writer and conform our implementations
// to that interface for consistence
type IOWriterAdapter struct {
	Writer Writer
}

func (w *IOWriterAdapter) Write(b []byte) (int, error) {
	err := w.Writer.Write(b)

	if err != nil {
		return 0, err
	}

	return len(b), nil
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

func (w *writer) WriteWithContext(_ context.Context, payload []byte) error {
	return w.Write(payload)
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

func (w *writerWithStats) WriteWithContext(_ context.Context, payload []byte) error {
	return w.Write(payload)
}

type noopWriter struct{}

func (n *noopWriter) Write(payload []byte) error {
	return nil
}

func (n *noopWriter) WriteWithContext(_ context.Context, payload []byte) error {
	return n.Write(payload)
}
