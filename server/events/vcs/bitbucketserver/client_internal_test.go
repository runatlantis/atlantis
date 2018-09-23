package bitbucketserver

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
		{
			"12345678901" +
			"```diff\n" +
			"12345678901234567890" +
			"12345678901234567890" +
			"12345678901234567890" +
			"```\n\n" +
			"1234567890",
			25,
			[]string{"12345678901```diff\n12\n```\n\njoin",
					"```diff\n345678901234567890123\n```\n\njoin",
					"```diff\n456789012345678901234\n```\n\njoin",
					"```diff\n5678901234567890```\n\njoin",
					"1234567890"},
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
		client := Client{}
		t.Run(c.comment, func(t *testing.T) {
			Equals(t, c.exp, client.splitAtMaxChars(c.comment, c.max, "join"))
		})
	}
}
