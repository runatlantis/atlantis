package fetchers

import (
	"errors"
	"testing"

	. "github.com/runatlantis/atlantis/testing"
)

func TestGetType(t *testing.T) {
	testCases := []struct {
		Name       string
		Reference  string
		SourceType ConfigSourceType
		Error      error
	}{
		{
			Name:       "Github HTTPS",
			Reference:  "https://github.com/runatlantis/atlantis",
			SourceType: Github,
			Error:      nil,
		},
		{
			Name:       "Github HTTP",
			Reference:  "http://github.com/runatlantis/atlantis",
			SourceType: Github,
			Error:      nil,
		},
		{
			Name:       "Unknown Hostname",
			Reference:  "https://example.com/runatlantis/atlantis",
			SourceType: 0,
			Error:      errors.New("unknown source in remote reference: https://example.com/runatlantis/atlantis"),
		},
		{
			Name:       "Random String",
			Reference:  "some-random-string",
			SourceType: 0,
			Error:      errors.New("unknown source in remote reference: some-random-string"),
		},
		{
			Name:       "Empty String",
			Reference:  "",
			SourceType: 0,
			Error:      errors.New("unknown source in remote reference: "),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Reference, func(t *testing.T) {
			sourceType, err := GetType(tc.Reference)

			if tc.Error == nil {
				Equals(t, nil, err)
			} else {
				ErrContains(t, tc.Error.Error(), err)
			}
			Equals(t, tc.SourceType, sourceType)
		})
	}
}
