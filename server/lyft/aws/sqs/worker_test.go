package sqs_test

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	awssqs "github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	. "github.com/petergtz/pegomock"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/lyft/aws/sqs"
	"github.com/runatlantis/atlantis/server/lyft/aws/sqs/mocks"
	"github.com/runatlantis/atlantis/server/lyft/aws/sqs/mocks/matchers"
	. "github.com/runatlantis/atlantis/testing"
	"github.com/uber-go/tally/v4"

	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type testQueue struct {
	receiveError error

	// represents an underlying queue with messages.
	// ReceiveMessage retrieves these messages while
	// DeleteMessage will remove items from this list.
	// Note: This is not threadsafe, tests should only have one thread
	// capable of mutating this.
	messages []types.Message

	// This should be called during ReceiveMessage so that
	// future calls cannot be made which therefore ends the worker.
	cancel context.CancelFunc
}

func (t *testQueue) ReceiveMessage(ctx context.Context, req *awssqs.ReceiveMessageInput, optFns ...func(*awssqs.Options)) (*awssqs.ReceiveMessageOutput, error) {
	t.cancel()
	if t.receiveError != nil {
		return nil, t.receiveError
	}

	return &awssqs.ReceiveMessageOutput{Messages: t.messages}, nil
}

func (t *testQueue) DeleteMessage(ctx context.Context, req *awssqs.DeleteMessageInput, optFns ...func(*awssqs.Options)) (*awssqs.DeleteMessageOutput, error) {
	var prunedMsgs []types.Message

	// remove deleted message from array.
	for _, msg := range t.messages {
		if msg.ReceiptHandle == req.ReceiptHandle {
			continue
		}
		prunedMsgs = append(prunedMsgs, msg)
	}

	t.messages = prunedMsgs
	return &awssqs.DeleteMessageOutput{}, nil
}

func TestWorker_Success(t *testing.T) {
	RegisterMockTestingT(t)
	ctx, cancelFunc := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	testScope := tally.NewTestScope("test", nil)

	expectedMessage := types.Message{
		Body:          aws.String("body"),
		ReceiptHandle: aws.String("receipt_handle"),
		MessageId:     aws.String("message_id"),
	}
	tq := &testQueue{
		messages: []types.Message{
			expectedMessage,
		},
		cancel: cancelFunc,
	}
	queue := &sqs.QueueWithStats{
		Queue:    tq,
		Scope:    testScope,
		QueueURL: "testUrl",
	}
	handler := mocks.NewMockMessageProcessor()
	When(handler.ProcessMessage(matchers.AnyTypesMessage())).ThenReturn(nil)
	worker := &sqs.Worker{
		Queue:            queue,
		QueueURL:         "testUrl",
		MessageProcessor: handler,
		Logger:           logging.NewNoopCtxLogger(t),
	}

	wg.Add(1)
	go func() {
		worker.Work(ctx)
		wg.Done()
	}()

	// wait for listen to complete or timeout.
	assertCompletes(t, &wg, time.Second)
	Assert(t, testScope.Snapshot().Counters()["test.receive.success+"].Value() == 1, "should have received message")
	Assert(t, testScope.Snapshot().Counters()["test.delete.success+"].Value() == 1, "should have deleted message")
	Assert(t, len(tq.messages) == 0, "should have processed all messages")
	handler.VerifyWasCalledOnce().ProcessMessage(matchers.AnyTypesMessage())
}

func TestWorker_Error(t *testing.T) {
	RegisterMockTestingT(t)
	ctx, cancelFunc := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	testScope := tally.NewTestScope("test", nil)

	queue := &sqs.QueueWithStats{
		Queue: &testQueue{
			receiveError: errors.New("reading messages off of SQS queue"),
			cancel:       cancelFunc,
		},
		Scope:    testScope,
		QueueURL: "foo",
	}
	handler := mocks.NewMockMessageProcessor()
	When(handler.ProcessMessage(matchers.AnyTypesMessage())).ThenReturn(nil)
	worker := &sqs.Worker{
		Queue:            queue,
		QueueURL:         "testUrl",
		MessageProcessor: handler,
		Logger:           logging.NewNoopCtxLogger(t),
	}

	wg.Add(1)
	go func() {
		worker.Work(ctx)
		wg.Done()
	}()

	// wait for listen to complete or timeout.
	assertCompletes(t, &wg, time.Second)
	Assert(t, testScope.Snapshot().Counters()["test.receive.error+"].Value() == 1, "should have not received message")
	handler.VerifyWasCalled(Never()).ProcessMessage(matchers.AnyTypesMessage())
}

func TestWorker_HandlerError(t *testing.T) {
	RegisterMockTestingT(t)
	ctx, cancelFunc := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	testScope := tally.NewTestScope("test", nil)

	expectedMessage := types.Message{
		Body:          aws.String("body"),
		ReceiptHandle: aws.String("receipt_handle"),
		MessageId:     aws.String("message_id"),
	}
	tq := &testQueue{
		messages: []types.Message{
			expectedMessage,
		},
		cancel: cancelFunc,
	}

	queue := &sqs.QueueWithStats{
		Queue:    tq,
		Scope:    testScope,
		QueueURL: "foo",
	}
	handler := mocks.NewMockMessageProcessor()
	When(handler.ProcessMessage(matchers.AnyTypesMessage())).ThenReturn(errors.New("unable to process msg"))
	worker := &sqs.Worker{
		Queue:            queue,
		QueueURL:         "testUrl",
		MessageProcessor: handler,
		Logger:           logging.NewNoopCtxLogger(t),
	}

	wg.Add(1)
	go func() {
		worker.Work(ctx)
		wg.Done()
	}()

	// wait for listen to complete or timeout.
	assertCompletes(t, &wg, time.Second)
	Assert(t, testScope.Snapshot().Counters()["test.receive.success+"].Value() == 1, "should have received message")
	Assert(t, len(tq.messages) == 1, "should have not successfully processed message")
	handler.VerifyWasCalled(Once()).ProcessMessage(matchers.AnyTypesMessage())
}

// assertCompletes places a timeout on a sync.WaitGroup and fails if the
// groups doesn't complete before the timeout occurs
func assertCompletes(t *testing.T, waitGroup *sync.WaitGroup, timeout time.Duration) {
	Assert(t, !timedOut(waitGroup, timeout), "wait group timed out after %s", timeout)
}

func timedOut(waitGroup *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		waitGroup.Wait()
	}()
	select {
	case <-c:
		return false
	case <-time.After(timeout):
		return true
	}
}
