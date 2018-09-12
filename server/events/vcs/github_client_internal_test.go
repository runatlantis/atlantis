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

func TestSplitAtMaxChars(t *testing.T) {
	cases := []struct {
		comment string
		max     int
		exp     []string
	}{
		// Test when comment is <= max length.
		{
			"",
			5,
			[]string{""},
		},
		{
			"1",
			5,
			[]string{"1"},
		},
		{
			"12345",
			5,
			[]string{"12345"},
		},
		// Now test when we need to join.
		{
			"123456",
			5,
			[]string{"1join", "2join", "3join", "4join", "5join", "6"},
		},
		{
			"123456",
			10,
			[]string{"123456"},
		},
		{
			"12345678901",
			10,
			[]string{"123456join", "78901"},
		},
		// Test the edge case of max < len("join")
		{
			"abc",
			2,
			nil,
		},
		{
			"abcde",
			4,
			nil,
		},
	}
	for _, c := range cases {
		client := GithubClient{}
		t.Run(c.comment, func(t *testing.T) {
			Equals(t, c.exp, client.splitAtMaxChars(c.comment, c.max, "join"))
		})
	}
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
