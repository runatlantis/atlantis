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
//
package models_test

import (
	"fmt"
	"testing"

	"github.com/runatlantis/atlantis/server/events/models"
	. "github.com/runatlantis/atlantis/testing"
)

func TestNewRepo_EmptyRepoFullName(t *testing.T) {
	_, err := models.NewRepo(models.Github, "", "https://github.com/notowner/repo.git", "u", "p")
	ErrEquals(t, "repoFullName can't be empty", err)
}

func TestNewRepo_EmptyCloneURL(t *testing.T) {
	_, err := models.NewRepo(models.Github, "owner/repo", "", "u", "p")
	ErrEquals(t, "cloneURL can't be empty", err)
}

func TestNewRepo_InvalidCloneURL(t *testing.T) {
	_, err := models.NewRepo(models.Github, "owner/repo", ":", "u", "p")
	ErrEquals(t, "invalid clone url: parse :.git: missing protocol scheme", err)
}

func TestNewRepo_CloneURLWrongRepo(t *testing.T) {
	_, err := models.NewRepo(models.Github, "owner/repo", "https://github.com/notowner/repo.git", "u", "p")
	ErrEquals(t, `expected clone url to have path "/owner/repo.git" but had "/notowner/repo.git"`, err)
}

// For bitbucket server we don't validate the clone URL because the callers
// are actually constructing it.
func TestNewRepo_CloneURLBitbucketServer(t *testing.T) {
	repo, err := models.NewRepo(models.BitbucketServer, "owner/repo", "http://mycorp.com:7990/scm/at/atlantis-example.git", "u", "p")
	Ok(t, err)
	Equals(t, models.Repo{
		FullName:          "owner/repo",
		Owner:             "owner",
		Name:              "repo",
		CloneURL:          "http://u:p@mycorp.com:7990/scm/at/atlantis-example.git",
		SanitizedCloneURL: "http://mycorp.com:7990/scm/at/atlantis-example.git",
		VCSHost: models.VCSHost{
			Hostname: "mycorp.com",
			Type:     models.BitbucketServer,
		},
	}, repo)
}

func TestNewRepo_GithubFullNameWrongFormat(t *testing.T) {
	cases := []string{
		"owner/repo/extra",
		"/",
		"//",
		"///",
		"a/",
		"/b",
	}

	for _, c := range cases {
		t.Run(c, func(t *testing.T) {
			cloneURL := fmt.Sprintf("https://github.com/%s.git", c)
			_, err := models.NewRepo(models.Github, c, cloneURL, "u", "p")
			ErrEquals(t, fmt.Sprintf(`invalid repo format "%s"`, c), err)
		})
	}
}

func TestNewRepo_BitbucketFullNameFormatCases(t *testing.T) {

	testCases := []struct {
		Case            string
		Vcs             models.VCSHostType
		RepoName        string
		ExpectedMessage string
	}{
		{"Valid Bitbucket cloud repository testing %s case", models.BitbucketCloud, "owner/repo", ""},
		{"Invalid Bitbucket cloud repository testing %s case", models.BitbucketCloud, "owner/subgroup/repo", `invalid repo format "%s"`},
		{"Invalid Bitbucket cloud repository testing `%s` case", models.BitbucketCloud, "/", `invalid repo format "%s"`},
		{"Invalid Bitbucket cloud repository testing `%s` case", models.BitbucketCloud, "//", `invalid repo format "%s"`},
		{"Invalid Bitbucket cloud repository testing `%s` case", models.BitbucketCloud, "///", `invalid repo format "%s"`},
		{"Invalid Bitbucket cloud repository testing `%s` case", models.BitbucketCloud, "a/", `invalid repo format "%s"`},
		{"Invalid Bitbucket cloud repository testing `%s` case", models.BitbucketCloud, "/b", `invalid repo format "%s"`},

		{"Valid Bitbucket server repository testing %s case", models.BitbucketCloud, "owner/repo", ""},
		{"Invalid Bitbucket server repository testing %s case", models.BitbucketServer, "owner/subgroup/repo", `invalid repo format "%s"`},
		{"Invalid Bitbucket server repository testing `%s` case", models.BitbucketServer, "/", `invalid repo format "%s"`},
		{"Invalid Bitbucket server repository testing `%s` case", models.BitbucketServer, "//", `invalid repo format "%s"`},
		{"Invalid Bitbucket server repository testing `%s` case", models.BitbucketServer, "///", `invalid repo format "%s"`},
		{"Invalid Bitbucket server repository testing `%s` case", models.BitbucketServer, "a/", `invalid repo format "%s"`},
		{"Invalid Bitbucket server repository testing `%s` case", models.BitbucketServer, "/b", `invalid repo format "%s"`},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf(testCase.Case, testCase.RepoName), func(t *testing.T) {
			cloneURL := fmt.Sprintf("https://bitbucket.org/%s.git", testCase.RepoName)
			_, err := models.NewRepo(testCase.Vcs, testCase.RepoName, cloneURL, "u", "p")

			if err != nil {
				ErrEquals(t, fmt.Sprintf(testCase.ExpectedMessage, testCase.RepoName), err)
			}
		})
	}
}

func TestNewRepo_GitlabFullNameFormatCases(t *testing.T) {

	testCases := []struct {
		Case            string
		Vcs             models.VCSHostType
		RepoName        string
		ExpectedMessage string
	}{
		{"Valid gitlab repository testing %s case", models.Gitlab, "owner/repo", ""},
		{"Valid gitlab repository testing %s case", models.Gitlab, "owner/repo/extra", ""},
		{"Invalid gitlab repository testing `%s` case", models.Gitlab, "/", `invalid repo format "%s"`},
		{"Invalid gitlab repository testing `%s` case", models.Gitlab, "//", `invalid repo format "%s"`},
		{"Invalid gitlab repository testing `%s` case", models.Gitlab, "///", `invalid repo format "%s"`},
		{"Invalid gitlab repository testing `%s` case", models.Gitlab, "a/", `invalid repo format "%s"`},
		{"Invalid gitlab repository testing `%s` case", models.Gitlab, "/b", `invalid repo format "%s"`},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf(testCase.Case, testCase.RepoName), func(t *testing.T) {
			cloneURL := fmt.Sprintf("https://gitlab.com/%s.git", testCase.RepoName)
			_, err := models.NewRepo(testCase.Vcs, testCase.RepoName, cloneURL, "u", "p")

			if err != nil {
				ErrEquals(t, fmt.Sprintf(testCase.ExpectedMessage, testCase.RepoName), err)
			}
		})
	}
}

// If the clone url doesn't end with .git it is appended
func TestNewRepo_MissingDotGit(t *testing.T) {
	repo, err := models.NewRepo(models.BitbucketCloud, "owner/repo", "https://bitbucket.org/owner/repo", "u", "p")
	Ok(t, err)
	Equals(t, repo.CloneURL, "https://u:p@bitbucket.org/owner/repo.git")
	Equals(t, repo.SanitizedCloneURL, "https://bitbucket.org/owner/repo.git")
}

func TestNewRepo_HTTPAuth(t *testing.T) {
	// When the url has http the auth should be added.
	repo, err := models.NewRepo(models.Github, "owner/repo", "http://github.com/owner/repo.git", "u", "p")
	Ok(t, err)
	Equals(t, models.Repo{
		VCSHost: models.VCSHost{
			Hostname: "github.com",
			Type:     models.Github,
		},
		SanitizedCloneURL: "http://github.com/owner/repo.git",
		CloneURL:          "http://u:p@github.com/owner/repo.git",
		FullName:          "owner/repo",
		Owner:             "owner",
		Name:              "repo",
	}, repo)
}

func TestNewRepo_HTTPSAuth(t *testing.T) {
	// When the url has https the auth should be added.
	repo, err := models.NewRepo(models.Github, "owner/repo", "https://github.com/owner/repo.git", "u", "p")
	Ok(t, err)
	Equals(t, models.Repo{
		VCSHost: models.VCSHost{
			Hostname: "github.com",
			Type:     models.Github,
		},
		SanitizedCloneURL: "https://github.com/owner/repo.git",
		CloneURL:          "https://u:p@github.com/owner/repo.git",
		FullName:          "owner/repo",
		Owner:             "owner",
		Name:              "repo",
	}, repo)
}

func TestProject_String(t *testing.T) {
	Equals(t, "repofullname=owner/repo path=my/path", (models.Project{
		RepoFullName: "owner/repo",
		Path:         "my/path",
	}).String())
}

func TestNewProject(t *testing.T) {
	cases := []struct {
		path    string
		expPath string
	}{
		{
			"/",
			".",
		},
		{
			"./another/path",
			"another/path",
		},
		{
			".",
			".",
		},
	}

	for _, c := range cases {
		t.Run(c.path, func(t *testing.T) {
			p := models.NewProject("repo/owner", c.path)
			Equals(t, c.expPath, p.Path)
		})
	}
}

func TestVCSHostType_ToString(t *testing.T) {
	cases := []struct {
		vcsType models.VCSHostType
		exp     string
	}{
		{
			models.Github,
			"Github",
		},
		{
			models.Gitlab,
			"Gitlab",
		},
		{
			models.BitbucketCloud,
			"BitbucketCloud",
		},
		{
			models.BitbucketServer,
			"BitbucketServer",
		},
	}

	for _, c := range cases {
		t.Run(c.exp, func(t *testing.T) {
			Equals(t, c.exp, c.vcsType.String())
		})
	}
}
