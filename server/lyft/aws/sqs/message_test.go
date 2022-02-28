package sqs_test

import (
	"bytes"
	"net/http"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	. "github.com/petergtz/pegomock"
	controller_mocks "github.com/runatlantis/atlantis/server/controllers/events/mocks"
	"github.com/runatlantis/atlantis/server/controllers/events/mocks/matchers"
	"github.com/runatlantis/atlantis/server/lyft/aws/sqs"
	. "github.com/runatlantis/atlantis/testing"
	"github.com/stretchr/testify/assert"
	"github.com/uber-go/tally"

	"testing"
)

func TestAtlantisMessageHandler_PostSuccess(t *testing.T) {
	RegisterMockTestingT(t)
	testScope := tally.NewTestScope("test", nil)
	req := createExampleRequest(t)
	mockPostHandler := controller_mocks.NewMockVCSPostHandler()
	handler := &sqs.VCSEventMessageProcessorStats{
		VCSEventMessageProcessor: sqs.VCSEventMessageProcessor{
			PostHandler: mockPostHandler,
		},
		Scope: testScope,
	}

	err := handler.ProcessMessage(toSqsMessage(t, req))
	assert.NoError(t, err)
	mockPostHandler.VerifyWasCalledOnce().Post(matchers.AnyHttpResponseWriter(), matchers.AnyPtrToHttpRequest())
	Assert(t, testScope.Snapshot().Counters()["test.success+"].Value() == 1, "message handler was successful")
}

func TestAtlantisMessageHandler_Error(t *testing.T) {
	RegisterMockTestingT(t)
	testScope := tally.NewTestScope("test", nil)
	mockPostHandler := controller_mocks.NewMockVCSPostHandler()
	handler := &sqs.VCSEventMessageProcessorStats{
		VCSEventMessageProcessor: sqs.VCSEventMessageProcessor{
			PostHandler: mockPostHandler,
		},
		Scope: testScope,
	}
	invalidMessage := types.Message{}
	err := handler.ProcessMessage(invalidMessage)
	assert.Error(t, err)
	mockPostHandler.VerifyWasCalled(Never()).Post(matchers.AnyHttpResponseWriter(), matchers.AnyPtrToHttpRequest())
	Assert(t, testScope.Snapshot().Counters()["test.error+"].Value() == 1, "message handler was not successful")
}

func toSqsMessage(t *testing.T, req *http.Request) types.Message {
	buffer := bytes.NewBuffer([]byte{})
	err := req.Write(buffer)
	assert.NoError(t, err)
	return types.Message{
		Body: aws.String(string(buffer.Bytes())),
	}
}

func createExampleRequest(t *testing.T) *http.Request {
	url, err := url.Parse("http://www.atlantis.com")
	assert.NoError(t, err)
	req := &http.Request{
		URL: url,
	}
	return req
}
