package sqs

import (
	"bufio"
	"bytes"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/pkg/errors"
	"github.com/uber-go/tally/v4"
)

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_sqs_message_handler.go MessageProcessor
type MessageProcessor interface {
	ProcessMessage(types.Message) error
}

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_vcs_post_handler.go VCSPostHandler
type VCSPostHandler interface {
	Post(w http.ResponseWriter, r *http.Request)
}

type VCSEventMessageProcessor struct {
	PostHandler VCSPostHandler
}

func (p *VCSEventMessageProcessor) ProcessMessage(msg types.Message) error {
	if msg.Body == nil {
		return errors.New("message received from sqs has no body")
	}

	buffer := bytes.NewBufferString(*msg.Body)
	buf := bufio.NewReader(buffer)
	req, err := http.ReadRequest(buf)
	if err != nil {
		return errors.Wrap(err, "reading bytes from sqs into http request")
	}

	// using a no-op writer since we shouldn't send response back in worker mode
	p.PostHandler.Post(&NoOpResponseWriter{}, req)
	return nil
}

type VCSEventMessageProcessorStats struct {
	Scope tally.Scope
	VCSEventMessageProcessor
}

func (s *VCSEventMessageProcessorStats) ProcessMessage(msg types.Message) error {
	successCount := s.Scope.Counter(Success)
	errorCount := s.Scope.Counter(Error)

	timer := s.Scope.Timer(Latency)
	span := timer.Start()
	defer span.Stop()

	if err := s.VCSEventMessageProcessor.ProcessMessage(msg); err != nil {
		errorCount.Inc(1)
		return err
	}
	successCount.Inc(1)
	return nil
}

type NoOpResponseWriter struct{}

func (n *NoOpResponseWriter) Header() http.Header {
	return nil
}

func (n *NoOpResponseWriter) Write([]byte) (int, error) {
	return 0, nil
}

func (n *NoOpResponseWriter) WriteHeader(statusCode int) {}
