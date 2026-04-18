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

package models_test

import (
	"testing"

	"github.com/runatlantis/atlantis/server/events/models"
	. "github.com/runatlantis/atlantis/testing"
)

func TestParseGitlabHostname(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		host     string
		basePath string
	}{
		{"saas no scheme", "gitlab.com", "gitlab.com", ""},
		{"self-hosted no subpath", "acme.com", "acme.com", ""},
		{"subpath no scheme", "acme.com/gitlab", "acme.com", "/gitlab"},
		{"subpath trailing slash", "acme.com/gitlab/", "acme.com", "/gitlab"},
		{"https scheme with subpath", "https://acme.com/gitlab", "acme.com", "/gitlab"},
		{"http with port and subpath", "http://acme.com:8080/gitlab", "acme.com:8080", "/gitlab"},
		{"multi-level subpath trailing slash", "acme.com/group/sub/", "acme.com", "/group/sub"},
		{"https trailing slash", "https://acme.com/", "acme.com", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			host, basePath, err := models.ParseGitlabHostname(c.input)
			Ok(t, err)
			Equals(t, c.host, host)
			Equals(t, c.basePath, basePath)
		})
	}
}

func TestParseGitlabHostname_Errors(t *testing.T) {
	t.Run("empty input", func(t *testing.T) {
		_, _, err := models.ParseGitlabHostname("")
		ErrEquals(t, "gitlab hostname is empty", err)
	})

	t.Run("no host component", func(t *testing.T) {
		_, _, err := models.ParseGitlabHostname("https:///only/path")
		ErrContains(t, "has no host component", err)
	})

	t.Run("unparseable", func(t *testing.T) {
		_, _, err := models.ParseGitlabHostname("http://[::1")
		ErrContains(t, "parsing gitlab hostname", err)
	})
}
