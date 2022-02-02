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

import (
	"fmt"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events/models"
)

var Pull = models.PullRequest{
	Num:        1,
	HeadCommit: "16ca62f65c18ff456c6ef4cacc8d4826e264bb17",
	HeadBranch: "branch",
	Author:     "lkysow",
	URL:        "url",
}

var GithubRepo = models.Repo{
	CloneURL:          "https://user:password@github.com/runatlantis/atlantis.git",
	FullName:          "runatlantis/atlantis",
	Owner:             "runatlantis",
	SanitizedCloneURL: "https://github.com/runatlantis/atlantis.git",
	Name:              "atlantis",
	VCSHost: models.VCSHost{
		Hostname: "github.com",
		Type:     models.Github,
	},
}

var GitlabRepo = models.Repo{
	CloneURL:          "https://user:password@github.com/runatlantis/atlantis.git",
	FullName:          "runatlantis/atlantis",
	Owner:             "runatlantis",
	SanitizedCloneURL: "https://gitlab.com/runatlantis/atlantis.git",
	Name:              "atlantis",
	VCSHost: models.VCSHost{
		Hostname: "gitlab.com",
		Type:     models.Gitlab,
	},
}

var User = models.User{
	Username: "lkysow",
}

var projectName = "test-project"

var Project = valid.Project{
	Name: &projectName,
}

var PullInfo = fmt.Sprintf("%s/%d/%s", GithubRepo.FullName, Pull.Num, *Project.Name)
