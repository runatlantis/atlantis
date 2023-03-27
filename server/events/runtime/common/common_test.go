package common

import (
	"reflect"
	"testing"
)

func Test_DeDuplicateExtraArgs(t *testing.T) {
	cases := []struct {
		description  string
		inputArgs    []string
		extraArgs    []string
		expectedArgs []string
	}{
		{
			"No extra args",
			[]string{"init", "-input=false", "-no-color", "-upgrade"},
			[]string{},
			[]string{"init", "-input=false", "-no-color", "-upgrade"},
		},
		{
			"Override -upgrade",
			[]string{"init", "-input=false", "-no-color", "-upgrade"},
			[]string{"-upgrade=false"},
			[]string{"init", "-input=false", "-no-color", "-upgrade=false"},
		},
		{
			"Override -input",
			[]string{"init", "-input=false", "-no-color", "-upgrade"},
			[]string{"-input=true"},
			[]string{"init", "-input=true", "-no-color", "-upgrade"},
		},
		{
			"Override -input and -upgrade",
			[]string{"init", "-input=false", "-no-color", "-upgrade"},
			[]string{"-input=true", "-upgrade=false"},
			[]string{"init", "-input=true", "-no-color", "-upgrade=false"},
		},
		{
			"Non duplicate extra args",
			[]string{"init", "-input=false", "-no-color", "-upgrade"},
			[]string{"extra", "args"},
			[]string{"init", "-input=false", "-no-color", "-upgrade", "extra", "args"},
		},
		{
			"Override upgrade with extra args",
			[]string{"init", "-input=false", "-no-color", "-upgrade"},
			[]string{"extra", "args", "-upgrade=false"},
			[]string{"init", "-input=false", "-no-color", "-upgrade=false", "extra", "args"},
		},
		{
			"Override -input (using --input)",
			[]string{"init", "-input=false", "-no-color", "-upgrade"},
			[]string{"--input=true"},
			[]string{"init", "--input=true", "-no-color", "-upgrade"},
		},
		{
			"Override -input (using --input) and -upgrade (using --upgrade)",
			[]string{"init", "-input=false", "-no-color", "-upgrade"},
			[]string{"--input=true", "--upgrade=false"},
			[]string{"init", "--input=true", "-no-color", "--upgrade=false"},
		},
		{
			"Override long form flag ",
			[]string{"init", "--input=false", "-no-color", "-upgrade"},
			[]string{"--input=true"},
			[]string{"init", "--input=true", "-no-color", "-upgrade"},
		},
		{
			"Override --input using (-input) ",
			[]string{"init", "--input=false", "-no-color", "-upgrade"},
			[]string{"-input=true"},
			[]string{"init", "-input=true", "-no-color", "-upgrade"},
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			finalArgs := DeDuplicateExtraArgs(c.inputArgs, c.extraArgs)

			if !reflect.DeepEqual(c.expectedArgs, finalArgs) {
				t.Fatalf("finalArgs (%v) does not match expectedArgs (%v)", finalArgs, c.expectedArgs)
			}
		})
	}
}
