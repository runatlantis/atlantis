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

package fixtures

import "github.com/google/go-github/github"

var PullEvent = github.PullRequestEvent{
	Sender: &github.User{
		Login: github.String("user"),
	},
	Repo:        &Repo,
	PullRequest: &Pull,
	Action:      github.String("opened"),
}

var Pull = github.PullRequest{
	Head: &github.PullRequestBranch{
		SHA:  github.String("sha256"),
		Ref:  github.String("ref"),
		Repo: &Repo,
	},
	Base: &github.PullRequestBranch{
		SHA:  github.String("sha256"),
		Repo: &Repo,
		Ref:  github.String("basebranch"),
	},
	HTMLURL: github.String("html-url"),
	User: &github.User{
		Login: github.String("user"),
	},
	Number: github.Int(1),
	State:  github.String("open"),
}

var Repo = github.Repository{
	FullName: github.String("owner/repo"),
	Owner:    &github.User{Login: github.String("owner")},
	Name:     github.String("repo"),
	CloneURL: github.String("https://github.com/owner/repo.git"),
}
