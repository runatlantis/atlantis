// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package webhooks_test

import (
	"regexp"
	"testing"

	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/webhooks"
	"github.com/runatlantis/atlantis/server/events/webhooks/mocks"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
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
		BranchRegex:    regex,
		ProjectRegex:   regex,
		DirectoryRegex: regex,
		Channel:        channel,
	}
	result := webhooks.ApplyResult{
		Workspace: "production",
		Pull: models.PullRequest{
			BaseBranch: "main",
		},
	}

	t.Log("PostMessage should be called, doesn't matter if it errors or not")
	_ = hook.Send(logging.NewNoopLogger(t), result)
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
		BranchRegex:    regex,
		ProjectRegex:   regex,
		DirectoryRegex: regex,
		Channel:        channel,
	}
	result := webhooks.ApplyResult{
		Workspace: "production",
		Pull: models.PullRequest{
			BaseBranch: "main",
		},
	}
	err = hook.Send(logging.NewNoopLogger(t), result)
	Ok(t, err)
	client.VerifyWasCalled(Never()).PostMessage(channel, result)
}
