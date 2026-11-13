// Copyright 2024 Florian Beisel
// SPDX-License-Identifier: Apache-2.0

package gitea

import "code.gitea.io/sdk/gitea"

type GiteaWebhookPayload struct {
	Action      string            `json:"action"`
	Number      int               `json:"number"`
	PullRequest gitea.PullRequest `json:"pull_request"`
}

type GiteaIssueCommentPayload struct {
	Action     string           `json:"action"`
	Comment    gitea.Comment    `json:"comment"`
	Repository gitea.Repository `json:"repository"`
	Issue      gitea.Issue      `json:"issue"`
}
