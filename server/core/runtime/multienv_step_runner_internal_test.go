package runtime

import (
	"errors"
	"testing"
)

func TestMultiEnvStepRunner_Run_parser(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := map[string][]string{
			"":                nil,
			"KEY=value":       {"KEY", "value"},
			`KEY="value"`:     {"KEY", "value"},
			"KEY==":           {"KEY", "="},
			`KEY="'"`:         {"KEY", "'"},
			`KEY=""`:          {"KEY", ""},
			`KEY=a\"b`:        {"KEY", `a"b`},
			`KEY="va\"l\"ue"`: {"KEY", `va"l"ue`},

			"KEY='value'":   {"KEY", "value"},
			`KEY='va"l"ue'`: {"KEY", `va"l"ue`},
			`KEY='"'`:       {"KEY", `"`},
			"KEY=a'b":       {"KEY", "a'b"},
			"KEY=''":        {"KEY", ""},
			"KEY='a\\'b'":   {"KEY", "a\\'b"},

			"FOO=bar,QUUX=baz":     {"FOO", "bar", "QUUX", "baz"},
			"FOO='bar',QUUX=baz":   {"FOO", "bar", "QUUX", "baz"},
			"FOO=bar,QUUX='baz'":   {"FOO", "bar", "QUUX", "baz"},
			`FOO="bar",QUUX=baz`:   {"FOO", "bar", "QUUX", "baz"},
			`FOO=bar,QUUX="baz"`:   {"FOO", "bar", "QUUX", "baz"},
			`FOO="bar",QUUX='baz'`: {"FOO", "bar", "QUUX", "baz"},
			`FOO='bar',QUUX="baz"`: {"FOO", "bar", "QUUX", "baz"},

			`KEY="foo='bar',lorem=ipsum"`: {"KEY", "foo='bar',lorem=ipsum"},
			`FOO=bar,QUUX="lorem ipsum"`:  {"FOO", "bar", "QUUX", "lorem ipsum"},

			`JSON="{\"ID\":1,\"Name\":\"Reds\",\"Colors\":[\"Crimson\",\"Red\",\"Ruby\",\"Maroon\"]}"`: {"JSON", `{"ID":1,"Name":"Reds","Colors":["Crimson","Red","Ruby","Maroon"]}`},

			`JSON='{"ID":1,"Name":"Reds","Colors":["Crimson","Red","Ruby","Maroon"]}'`: {"JSON", `{"ID":1,"Name":"Reds","Colors":["Crimson","Red","Ruby","Maroon"]}`},
		}

		for in, exp := range tests {
			t.Run(in, func(t *testing.T) {
				got, err := parseMultienvLine(in)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				t.Logf("\n%q\n%q", exp, got)

				if e, g := len(exp), len(got); e != g {
					t.Fatalf("expecting %d elements, got %d", e, g)
				}

				for i, e := range exp {
					if g := got[i]; g != e {
						t.Errorf("expecting %q at index %d, got %q", e, i, g)
					}
				}
			})
		}
	})

	t.Run("error", func(t *testing.T) {
		tests := map[string]error{
			"BAD KEY":           errInvalidKeySyntax,
			"KEY='missingquote": errMisquoted,
			`KEY="missingquote`: errMisquoted,
			`KEY="missquoted'`:  errMisquoted,
			`KEY=a"b`:           errMisquoted,
			`KEY=value,rem`:     errRemaining,
		}

		for in, exp := range tests {
			t.Run(in, func(t *testing.T) {
				if _, err := parseMultienvLine(in); !errors.Is(err, exp) {
					t.Fatalf("expecting error %v, got %v", exp, err)
				}
			})
		}
	})
}
