// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package azuredevops

import (
	"testing"

	. "github.com/runatlantis/atlantis/testing"
)

func TestGitStatusContextFromSrc(t *testing.T) {
	cases := []struct {
		src      string
		expGenre string
		expName  string
	}{
		{
			"atlantis/plan",
			"Atlantis Bot/atlantis",
			"plan",
		},
		{
			"atlantis/foo/bar/biz/baz",
			"Atlantis Bot/atlantis/foo/bar/biz",
			"baz",
		},
		{
			"foo",
			"Atlantis Bot",
			"foo",
		},
		{
			"",
			"Atlantis Bot",
			"",
		},
	}

	for _, c := range cases {
		result := gitStatusContextFromSrc(c.src)
		expName := c.expName
		expGenre := c.expGenre
		Equals(t, &expName, result.Name)
		Equals(t, &expGenre, result.Genre)
	}
}
