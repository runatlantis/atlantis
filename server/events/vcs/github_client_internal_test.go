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

package vcs

import (
	"testing"

	. "github.com/runatlantis/atlantis/testing"
)

// If under the maximum number of chars, we shouldn't split the comments.
func TestSplitAtMaxChars_UnderMax(t *testing.T) {
	client := &GithubClient{}
	comment := "comment under max size"
	split := client.splitAtMaxChars(comment, len(comment)+1)
	Equals(t, []string{comment}, split)
}

// If the comment is over the max number of chars, we should split it into
// multiple comments.
func TestSplitAtMaxChars_OverMaxOnce(t *testing.T) {
	client := &GithubClient{}
	comment := "comment over max size"
	split := client.splitAtMaxChars(comment, len(comment)-1)
	Equals(t, []string{"comment over max siz" + detailsClose, detailsOpen + "e"}, split)
}

// Test that it works for multiple comments.
func TestSplitAtMaxChars_OverMaxMultiple(t *testing.T) {
	client := &GithubClient{}
	comment := "comment over max size"
	third := len(comment) / 3
	split := client.splitAtMaxChars(comment, third)
	Equals(t, []string{
		comment[:third] + detailsClose,
		detailsOpen + comment[third:third*2] + detailsClose,
		detailsOpen + comment[third*2:]}, split)
}

// If the hostname is github.com, should use normal BaseURL.
func TestNewGithubClient_GithubCom(t *testing.T) {
	client, err := NewGithubClient("github.com", "user", "pass")
	Ok(t, err)
	Equals(t, "https://api.github.com/", client.client.BaseURL.String())
}

// If the hostname is a non-github hostname should use the right BaseURL.
func TestNewGithubClient_NonGithub(t *testing.T) {
	client, err := NewGithubClient("example.com", "user", "pass")
	Ok(t, err)
	Equals(t, "https://example.com/api/v3/", client.client.BaseURL.String())
}
