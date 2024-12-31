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

package testdata

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/go-github/v66/github"
	"github.com/mcdafydd/go-azuredevops/azuredevops"
)

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

var ADPullEvent = azuredevops.Event{
	EventType: "git.pullrequest.created",
	Resource:  &ADPull,
}

var ADPullUpdatedEvent = azuredevops.Event{
	EventType: "git.pullrequest.updated",
	Resource:  &ADPull,
}

var ADPullClosedEvent = azuredevops.Event{
	EventType: "git.pullrequest.merged",
	Resource:  &ADPullCompleted,
}

var ADPull = azuredevops.GitPullRequest{
	CreatedBy: &azuredevops.IdentityRef{
		ID:          azuredevops.String("d6245f20-2af8-44f4-9451-8107cb2767db"),
		DisplayName: azuredevops.String("User"),
		UniqueName:  azuredevops.String("user@example.com"),
	},
	LastMergeSourceCommit: &azuredevops.GitCommitRef{
		CommitID: azuredevops.String("b60280bc6e62e2f880f1b63c1e24987664d3bda3"),
		URL:      azuredevops.String("https://dev.azure.com/owner/_apis/git/repositories/3411ebc1-d5aa-464f-9615-0b527bc66719/commits/b60280bc6e62e2f880f1b63c1e24987664d3bda3"),
	},
	PullRequestID: azuredevops.Int(1),
	Repository:    &ADRepo,
	SourceRefName: azuredevops.String("refs/heads/feature/sourceBranch"),
	Status:        azuredevops.String("active"),
	TargetRefName: azuredevops.String("refs/heads/targetBranch"),
	URL:           azuredevops.String("https://dev.azure.com/fabrikam/_apis/git/repositories/3411ebc1-d5aa-464f-9615-0b527bc66719/pullRequests/21"),
}

var ADPullCompleted = azuredevops.GitPullRequest{
	CreatedBy: &azuredevops.IdentityRef{
		ID:          azuredevops.String("d6245f20-2af8-44f4-9451-8107cb2767db"),
		DisplayName: azuredevops.String("User"),
		UniqueName:  azuredevops.String("user@example.com"),
	},
	LastMergeSourceCommit: &azuredevops.GitCommitRef{
		CommitID: azuredevops.String("b60280bc6e62e2f880f1b63c1e24987664d3bda3"),
		URL:      azuredevops.String("https://dev.azure.com/owner/_apis/git/repositories/3411ebc1-d5aa-464f-9615-0b527bc66719/commits/b60280bc6e62e2f880f1b63c1e24987664d3bda3"),
	},
	PullRequestID: azuredevops.Int(1),
	Repository:    &ADRepo,
	SourceRefName: azuredevops.String("refs/heads/owner/sourceBranch"),
	Status:        azuredevops.String("completed"),
	TargetRefName: azuredevops.String("refs/heads/targetBranch"),
	URL:           azuredevops.String("https://dev.azure.com/fabrikam/_apis/git/repositories/3411ebc1-d5aa-464f-9615-0b527bc66719/pullRequests/21"),
}

var ADRepo = azuredevops.GitRepository{
	DefaultBranch: azuredevops.String("refs/heads/main"),
	Name:          azuredevops.String("repo"),
	ParentRepository: &azuredevops.GitRepositoryRef{
		Name: azuredevops.String("owner"),
	},
	Project: &azuredevops.TeamProjectReference{
		ID:    azuredevops.String("a21f5f20-4a12-aaf4-ab12-9a0927cbbb90"),
		Name:  azuredevops.String("project"),
		State: azuredevops.String("unchanged"),
	},
	WebURL: azuredevops.String("https://dev.azure.com/owner/project/_git/repo"),
}

var ADPullJSON = `{
	"repository": {
		"id": "3411ebc1-d5aa-464f-9615-0b527bc66719",
		"name": "repo",
		"url": "https://dev.azure.com/owner/project/_apis/git/repositories/3411ebc1-d5aa-464f-9615-0b527bc66719",
		"webUrl": "https://dev.azure.com/owner/project/_apis/git/repositories/3411ebc1-d5aa-464f-9615-0b527bc66719",
		"project": {
			"id": "a7573007-bbb3-4341-b726-0c4148a07853",
			"name": "project",
			"description": "test project created on Halloween 2016",
			"url": "https://dev.azure.com/owner/_apis/projects/a7573007-bbb3-4341-b726-0c4148a07853",
			"state": "wellFormed",
			"revision": 7
		},
		"remoteUrl": "https://dev.azure.com/owner/project/_git/repo"
	},
	"pullRequestId": 22,
	"codeReviewId": 22,
	"status": "active",
	"createdBy": {
		"id": "d6245f20-2af8-44f4-9451-8107cb2767db",
		"displayName": "Normal Paulk",
		"uniqueName": "fabrikamfiber16@hotmail.com",
		"url": "https://dev.azure.com/owner/_apis/Identities/d6245f20-2af8-44f4-9451-8107cb2767db",
		"imageUrl": "https://dev.azure.com/owner/_api/_common/identityImage?id=d6245f20-2af8-44f4-9451-8107cb2767db"
	},
	"creationDate": "2016-11-01T16:30:31.6655471Z",
	"title": "A new feature",
	"description": "Adding a new feature",
	"sourceRefName": "refs/heads/npaulk/my_work",
	"targetRefName": "refs/heads/new_feature",
	"mergeStatus": "succeeded",
	"mergeId": "f5fc8381-3fb2-49fe-8a0d-27dcc2d6ef82",
	"lastMergeSourceCommit": {
		"commitId": "b60280bc6e62e2f880f1b63c1e24987664d3bda3",
		"url": "https://dev.azure.com/owner/_apis/git/repositories/3411ebc1-d5aa-464f-9615-0b527bc66719/commits/b60280bc6e62e2f880f1b63c1e24987664d3bda3"
	},
	"lastMergeTargetCommit": {
		"commitId": "f47bbc106853afe3c1b07a81754bce5f4b8dbf62",
		"url": "https://dev.azure.com/owner/_apis/git/repositories/3411ebc1-d5aa-464f-9615-0b527bc66719/commits/f47bbc106853afe3c1b07a81754bce5f4b8dbf62"
	},
	"lastMergeCommit": {
		"commitId": "39f52d24533cc712fc845ed9fd1b6c06b3942588",
		"author": {
			"name": "Normal Paulk",
			"email": "fabrikamfiber16@hotmail.com",
			"date": "2016-11-01T16:30:32Z"
		},
		"committer": {
			"name": "Normal Paulk",
			"email": "fabrikamfiber16@hotmail.com",
			"date": "2016-11-01T16:30:32Z"
		},
		"comment": "Merge pull request 22 from npaulk/my_work into new_feature",
		"url": "https://dev.azure.com/owner/_apis/git/repositories/3411ebc1-d5aa-464f-9615-0b527bc66719/commits/39f52d24533cc712fc845ed9fd1b6c06b3942588"
	},
	"reviewers": [
		{
			"reviewerUrl": "https://dev.azure.com/owner/_apis/git/repositories/3411ebc1-d5aa-464f-9615-0b527bc66719/pullRequests/22/reviewers/d6245f20-2af8-44f4-9451-8107cb2767db",
			"vote": 0,
			"id": "d6245f20-2af8-44f4-9451-8107cb2767db",
			"displayName": "Normal Paulk",
			"uniqueName": "fabrikamfiber16@hotmail.com",
			"url": "https://dev.azure.com/owner/_apis/Identities/d6245f20-2af8-44f4-9451-8107cb2767db",
			"imageUrl": "https://dev.azure.com/owner/_api/_common/identityImage?id=d6245f20-2af8-44f4-9451-8107cb2767db"
		}
	],
	"url": "https://dev.azure.com/owner/_apis/git/repositories/3411ebc1-d5aa-464f-9615-0b527bc66719/pullRequests/22",
	"_links": {
		"self": {
			"href": "https://dev.azure.com/owner/_apis/git/repositories/3411ebc1-d5aa-464f-9615-0b527bc66719/pullRequests/22"
		},
		"repository": {
			"href": "https://dev.azure.com/owner/_apis/git/repositories/3411ebc1-d5aa-464f-9615-0b527bc66719"
		},
		"workItems": {
			"href": "https://dev.azure.com/owner/_apis/git/repositories/3411ebc1-d5aa-464f-9615-0b527bc66719/pullRequests/22/workitems"
		},
		"sourceBranch": {
			"href": "https://dev.azure.com/owner/_apis/git/repositories/3411ebc1-d5aa-464f-9615-0b527bc66719/refs"
		},
		"targetBranch": {
			"href": "https://dev.azure.com/owner/_apis/git/repositories/3411ebc1-d5aa-464f-9615-0b527bc66719/refs"
		},
		"sourceCommit": {
			"href": "https://dev.azure.com/owner/_apis/git/repositories/3411ebc1-d5aa-464f-9615-0b527bc66719/commits/b60280bc6e62e2f880f1b63c1e24987664d3bda3"
		},
		"targetCommit": {
			"href": "https://dev.azure.com/owner/_apis/git/repositories/3411ebc1-d5aa-464f-9615-0b527bc66719/commits/f47bbc106853afe3c1b07a81754bce5f4b8dbf62"
		},
		"createdBy": {
			"href": "https://dev.azure.com/owner/_apis/Identities/d6245f20-2af8-44f4-9451-8107cb2767db"
		},
		"iterations": {
			"href": "https://dev.azure.com/owner/_apis/git/repositories/3411ebc1-d5aa-464f-9615-0b527bc66719/pullRequests/22/iterations"
		}
	},
	"supportsIterations": true,
	"artifactId": "vstfs:///Git/PullRequestId/a7573007-bbb3-4341-b726-0c4148a07853%2f3411ebc1-d5aa-464f-9615-0b527bc66719%2f22"
}`

var ADSelfPullEvent = azuredevops.Event{
	EventType: "git.pullrequest.created",
	Resource:  &ADSelfPull,
}

var ADSelfPullUpdatedEvent = azuredevops.Event{
	EventType: "git.pullrequest.updated",
	Resource:  &ADSelfPull,
}

var ADSelfPullClosedEvent = azuredevops.Event{
	EventType: "git.pullrequest.merged",
	Resource:  &ADSelfPullCompleted,
}

var ADSelfPull = azuredevops.GitPullRequest{
	CreatedBy: &azuredevops.IdentityRef{
		ID:          azuredevops.String("d6245f20-2af8-44f4-9451-8107cb2767db"),
		DisplayName: azuredevops.String("User"),
		UniqueName:  azuredevops.String("user@example.com"),
	},
	LastMergeSourceCommit: &azuredevops.GitCommitRef{
		CommitID: azuredevops.String("b60280bc6e62e2f880f1b63c1e24987664d3bda3"),
		URL:      azuredevops.String("https://devops.abc.com/owner/project/_apis/git/repositories/3411ebc1-d5aa-464f-9615-0b527bc66719/commits/b60280bc6e62e2f880f1b63c1e24987664d3bda3"),
	},
	PullRequestID: azuredevops.Int(1),
	Repository:    &ADSelfRepo,
	SourceRefName: azuredevops.String("refs/heads/feature/sourceBranch"),
	Status:        azuredevops.String("active"),
	TargetRefName: azuredevops.String("refs/heads/targetBranch"),
	URL:           azuredevops.String("https://devops.abc.com/owner/project/_apis/git/repositories/3411ebc1-d5aa-464f-9615-0b527bc66719/pullRequests/21"),
}

var ADSelfPullCompleted = azuredevops.GitPullRequest{
	CreatedBy: &azuredevops.IdentityRef{
		ID:          azuredevops.String("d6245f20-2af8-44f4-9451-8107cb2767db"),
		DisplayName: azuredevops.String("User"),
		UniqueName:  azuredevops.String("user@example.com"),
	},
	LastMergeSourceCommit: &azuredevops.GitCommitRef{
		CommitID: azuredevops.String("b60280bc6e62e2f880f1b63c1e24987664d3bda3"),
		URL:      azuredevops.String("https://https://devops.abc.com/owner/project/_apis/git/repositories/3411ebc1-d5aa-464f-9615-0b527bc66719/commits/b60280bc6e62e2f880f1b63c1e24987664d3bda3"),
	},
	PullRequestID: azuredevops.Int(1),
	Repository:    &ADSelfRepo,
	SourceRefName: azuredevops.String("refs/heads/owner/sourceBranch"),
	Status:        azuredevops.String("completed"),
	TargetRefName: azuredevops.String("refs/heads/targetBranch"),
	URL:           azuredevops.String("https://devops.abc.com/owner/project/_apis/git/repositories/3411ebc1-d5aa-464f-9615-0b527bc66719/pullRequests/21"),
}

var ADSelfRepo = azuredevops.GitRepository{
	DefaultBranch: azuredevops.String("refs/heads/main"),
	Name:          azuredevops.String("repo"),
	ParentRepository: &azuredevops.GitRepositoryRef{
		Name: azuredevops.String("owner"),
	},
	Project: &azuredevops.TeamProjectReference{
		ID:    azuredevops.String("a21f5f20-4a12-aaf4-ab12-9a0927cbbb90"),
		Name:  azuredevops.String("project"),
		State: azuredevops.String("unchanged"),
	},
	WebURL: azuredevops.String("https://devops.abc.com/owner/project/_git/repo"),
}

var ADSelfPullJSON = `{
	"repository": {
		"id": "3411ebc1-d5aa-464f-9615-0b527bc66719",
		"name": "repo",
		"url": "https://devops.abc.com/owner/project/_apis/git/repositories/3411ebc1-d5aa-464f-9615-0b527bc66719",
		"webUrl": "https://devops.abc.com/owner/project/_apis/git/repositories/3411ebc1-d5aa-464f-9615-0b527bc66719",
		"project": {
			"id": "a7573007-bbb3-4341-b726-0c4148a07853",
			"name": "project",
			"description": "test project created on Halloween 2016",
			"url": "https://dev.azure.com/owner/_apis/projects/a7573007-bbb3-4341-b726-0c4148a07853",
			"state": "wellFormed",
			"revision": 7
		},
		"remoteUrl": "https://devops.abc.com/owner/project/_git/repo"
	},
	"pullRequestId": 22,
	"codeReviewId": 22,
	"status": "active",
	"createdBy": {
		"id": "d6245f20-2af8-44f4-9451-8107cb2767db",
		"displayName": "Normal Paulk",
		"uniqueName": "fabrikamfiber16@hotmail.com",
		"url": "https://dev.azure.com/owner/_apis/Identities/d6245f20-2af8-44f4-9451-8107cb2767db",
		"imageUrl": "https://dev.azure.com/owner/_api/_common/identityImage?id=d6245f20-2af8-44f4-9451-8107cb2767db"
	},
	"creationDate": "2016-11-01T16:30:31.6655471Z",
	"title": "A new feature",
	"description": "Adding a new feature",
	"sourceRefName": "refs/heads/npaulk/my_work",
	"targetRefName": "refs/heads/new_feature",
	"mergeStatus": "succeeded",
	"mergeId": "f5fc8381-3fb2-49fe-8a0d-27dcc2d6ef82",
	"lastMergeSourceCommit": {
		"commitId": "b60280bc6e62e2f880f1b63c1e24987664d3bda3",
		"url": "https://dev.azure.com/owner/_apis/git/repositories/3411ebc1-d5aa-464f-9615-0b527bc66719/commits/b60280bc6e62e2f880f1b63c1e24987664d3bda3"
	},
	"lastMergeTargetCommit": {
		"commitId": "f47bbc106853afe3c1b07a81754bce5f4b8dbf62",
		"url": "https://dev.azure.com/owner/_apis/git/repositories/3411ebc1-d5aa-464f-9615-0b527bc66719/commits/f47bbc106853afe3c1b07a81754bce5f4b8dbf62"
	},
	"lastMergeCommit": {
		"commitId": "39f52d24533cc712fc845ed9fd1b6c06b3942588",
		"author": {
			"name": "Normal Paulk",
			"email": "fabrikamfiber16@hotmail.com",
			"date": "2016-11-01T16:30:32Z"
		},
		"committer": {
			"name": "Normal Paulk",
			"email": "fabrikamfiber16@hotmail.com",
			"date": "2016-11-01T16:30:32Z"
		},
		"comment": "Merge pull request 22 from npaulk/my_work into new_feature",
		"url": "https://dev.azure.com/owner/_apis/git/repositories/3411ebc1-d5aa-464f-9615-0b527bc66719/commits/39f52d24533cc712fc845ed9fd1b6c06b3942588"
	},
	"reviewers": [
		{
			"reviewerUrl": "https://dev.azure.com/owner/_apis/git/repositories/3411ebc1-d5aa-464f-9615-0b527bc66719/pullRequests/22/reviewers/d6245f20-2af8-44f4-9451-8107cb2767db",
			"vote": 0,
			"id": "d6245f20-2af8-44f4-9451-8107cb2767db",
			"displayName": "Normal Paulk",
			"uniqueName": "fabrikamfiber16@hotmail.com",
			"url": "https://dev.azure.com/owner/_apis/Identities/d6245f20-2af8-44f4-9451-8107cb2767db",
			"imageUrl": "https://dev.azure.com/owner/_api/_common/identityImage?id=d6245f20-2af8-44f4-9451-8107cb2767db"
		}
	],
	"url": "https://dev.azure.com/owner/_apis/git/repositories/3411ebc1-d5aa-464f-9615-0b527bc66719/pullRequests/22",
	"_links": {
		"self": {
			"href": "https://dev.azure.com/owner/_apis/git/repositories/3411ebc1-d5aa-464f-9615-0b527bc66719/pullRequests/22"
		},
		"repository": {
			"href": "https://dev.azure.com/owner/_apis/git/repositories/3411ebc1-d5aa-464f-9615-0b527bc66719"
		},
		"workItems": {
			"href": "https://dev.azure.com/owner/_apis/git/repositories/3411ebc1-d5aa-464f-9615-0b527bc66719/pullRequests/22/workitems"
		},
		"sourceBranch": {
			"href": "https://dev.azure.com/owner/_apis/git/repositories/3411ebc1-d5aa-464f-9615-0b527bc66719/refs"
		},
		"targetBranch": {
			"href": "https://dev.azure.com/owner/_apis/git/repositories/3411ebc1-d5aa-464f-9615-0b527bc66719/refs"
		},
		"sourceCommit": {
			"href": "https://dev.azure.com/owner/_apis/git/repositories/3411ebc1-d5aa-464f-9615-0b527bc66719/commits/b60280bc6e62e2f880f1b63c1e24987664d3bda3"
		},
		"targetCommit": {
			"href": "https://dev.azure.com/owner/_apis/git/repositories/3411ebc1-d5aa-464f-9615-0b527bc66719/commits/f47bbc106853afe3c1b07a81754bce5f4b8dbf62"
		},
		"createdBy": {
			"href": "https://dev.azure.com/owner/_apis/Identities/d6245f20-2af8-44f4-9451-8107cb2767db"
		},
		"iterations": {
			"href": "https://dev.azure.com/owner/_apis/git/repositories/3411ebc1-d5aa-464f-9615-0b527bc66719/pullRequests/22/iterations"
		}
	},
	"supportsIterations": true,
	"artifactId": "vstfs:///Git/PullRequestId/a7573007-bbb3-4341-b726-0c4148a07853%2f3411ebc1-d5aa-464f-9615-0b527bc66719%2f22"
}`

const GithubPrivateKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAuEPzOUE+kiEH1WLiMeBytTEF856j0hOVcSUSUkZxKvqczkWM
9vo1gDyC7ZXhdH9fKh32aapba3RSsp4ke+giSmYTk2mGR538ShSDxh0OgpJmjiKP
X0Bj4j5sFqfXuCtl9SkH4iueivv4R53ktqM+n6hk98l6hRwC39GVIblAh2lEM4L/
6WvYwuQXPMM5OG2Ryh2tDZ1WS5RKfgq+9ksNJ5Q9UtqtqHkO+E63N5OK9sbzpUUm
oNaOl3udTlZD3A8iqwMPVxH4SxgATBPAc+bmjk6BMJ0qIzDcVGTrqrzUiywCTLma
szdk8GjzXtPDmuBgNn+o6s02qVGpyydgEuqmTQIDAQABAoIBACL6AvkjQVVLn8kJ
dBYznJJ4M8ECo+YEgaFwgAHODT0zRQCCgzd+Vxl4YwHmKV2Lr+y2s0drZt8GvYva
KOK8NYYZyi15IlwFyRXmvvykF1UBpSXluYFDH7KaVroWMgRreHcIys5LqVSIb6Bo
gDmK0yBLPp8qR29s2b7ScZRtLaqGJiX+j55rNzrZwxHkxFHyG9OG+u9IsBElcKCP
kYCVE8ZdYexfnKOZbgn2kZB9qu0T/Mdvki8yk3I2bI6xYO24oQmhnT36qnqWoCBX
NuCNsBQgpYZeZET8mEAUmo9d+ABmIHIvSs005agK8xRaP4+6jYgy6WwoejJRF5yd
NBuF7aECgYEA50nZ4FiZYV0vcJDxFYeY3kYOvVuKn8OyW+2rg7JIQTremIjv8FkE
ZnwuF9ZRxgqLxUIfKKfzp/5l5LrycNoj2YKfHKnRejxRWXqG+ZETfxxlmlRns0QG
J4+BYL0CoanDSeA4fuyn4Bv7cy/03TDhfg/Uq0Aeg+hhcPE/vx3ebPsCgYEAy/Pv
eDLssOSdeyIxf0Brtocg6aPXIVaLdus+bXmLg77rJIFytAZmTTW8SkkSczWtucI3
FI1I6sei/8FdPzAl62/JDdlf7Wd9K7JIotY4TzT7Tm7QU7xpfLLYIP1bOFjN81rk
77oOD4LsXcosB/U6s1blPJMZ6AlO2EKs10UuR1cCgYBipzuJ2ADEaOz9RLWwi0AH
Pza2Sj+c2epQD9ZivD7Zo/Sid3ZwvGeGF13JyR7kLEdmAkgsHUdu1rI7mAolXMaB
1pdrsHureeLxGbRM6za3tzMXWv1Il7FQWoPC8ZwXvMOR1VQDv4nzq7vbbA8z8c+c
57+8tALQHOTDOgQIzwK61QKBgERGVc0EJy4Uag+VY8J4m1ZQKBluqo7TfP6DQ7O8
M5MX73maB/7yAX8pVO39RjrhJlYACRZNMbK+v/ckEQYdJSSKmGCVe0JrGYDuPtic
I9+IGfSorf7KHPoMmMN6bPYQ7Gjh7a++tgRFTMEc8956Hnt4xGahy9NcglNtBpVN
6G8jAoGBAMCh028pdzJa/xeBHLLaVB2sc0Fe7993WlsPmnVE779dAz7qMscOtXJK
fgtriltLSSD6rTA9hUAsL/X62rY0wdXuNdijjBb/qvrx7CAV6i37NK1CjABNjsfG
ZM372Ac6zc1EqSrid2IjET1YqyIW2KGLI1R2xbQc98UGlt48OdWu
-----END RSA PRIVATE KEY-----
`

// https://developer.github.com/v3/apps/#response-9
var githubConversionJSON = `{
	"id":      1,
	"node_id": "MDM6QXBwNTk=",
	"owner": {
		"login":               "octocat",
		"id":                  1,
		"node_id":             "MDQ6VXNlcjE=",
		"avatar_url":          "https://github.com/images/error/octocat_happy.gif",
		"gravatar_id":         "",
		"url":                 "https://api.github.com/users/octocat",
		"html_url":            "https://github.com/octocat",
		"followers_url":       "https://api.github.com/users/octocat/followers",
		"following_url":       "https://api.github.com/users/octocat/following{/other_user}",
		"gists_url":           "https://api.github.com/users/octocat/gists{/gist_id}",
		"starred_url":         "https://api.github.com/users/octocat/starred{/owner}{/repo}",
		"subscriptions_url":   "https://api.github.com/users/octocat/subscriptions",
		"organizations_url":   "https://api.github.com/users/octocat/orgs",
		"repos_url":           "https://api.github.com/users/octocat/repos",
		"events_url":          "https://api.github.com/users/octocat/events{/privacy}",
		"received_events_url": "https://api.github.com/users/octocat/received_events",
		"type":                "User",
		"site_admin":          false
	},
	"name":           "Atlantis",
	"description":    null,
	"external_url":   "https://atlantis.example.com",
	"html_url":       "https://github.com/apps/atlantis",
	"created_at":     "2018-09-13T12:28:37Z",
	"updated_at":     "2018-09-13T12:28:37Z",
	"client_id":      "Iv1.8a61f9b3a7aba766",
	"client_secret":  "1726be1638095a19edd134c77bde3aa2ece1e5d8",
	"webhook_secret": "e340154128314309424b7c8e90325147d99fdafa",
	"pem":            "%s"
}`

var githubAppInstallationJSON = `[
	{
		"id": 1,
		"account": {
			"login": "github",
			"id": 1,
			"node_id": "MDEyOk9yZ2FuaXphdGlvbjE=",
			"url": "https://api.github.com/orgs/github",
			"repos_url": "https://api.github.com/orgs/github/repos",
			"events_url": "https://api.github.com/orgs/github/events",
			"hooks_url": "https://api.github.com/orgs/github/hooks",
			"issues_url": "https://api.github.com/orgs/github/issues",
			"members_url": "https://api.github.com/orgs/github/members{/member}",
			"public_members_url": "https://api.github.com/orgs/github/public_members{/member}",
			"avatar_url": "https://github.com/images/error/octocat_happy.gif",
			"description": "A great organization"
		},
		"access_tokens_url": "https://api.github.com/installations/1/access_tokens",
		"repositories_url": "https://api.github.com/installation/repositories",
		"html_url": "https://github.com/organizations/github/settings/installations/1",
		"app_id": 1,
		"target_id": 1,
		"target_type": "Organization",
		"permissions": {
			"metadata": "read",
			"contents": "read",
			"issues": "write",
			"single_file": "write"
		},
		"events": [
			"push",
			"pull_request"
		],
		"single_file_name": "config.yml",
		"repository_selection": "selected"
	}
]`

var githubAppMultipleInstallationJSON = `[
	{
		"id": 1,
		"account": {
			"login": "github",
			"id": 1,
			"node_id": "MDEyOk9yZ2FuaXphdGlvbjE=",
			"url": "https://api.github.com/orgs/github",
			"repos_url": "https://api.github.com/orgs/github/repos",
			"events_url": "https://api.github.com/orgs/github/events",
			"hooks_url": "https://api.github.com/orgs/github/hooks",
			"issues_url": "https://api.github.com/orgs/github/issues",
			"members_url": "https://api.github.com/orgs/github/members{/member}",
			"public_members_url": "https://api.github.com/orgs/github/public_members{/member}",
			"avatar_url": "https://github.com/images/error/octocat_happy.gif",
			"description": "A great organization"
		},
		"access_tokens_url": "https://api.github.com/installations/1/access_tokens",
		"repositories_url": "https://api.github.com/installation/repositories",
		"html_url": "https://github.com/organizations/github/settings/installations/1",
		"app_id": 1,
		"target_id": 1,
		"target_type": "Organization",
		"permissions": {
			"metadata": "read",
			"contents": "read",
			"issues": "write",
			"single_file": "write"
		},
		"events": [
			"push",
			"pull_request"
		],
		"single_file_name": "config.yml",
		"repository_selection": "selected"
	},
	{
		"id": 2,
		"account": {
			"login": "github",
			"id": 1,
			"node_id": "MDEyOk9yZ2FuaXphdGlvbjE=",
			"url": "https://api.github.com/orgs/github",
			"repos_url": "https://api.github.com/orgs/github/repos",
			"events_url": "https://api.github.com/orgs/github/events",
			"hooks_url": "https://api.github.com/orgs/github/hooks",
			"issues_url": "https://api.github.com/orgs/github/issues",
			"members_url": "https://api.github.com/orgs/github/members{/member}",
			"public_members_url": "https://api.github.com/orgs/github/public_members{/member}",
			"avatar_url": "https://github.com/images/error/octocat_happy.gif",
			"description": "A great organization"
		},
		"access_tokens_url": "https://api.github.com/installations/1/access_tokens",
		"repositories_url": "https://api.github.com/installation/repositories",
		"html_url": "https://github.com/organizations/github/settings/installations/1",
		"app_id": 1,
		"target_id": 1,
		"target_type": "Organization",
		"permissions": {
			"metadata": "read",
			"contents": "read",
			"issues": "write",
			"single_file": "write"
		},
		"events": [
			"push",
			"pull_request"
		],
		"single_file_name": "config.yml",
		"repository_selection": "selected"
	}
]`

// nolint: gosec
var githubAppTokenJSON = `{
	"token":      "some-token",
	"expires_at": "2050-01-01T00:00:00Z",
	"permissions": {
		"issues":   "write",
		"contents": "read"
	},
	"repositories": [
		{
			"id":        1296269,
			"node_id":   "MDEwOlJlcG9zaXRvcnkxMjk2MjY5",
			"name":      "Hello-World",
			"full_name": "octocat/Hello-World",
			"owner": {
				"login":               "octocat",
				"id":                  1,
				"node_id":             "MDQ6VXNlcjE=",
				"avatar_url":          "https://github.com/images/error/octocat_happy.gif",
				"gravatar_id":         "",
				"url":                 "https://api.github.com/users/octocat",
				"html_url":            "https://github.com/octocat",
				"followers_url":       "https://api.github.com/users/octocat/followers",
				"following_url":       "https://api.github.com/users/octocat/following{/other_user}",
				"gists_url":           "https://api.github.com/users/octocat/gists{/gist_id}",
				"starred_url":         "https://api.github.com/users/octocat/starred{/owner}{/repo}",
				"subscriptions_url":   "https://api.github.com/users/octocat/subscriptions",
				"organizations_url":   "https://api.github.com/users/octocat/orgs",
				"repos_url":           "https://api.github.com/users/octocat/repos",
				"events_url":          "https://api.github.com/users/octocat/events{/privacy}",
				"received_events_url": "https://api.github.com/users/octocat/received_events",
				"type":                "User",
				"site_admin":          false
			},
			"private":           false,
			"html_url":          "https://github.com/octocat/Hello-World",
			"description":       "This your first repo!",
			"fork":              false,
			"url":               "https://api.github.com/repos/octocat/Hello-World",
			"archive_url":       "http://api.github.com/repos/octocat/Hello-World/{archive_format}{/ref}",
			"assignees_url":     "http://api.github.com/repos/octocat/Hello-World/assignees{/user}",
			"blobs_url":         "http://api.github.com/repos/octocat/Hello-World/git/blobs{/sha}",
			"branches_url":      "http://api.github.com/repos/octocat/Hello-World/branches{/branch}",
			"collaborators_url": "http://api.github.com/repos/octocat/Hello-World/collaborators{/collaborator}",
			"comments_url":      "http://api.github.com/repos/octocat/Hello-World/comments{/number}",
			"commits_url":       "http://api.github.com/repos/octocat/Hello-World/commits{/sha}",
			"compare_url":       "http://api.github.com/repos/octocat/Hello-World/compare/{base}...{head}",
			"contents_url":      "http://api.github.com/repos/octocat/Hello-World/contents/{+path}",
			"contributors_url":  "http://api.github.com/repos/octocat/Hello-World/contributors",
			"deployments_url":   "http://api.github.com/repos/octocat/Hello-World/deployments",
			"downloads_url":     "http://api.github.com/repos/octocat/Hello-World/downloads",
			"events_url":        "http://api.github.com/repos/octocat/Hello-World/events",
			"forks_url":         "http://api.github.com/repos/octocat/Hello-World/forks",
			"git_commits_url":   "http://api.github.com/repos/octocat/Hello-World/git/commits{/sha}",
			"git_refs_url":      "http://api.github.com/repos/octocat/Hello-World/git/refs{/sha}",
			"git_tags_url":      "http://api.github.com/repos/octocat/Hello-World/git/tags{/sha}",
			"git_url":           "git:github.com/octocat/Hello-World.git",
			"issue_comment_url": "http://api.github.com/repos/octocat/Hello-World/issues/comments{/number}",
			"issue_events_url":  "http://api.github.com/repos/octocat/Hello-World/issues/events{/number}",
			"issues_url":        "http://api.github.com/repos/octocat/Hello-World/issues{/number}",
			"keys_url":          "http://api.github.com/repos/octocat/Hello-World/keys{/key_id}",
			"labels_url":        "http://api.github.com/repos/octocat/Hello-World/labels{/name}",
			"languages_url":     "http://api.github.com/repos/octocat/Hello-World/languages",
			"merges_url":        "http://api.github.com/repos/octocat/Hello-World/merges",
			"milestones_url":    "http://api.github.com/repos/octocat/Hello-World/milestones{/number}",
			"notifications_url": "http://api.github.com/repos/octocat/Hello-World/notifications{?since,all,participating}",
			"pulls_url":         "http://api.github.com/repos/octocat/Hello-World/pulls{/number}",
			"releases_url":      "http://api.github.com/repos/octocat/Hello-World/releases{/id}",
			"ssh_url":           "git@github.com:octocat/Hello-World.git",
			"stargazers_url":    "http://api.github.com/repos/octocat/Hello-World/stargazers",
			"statuses_url":      "http://api.github.com/repos/octocat/Hello-World/statuses/{sha}",
			"subscribers_url":   "http://api.github.com/repos/octocat/Hello-World/subscribers",
			"subscription_url":  "http://api.github.com/repos/octocat/Hello-World/subscription",
			"tags_url":          "http://api.github.com/repos/octocat/Hello-World/tags",
			"teams_url":         "http://api.github.com/repos/octocat/Hello-World/teams",
			"trees_url":         "http://api.github.com/repos/octocat/Hello-World/git/trees{/sha}",
			"clone_url":         "https://github.com/octocat/Hello-World.git",
			"mirror_url":        "git:git.example.com/octocat/Hello-World",
			"hooks_url":         "http://api.github.com/repos/octocat/Hello-World/hooks",
			"svn_url":           "https://svn.github.com/octocat/Hello-World",
			"homepage":          "https://github.com",
			"language":          null,
			"forks_count":       9,
			"stargazers_count":  80,
			"watchers_count":    80,
			"size":              108,
			"default_branch":    "main",
			"open_issues_count": 0,
			"is_template":       true,
			"topics": [
				"octocat",
				"atom",
				"electron",
				"api"
			],
			"has_issues":    true,
			"has_projects":  true,
			"has_wiki":      true,
			"has_pages":     false,
			"has_downloads": true,
			"archived":      false,
			"disabled":      false,
			"visibility":    "public",
			"pushed_at":     "2011-01-26T19:06:43Z",
			"created_at":    "2011-01-26T19:01:12Z",
			"updated_at":    "2011-01-26T19:14:43Z",
			"permissions": {
				"admin": false,
				"push":  false,
				"pull":  true
			},
			"allow_rebase_merge":  true,
			"template_repository": null,
			"temp_clone_token":    "ABTLWHOULUVAXGTRYU7OC2876QJ2O",
			"allow_squash_merge":  true,
			"allow_merge_commit":  true,
			"subscribers_count":   42,
			"network_count":       0
		}
	]
}`

var githubAppJSON = `{
	"id": 1,
	"slug": "octoapp",
	"node_id": "MDExOkludGVncmF0aW9uMQ==",
	"owner": {
	  "login": "github",
	  "id": 1,
	  "node_id": "MDEyOk9yZ2FuaXphdGlvbjE=",
	  "url": "https://api.github.com/orgs/github",
	  "repos_url": "https://api.github.com/orgs/github/repos",
	  "events_url": "https://api.github.com/orgs/github/events",
	  "avatar_url": "https://github.com/images/error/octocat_happy.gif",
	  "gravatar_id": "",
	  "html_url": "https://github.com/octocat",
	  "followers_url": "https://api.github.com/users/octocat/followers",
	  "following_url": "https://api.github.com/users/octocat/following{/other_user}",
	  "gists_url": "https://api.github.com/users/octocat/gists{/gist_id}",
	  "starred_url": "https://api.github.com/users/octocat/starred{/owner}{/repo}",
	  "subscriptions_url": "https://api.github.com/users/octocat/subscriptions",
	  "organizations_url": "https://api.github.com/users/octocat/orgs",
	  "received_events_url": "https://api.github.com/users/octocat/received_events",
	  "type": "User",
	  "site_admin": true
	},
	"name": "Octocat App",
	"description": "",
	"external_url": "https://example.com",
	"html_url": "https://github.com/apps/octoapp",
	"created_at": "2017-07-08T16:18:44-04:00",
	"updated_at": "2017-07-08T16:18:44-04:00",
	"permissions": {
	  "metadata": "read",
	  "contents": "read",
	  "issues": "write",
	  "single_file": "write"
	},
	"events": [
	  "push",
	  "pull_request"
	]
  }`

func validateGithubToken(tokenString string) error {
	key, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(GithubPrivateKey))
	if err != nil {
		return fmt.Errorf("could not parse private key: %s", err)
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			err := fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])

			return nil, err
		}

		return key.Public(), nil
	})

	if err != nil {
		return err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); !ok || !token.Valid || claims["iss"] != "1" {
		return fmt.Errorf("Invalid token")
	}
	return nil
}

func GithubAppTestServer(t *testing.T) (string, error) {
	counter := 0
	testServer := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.RequestURI {
			case "/api/v3/app-manifests/good-code/conversions":
				encodedKey := strings.Join(strings.Split(GithubPrivateKey, "\n"), "\\n")
				appInfo := fmt.Sprintf(githubConversionJSON, encodedKey)
				w.Write([]byte(appInfo)) // nolint: errcheck
			// https://developer.github.com/v3/apps/#list-installations
			case "/api/v3/app/installations":
				token := strings.Replace(r.Header.Get("Authorization"), "Bearer ", "", 1)
				if err := validateGithubToken(token); err != nil {
					w.WriteHeader(403)
					w.Write([]byte("Invalid token")) // nolint: errcheck
					return
				}

				w.Write([]byte(githubAppInstallationJSON)) // nolint: errcheck
				return
			case "/api/v3/apps/some-app":
				token := strings.Replace(r.Header.Get("Authorization"), "token ", "", 1)

				// token is taken from githubAppTokenJSON
				if token != "some-token" {
					w.WriteHeader(403)
					w.Write([]byte("Invalid installation token")) // nolint: errcheck
					return
				}
				w.Write([]byte(githubAppJSON)) // nolint: errcheck
				return
			case "/api/v3/app/installations/1/access_tokens":
				token := strings.Replace(r.Header.Get("Authorization"), "Bearer ", "", 1)
				if err := validateGithubToken(token); err != nil {
					w.WriteHeader(403)
					w.Write([]byte("Invalid token")) // nolint: errcheck
					return
				}

				appToken := fmt.Sprintf(githubAppTokenJSON, counter)
				counter++
				w.Write([]byte(appToken)) // nolint: errcheck
				return
			default:
				t.Errorf("got unexpected request at %q", r.RequestURI)
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
		}))

	testServerURL, err := url.Parse(testServer.URL)

	return testServerURL.Host, err
}

func GithubMultipleAppTestServer(t *testing.T) (string, error) {
	counter := 0
	testServer := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.RequestURI {
			case "/api/v3/app-manifests/good-code/conversions":
				encodedKey := strings.Join(strings.Split(GithubPrivateKey, "\n"), "\\n")
				appInfo := fmt.Sprintf(githubConversionJSON, encodedKey)
				w.Write([]byte(appInfo)) // nolint: errcheck
			// https://developer.github.com/v3/apps/#list-installations
			case "/api/v3/app/installations":
				token := strings.Replace(r.Header.Get("Authorization"), "Bearer ", "", 1)
				if err := validateGithubToken(token); err != nil {
					w.WriteHeader(403)
					w.Write([]byte("Invalid token")) // nolint: errcheck
					return
				}

				w.Write([]byte(githubAppMultipleInstallationJSON)) // nolint: errcheck
				return
			case "/api/v3/apps/some-app":
				token := strings.Replace(r.Header.Get("Authorization"), "token ", "", 1)

				// token is taken from githubAppTokenJSON
				if token != "some-token" {
					w.WriteHeader(403)
					w.Write([]byte("Invalid installation token")) // nolint: errcheck
					return
				}
				w.Write([]byte(githubAppJSON)) // nolint: errcheck
				return
			case "/api/v3/app/installations/1/access_tokens":
				token := strings.Replace(r.Header.Get("Authorization"), "Bearer ", "", 1)
				if err := validateGithubToken(token); err != nil {
					w.WriteHeader(403)
					w.Write([]byte("Invalid token")) // nolint: errcheck
					return
				}

				appToken := fmt.Sprintf(githubAppTokenJSON, counter)
				counter++
				w.Write([]byte(appToken)) // nolint: errcheck
				return
			default:
				t.Errorf("got unexpected request at %q", r.RequestURI)
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
		}))

	testServerURL, err := url.Parse(testServer.URL)

	return testServerURL.Host, err
}
