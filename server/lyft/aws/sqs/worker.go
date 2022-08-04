package sqs

import (
	"context"
	"sync"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/uber-go/tally/v4"
)

const (
	ProcessMessageMetricName = "process"
	ReceiveMessageMetricName = "receive"
	DeleteMessageMetricName  = "delete"

	Latency = "latency"
	Success = "success"
	Error   = "error"
)

type Worker struct {
	Queue            Queue
	QueueURL         string
	MessageProcessor MessageProcessor
	Logger           logging.Logger
}

func NewGatewaySQSWorker(ctx context.Context, scope tally.Scope, logger logging.Logger, queueURL string, postHandler VCSPostHandler) (*Worker, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "error loading aws config for sqs worker")
	}
	scope = scope.SubScope("aws.sqs.msg")
	sqsQueueWrapper := &QueueWithStats{
		Queue:    sqs.NewFromConfig(cfg),
		Scope:    scope,
		QueueURL: queueURL,
	}

	handler := &VCSEventMessageProcessorStats{
		VCSEventMessageProcessor: VCSEventMessageProcessor{
			PostHandler: postHandler,
		},
		Scope: scope.SubScope(ProcessMessageMetricName),
	}

	return &Worker{
		Queue:            sqsQueueWrapper,
		QueueURL:         queueURL,
		MessageProcessor: handler,
		Logger:           logger,
	}, nil
}

func (w *Worker) Work(ctx context.Context) {
	messages := make(chan types.Message)
	// Used to synchronize stopping message retrieval and processing
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		w.Logger.InfoContext(ctx, "start processing sqs messages")
		w.processMessage(ctx, messages)
	}()
	request := &sqs.ReceiveMessageInput{
		QueueUrl:            &w.QueueURL,
		MaxNumberOfMessages: 10, //max number of batch-able messages
		WaitTimeSeconds:     20, //max duration long polling
	}
	w.Logger.InfoContext(ctx, "start receiving sqs messages")
	w.receiveMessages(ctx, messages, request)
	wg.Wait()
}

func (w *Worker) receiveMessages(ctx context.Context, messages chan types.Message, request *sqs.ReceiveMessageInput) {
	for {
		select {
		case <-ctx.Done():
			close(messages)
			w.Logger.InfoContext(ctx, "closed sqs messages channel")
			return
		default:
			response, err := w.Queue.ReceiveMessage(ctx, request)
			if err != nil {
				w.Logger.WarnContext(ctx, "unable to receive sqs message", map[string]interface{}{"err": err})
				continue
			}
			for _, message := range response.Messages {
				messages <- message
			}
		}
	}
}

func (w *Worker) processMessage(ctx context.Context, messages chan types.Message) {
	// VisibilityTimeout is 30s, ideally enough time to "processMessage" < 10 messages (i.e. spin up goroutine for each)
	for message := range messages {
		err := w.MessageProcessor.ProcessMessage(message)
		if err != nil {
			w.Logger.ErrorContext(ctx, "unable to process sqs message", map[string]interface{}{"err": err})
			continue
		}

		// Since we've successfully processed the message, let's go ahead and delete it from the queue
		_, err = w.Queue.DeleteMessage(ctx, &sqs.DeleteMessageInput{
			QueueUrl:      &w.QueueURL,
			ReceiptHandle: message.ReceiptHandle,
		})
		if err != nil {
			w.Logger.WarnContext(ctx, "unable to delete processed sqs message", map[string]interface{}{"err": err})
		}
	}
}
