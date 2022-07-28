package raw_test

import (
	"testing"

	"github.com/runatlantis/atlantis/server/core/config/raw"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestTerraformLogFilters_Unmarshal(t *testing.T) {
	t.Run("yaml", func(t *testing.T) {

		rawYaml := `
regexes: [abc, xyz, 123]
`

		var result raw.TerraformLogFilters

		err := yaml.UnmarshalStrict([]byte(rawYaml), &result)
		assert.NoError(t, err)
	})
}

func TestTerraformLogFilters_Validate_Success(t *testing.T) {

	cases := []struct {
		description string
		subject     raw.TerraformLogFilters
	}{
		{
			description: "success",
			subject: raw.TerraformLogFilters{
				Regexes: []string{"abc*", "[A-Z]test"},
			},
		},
		{
			description: "no regexes",
			subject: raw.TerraformLogFilters{
				Regexes: []string{},
			},
		},
		{
			description: "empty",
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			assert.NoError(t, c.subject.Validate())
		})
	}

}

func TestTerraformLogFilters_Validate_Error(t *testing.T) {
	cases := []struct {
		description string
		subject     raw.TerraformLogFilters
	}{
		{
			description: "invalid regex",
			subject: raw.TerraformLogFilters{
				Regexes: []string{"[A-==Z]test"},
			},
		},
		{
			description: "partial regex",
			subject: raw.TerraformLogFilters{
				Regexes: []string{"abc*", "[A-==Z]test"},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			assert.Error(t, c.subject.Validate())
		})
	}
}
