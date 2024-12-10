// Copyright 2024 Florian Beisel
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
