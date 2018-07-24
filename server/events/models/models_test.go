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

func TestNewRepo_FullNameWrongFormat(t *testing.T) {
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
