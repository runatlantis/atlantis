package webhooks_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/nlopes/slack"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/webhooks"
	"github.com/runatlantis/atlantis/server/events/webhooks/mocks"

	. "github.com/petergtz/pegomock"
	. "github.com/runatlantis/atlantis/testing"
)

var underlying *mocks.MockUnderlyingSlackClient
var client webhooks.DefaultSlackClient
var result webhooks.ApplyResult

func TestAuthTest_Success(t *testing.T) {
	t.Log("When the underylying client suceeds, function should succeed")
	setup(t)
	err := client.AuthTest()
	Ok(t, err)
}

func TestAuthTest_Error(t *testing.T) {
	t.Log("When the underylying slack client errors, an error should be returned")
	setup(t)
	When(underlying.AuthTest()).ThenReturn(nil, errors.New(""))
	err := client.AuthTest()
	Assert(t, err != nil, "expected error")
}

func TestTokenIsSet(t *testing.T) {
	t.Log("When the Token is an empty string, function should return false")
	c := webhooks.DefaultSlackClient{
		Token: "",
	}
	Equals(t, false, c.TokenIsSet())

	t.Log("When the Token is not an empty string, function should return true")
	c.Token = "random"
	Equals(t, true, c.TokenIsSet())
}

func TestChannelExists_False(t *testing.T) {
	t.Log("When the slack channel doesn't exist, function should return false")
	setup(t)
	When(underlying.GetChannels(true)).ThenReturn(nil, nil)
	exists, err := client.ChannelExists("somechannel")
	Ok(t, err)
	Equals(t, false, exists)
}

func TestChannelExists_True(t *testing.T) {
	t.Log("When the slack channel exists, function should return true")
	setup(t)
	channelJSON := `{"name":"existingchannel"}`
	var channel slack.Channel
	err := json.Unmarshal([]byte(channelJSON), &channel)
	Ok(t, err)
	When(underlying.GetChannels(true)).ThenReturn([]slack.Channel{channel}, nil)

	exists, err := client.ChannelExists("existingchannel")
	Ok(t, err)
	Equals(t, true, exists)
}

func TestChannelExists_Error(t *testing.T) {
	t.Log("When the underylying slack client errors, an error should be returned")
	setup(t)
	When(underlying.GetChannels(true)).ThenReturn(nil, errors.New(""))

	_, err := client.ChannelExists("anychannel")
	Assert(t, err != nil, "expected error")
}

func TestPostMessage_Success(t *testing.T) {
	t.Log("When apply succeds, function should succeed and indicate success")
	setup(t)

	expParams := slack.NewPostMessageParameters()
	expParams.Attachments = []slack.Attachment{{
		Color: "good",
		Text:  "Apply succeeded for <url|runatlantis/atlantis>",
		Fields: []slack.AttachmentField{
			{
				Title: "Workspace",
				Value: result.Workspace,
				Short: true,
			},
			{
				Title: "User",
				Value: result.User.Username,
				Short: true,
			},
		},
	}}
	expParams.AsUser = false
	expParams.EscapeText = false

	channel := "somechannel"
	err := client.PostMessage(channel, result)
	Ok(t, err)
	underlying.VerifyWasCalledOnce().PostMessage(channel, "", expParams)

	t.Log("When apply fails, function should succeed and indicate failure")
	result.Success = false
	expParams.Attachments[0].Color = "danger"
	expParams.Attachments[0].Text = "Apply failed for <url|runatlantis/atlantis>"

	err = client.PostMessage(channel, result)
	Ok(t, err)
	underlying.VerifyWasCalledOnce().PostMessage(channel, "", expParams)
}

func TestPostMessage_Error(t *testing.T) {
	t.Log("When the underylying slack client errors, an error should be returned")
	setup(t)

	expParams := slack.NewPostMessageParameters()
	expParams.Attachments = []slack.Attachment{{
		Color: "good",
		Text:  "Apply succeeded for <url|runatlantis/atlantis>",
		Fields: []slack.AttachmentField{
			{
				Title: "Workspace",
				Value: result.Workspace,
				Short: true,
			},
			{
				Title: "User",
				Value: result.User.Username,
				Short: true,
			},
		},
	}}
	expParams.AsUser = false
	expParams.EscapeText = false

	channel := "somechannel"
	When(underlying.PostMessage(channel, "", expParams)).ThenReturn("", "", errors.New(""))

	err := client.PostMessage(channel, result)
	Assert(t, err != nil, "expected error")
}

func setup(t *testing.T) {
	RegisterMockTestingT(t)
	underlying = mocks.NewMockUnderlyingSlackClient()
	client = webhooks.DefaultSlackClient{
		Slack: underlying,
		Token: "sometoken",
	}
	result = webhooks.ApplyResult{
		Workspace: "production",
		Repo: models.Repo{
			FullName: "runatlantis/atlantis",
		},
		Pull: models.PullRequest{
			Num: 1,
			URL: "url",
		},
		User: models.User{
			Username: "lkysow",
		},
		Success: true,
	}
}
