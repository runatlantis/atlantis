package raw_test

import (
	"encoding/json"
	"testing"

	"github.com/runatlantis/atlantis/server/core/config/raw"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestTemporal_Unmarshal(t *testing.T) {
	t.Run("yaml", func(t *testing.T) {

		rawYaml := `
host: 127.0.0.1
port: 1234
`

		var result raw.Temporal

		err := yaml.UnmarshalStrict([]byte(rawYaml), &result)
		assert.NoError(t, err)
	})

	t.Run("json", func(t *testing.T) {
		rawJSON := `
{
	"host": "127.0.0.1",
	"port": "1234"
}		
`

		var result raw.Temporal

		err := json.Unmarshal([]byte(rawJSON), &result)
		assert.NoError(t, err)
	})
}

func TestTemporal_Validate_Success(t *testing.T) {

	cases := []struct {
		description string
		subject     *raw.Temporal
	}{
		{
			description: "success",
			subject: &raw.Temporal{
				Host:               "127.0.0.1",
				Port:               "8125",
				TerraformTaskQueue: "taskqueue",
			},
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			assert.NoError(t, c.subject.Validate())
		})
	}

}

func TestTemporal_Validate_Error(t *testing.T) {
	cases := []struct {
		description string
		subject     raw.Temporal
	}{
		{
			description: "missing host",
			subject: raw.Temporal{
				Port: "8125",
			},
		},
		{
			description: "missing port",
			subject: raw.Temporal{
				Host: "127.0.0.1",
			},
		},
		{
			description: "invalid port",
			subject: raw.Temporal{
				Host: "127.0.0.1",
				Port: "string",
			},
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			assert.Error(t, c.subject.Validate())
		})
	}
}
