package valid_test

import (
	"testing"

	"github.com/runatlantis/atlantis/server/core/config/valid"
	. "github.com/runatlantis/atlantis/testing"
)

func TestConfig_IsPathIgnoredForAutoDiscover(t *testing.T) {
	cases := []struct {
		description  string
		autoDiscover valid.AutoDiscover
		path         string
		expIgnored   bool
	}{
		{
			description:  "auto discover configured, but not path",
			autoDiscover: valid.AutoDiscover{},
			path:         "foo",
			expIgnored:   false,
		},
		{
			description: "paths do not match pattern",
			autoDiscover: valid.AutoDiscover{
				IgnorePaths: []string{
					"bar",
				},
			},
			path:       "foo",
			expIgnored: false,
		},
		{
			description: "path does match pattern",
			autoDiscover: valid.AutoDiscover{
				IgnorePaths: []string{
					"fo?",
				},
			},
			path:       "foo",
			expIgnored: true,
		},
		{
			description: "one path matches pattern, another doesn't",
			autoDiscover: valid.AutoDiscover{
				IgnorePaths: []string{
					"fo*",
					"ba*",
				},
			},
			path:       "foo",
			expIgnored: true,
		},
		{
			description: "long path does match pattern",
			autoDiscover: valid.AutoDiscover{
				IgnorePaths: []string{
					"foo/*/baz",
				},
			},
			path:       "foo/bar/baz",
			expIgnored: true,
		},
		{
			description: "long path does not match pattern",
			autoDiscover: valid.AutoDiscover{
				IgnorePaths: []string{
					"foo/*/baz",
				},
			},
			path:       "foo/bar/boo",
			expIgnored: false,
		},
	}
	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {

			ignored := c.autoDiscover.IsPathIgnored(c.path)
			Equals(t, c.expIgnored, ignored)
		})
	}
}
