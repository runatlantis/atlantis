package runtime

import (
	. "github.com/runatlantis/atlantis/testing"
	"testing"
)

func TestCleanRemoteOpOutput(t *testing.T) {
	cases := []struct {
		out string
		exp string
	}{
		{
			`
above
------------------------------------------------------------------------
below`,
			"below",
		},
		{
			"nodelim",
			"nodelim",
		},
	}

	for _, c := range cases {
		t.Run(c.exp, func(t *testing.T) {
			a := ApplyStepRunner{}
			Equals(t, c.exp, a.cleanRemoteOpOutput(c.out))
		})
	}
}
