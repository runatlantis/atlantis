package gitlab

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestValidate(t *testing.T) {
	testCases := []struct {
		description string
		content     string
		response    string
		want        *LintResult
	}{
		{
			description: "valid",
			content: `
				build1:
					stage: build
					script:
						- echo "Do your build here"`,
			response: `{
				"status": "valid",
				"errors": []
			}`,
			want: &LintResult{
				Status: "valid",
				Errors: []string{},
			},
		},
		{
			description: "invalid",
			content: `
				build1:
					- echo "Do your build here"`,
			response: `{
				"status": "invalid",
				"errors": ["error message when content is invalid"]
			}`,
			want: &LintResult{
				Status: "invalid",
				Errors: []string{"error message when content is invalid"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			mux, server, client := setup()
			defer teardown(server)

			mux.HandleFunc("/api/v4/ci/lint", func(w http.ResponseWriter, r *http.Request) {
				testMethod(t, r, "POST")
				fmt.Fprint(w, tc.response)
			})

			got, _, err := client.Validate.Lint(tc.content)

			if err != nil {
				t.Errorf("Validate returned error: %v", err)
			}

			want := tc.want
			if !reflect.DeepEqual(got, want) {
				t.Errorf("Validate returned \ngot:\n%v\nwant:\n%v", Stringify(got), Stringify(want))
			}
		})
	}
}
