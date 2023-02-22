package utils_test

import (
	"fmt"
	"testing"

	"github.com/runatlantis/atlantis/server/utils"
	. "github.com/runatlantis/atlantis/testing"
)

func Test_IsSimilarWord(t *testing.T) {
	t.Log("check if given executable name is misspelled or just an unrelated word")

	spellings := []struct {
		Misspelled bool
		Given      string
		Want       string
	}{
		{
			false,
			"atlantis",
			"atlantis",
		},
		{
			false,
			"maybe",
			"atlantis",
		},
		{
			false,
			"atlantis-qa",
			"atlantis-prod",
		},
		{
			true,
			"altantis",
			"atlantis",
		},
		{
			true,
			"atlants",
			"atlantis",
		},
		{
			true,
			"teraform",
			"terraform",
		},
	}

	for _, s := range spellings {
		t.Run(fmt.Sprintf("given %s want %s", s.Given, s.Want), func(t *testing.T) {
			isMisspelled := utils.IsSimilarWord(s.Given, s.Want)

			if s.Misspelled {
				Equals(t, isMisspelled, true)
			}

			if !s.Misspelled {
				Equals(t, isMisspelled, false)
			}
		})
	}

}
