package sqs

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/uber-go/tally/v4"
)

// Queue mirrors a strict set of AWS SQS Interface
type Queue interface {
	ReceiveMessage(ctx context.Context, req *sqs.ReceiveMessageInput, optFns ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error)
	DeleteMessage(ctx context.Context, req *sqs.DeleteMessageInput, optFns ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error)
}

// QueueWithStats proxies request to the underlying queue and wraps it with metrics
// and error handling.
type QueueWithStats struct {
	Queue
	Scope    tally.Scope
	QueueURL string
}

func (q *QueueWithStats) ReceiveMessage(ctx context.Context, req *sqs.ReceiveMessageInput, optFns ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error) {
	scope := q.Scope.SubScope(ReceiveMessageMetricName)

	timer := scope.Timer(Latency).Start()
	defer timer.Stop()

	successCount := scope.Counter(Success)
	errorCount := scope.Counter(Error)

	response, err := q.Queue.ReceiveMessage(ctx, req, optFns...)
	// only consider it a failure if the error isn't due to a context cancellation
	if err != nil && !errors.Is(err, context.Canceled) {
		errorCount.Inc(1)
		return response, fmt.Errorf("receiving messages from queue: %s: %w", q.QueueURL, err)
	}

	successCount.Inc(1)
	return response, err
}

func (q *QueueWithStats) DeleteMessage(ctx context.Context, req *sqs.DeleteMessageInput, optFns ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error) {
	scope := q.Scope.SubScope(DeleteMessageMetricName)

	timer := scope.Timer(Latency).Start()
	defer timer.Stop()

	successCount := scope.Counter(Success)
	errorCount := scope.Counter(Error)

	response, err := q.Queue.DeleteMessage(ctx, req, optFns...)
	if err != nil {
		errorCount.Inc(1)
		return response, fmt.Errorf("deleting messages from queue: %s, receipt handle: %s: %w", q.QueueURL, *req.ReceiptHandle, err)
	}

	successCount.Inc(1)
	return response, err
}
