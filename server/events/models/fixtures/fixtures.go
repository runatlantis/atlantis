package fixtures

import "github.com/hootsuite/atlantis/server/events/models"

var Pull = models.PullRequest{
	Num:        1,
	HeadCommit: "16ca62f65c18ff456c6ef4cacc8d4826e264bb17",
	Branch:     "branch",
	Author:     "lkysow",
	URL:        "url",
	BaseCommit: "8ed0280678d49d42cd286610aabcfceb5bb673c6",
}

var Repo = models.Repo{
	CloneURL:          "https://user:password@github.com/hootsuite/atlantis.git",
	FullName:          "hootsuite/atlantis",
	Owner:             "hootsuite",
	SanitizedCloneURL: "https://github.com/hootsuite/atlantis.git",
	Name:              "atlantis",
}

var User = models.User{
	Username: "lkysow",
}
