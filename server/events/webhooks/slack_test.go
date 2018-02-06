package webhooks_test

import (
	"regexp"
	"testing"

	"github.com/atlantisnorth/atlantis/server/events/webhooks"
	"github.com/atlantisnorth/atlantis/server/events/webhooks/mocks"
	"github.com/atlantisnorth/atlantis/server/logging"
	. "github.com/atlantisnorth/atlantis/testing"
	. "github.com/petergtz/pegomock"
)

func TestSend_PostMessage(t *testing.T) {
	t.Log("Sending a hook with a matching regex should call PostMessage")
	RegisterMockTestingT(t)
	client := mocks.NewMockSlackClient()
	regex, err := regexp.Compile(".*")
	Ok(t, err)

	channel := "somechannel"
	hook := webhooks.SlackWebhook{
		Client:         client,
		WorkspaceRegex: regex,
		Channel:        channel,
	}
	result := webhooks.ApplyResult{
		Workspace: "production",
	}

	t.Log("PostMessage should be called, doesn't matter if it errors or not")
	_ = hook.Send(logging.NewNoopLogger(), result)
	client.VerifyWasCalledOnce().PostMessage(channel, result)
}

func TestSend_NoopSuccess(t *testing.T) {
	t.Log("Sending a hook with a non-matching regex should succeed")
	RegisterMockTestingT(t)
	client := mocks.NewMockSlackClient()
	regex, err := regexp.Compile("weirdemv")
	Ok(t, err)

	channel := "somechannel"
	hook := webhooks.SlackWebhook{
		Client:         client,
		WorkspaceRegex: regex,
		Channel:        channel,
	}
	result := webhooks.ApplyResult{
		Workspace: "production",
	}
	err = hook.Send(logging.NewNoopLogger(), result)
	Ok(t, err)
	client.VerifyWasCalled(Never()).PostMessage(channel, result)
}
