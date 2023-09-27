// Copyright 2017 HootSuite Media Inc.
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// Modified hereafter by contributors to runatlantis/atlantis.

package webhooks_test

import (
	"errors"
	"testing"

	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/webhooks"
	"github.com/runatlantis/atlantis/server/events/webhooks/mocks"

	. "github.com/petergtz/pegomock/v4"
	. "github.com/runatlantis/atlantis/testing"
)

var underlying *mocks.MockUnderlyingSlackClient
var client webhooks.DefaultSlackClient
var result webhooks.ApplyResult

func TestAuthTest_Success(t *testing.T) {
	t.Log("When the underlying client succeeds, function should succeed")
	setup(t)
	err := client.AuthTest()
	Ok(t, err)
}

func TestAuthTest_Error(t *testing.T) {
	t.Log("When the underlying slack client errors, an error should be returned")
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

/*
// The next 2 tests are commented out because they currently fail using the Pegamock's
// VerifyWasCalledOnce using variadic parameters.
// See issue https://github.com/petergtz/pegomock/issues/112
func TestPostMessage_Success(t *testing.T) {
	t.Log("When apply succeeds, function should succeed and indicate success")
	setup(t)

	attachments := []slack.Attachment{{
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
			{
				Title: "Directory",
				Value: result.Directory,
				Short: true,
			},
		},
	}}

	channel := "somechannel"
	err := client.PostMessage(channel, result)
	Ok(t, err)
	underlying.VerifyWasCalledOnce().PostMessage(
		channel,
		slack.MsgOptionAsUser(true),
		slack.MsgOptionText("", false),
		slack.MsgOptionAttachments(attachments[0]),
	)

	t.Log("When apply fails, function should succeed and indicate failure")
	result.Success = false
	attachments[0].Color = "danger"
	attachments[0].Text = "Apply failed for <url|runatlantis/atlantis>"

	err = client.PostMessage(channel, result)
	Ok(t, err)
	underlying.VerifyWasCalledOnce().PostMessage(
		channel,
		slack.MsgOptionAsUser(true),
		slack.MsgOptionText("", false),
		slack.MsgOptionAttachments(attachments[0]),
	)
}

func TestPostMessage_Error(t *testing.T) {
	t.Log("When the underlying slack client errors, an error should be returned")
	setup(t)

	attachments := []slack.Attachment{{
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
			{
				Title: "Directory",
				Value: result.Directory,
				Short: true,
			},
		},
	}}

	channel := "somechannel"
	When(underlying.PostMessage(
		channel,
		slack.MsgOptionAsUser(true),
		slack.MsgOptionText("", false),
		slack.MsgOptionAttachments(attachments[0]),
	)).ThenReturn("", "", errors.New(""))

	err := client.PostMessage(channel, result)
	Assert(t, err != nil, "expected error")
}
*/

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
			Num:        1,
			URL:        "url",
			BaseBranch: "main",
		},
		User: models.User{
			Username: "lkysow",
		},
		Success: true,
	}
}
