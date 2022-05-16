package raw_test

import (
	"encoding/json"
	"testing"

	"github.com/runatlantis/atlantis/server/core/config/raw"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestMetrics_Unmarshal(t *testing.T) {
	t.Run("yaml", func(t *testing.T) {

		rawYaml := `
statsd:
  host: 127.0.0.1
  port: 8125
`

		var result raw.Metrics

		err := yaml.UnmarshalStrict([]byte(rawYaml), &result)
		assert.NoError(t, err)
	})

	t.Run("json", func(t *testing.T) {
		rawJSON := `
{
	"statsd": {
		"host": "127.0.0.1",
		"port": "8125"
	}
}		
`

		var result raw.Metrics

		err := json.Unmarshal([]byte(rawJSON), &result)
		assert.NoError(t, err)
	})
}
func TestMetrics_Validate_Success(t *testing.T) {

	cases := []struct {
		description string
		subject     raw.Metrics
	}{
		{
			description: "success",
			subject: raw.Metrics{
				Statsd: &raw.Statsd{
					Host: "127.0.0.1",
					Port: "8125",
				},
			},
		},
		{
			description: "missing stats",
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			assert.NoError(t, c.subject.Validate())
		})
	}

}
func TestMetrics_Validate_Error(t *testing.T) {
	cases := []struct {
		description string
		subject     raw.Metrics
	}{
		{
			description: "missing host",
			subject: raw.Metrics{
				Statsd: &raw.Statsd{
					Port: "8125",
				},
			},
		},
		{
			description: "missing port",
			subject: raw.Metrics{
				Statsd: &raw.Statsd{
					Host: "127.0.0.1",
				},
			},
		},
		{
			description: "invalid port",
			subject: raw.Metrics{
				Statsd: &raw.Statsd{
					Host: "127.0.0.1",
					Port: "string",
				},
			},
		},
		{
			description: "invalid host",
			subject: raw.Metrics{
				Statsd: &raw.Statsd{
					Host: "127.0.1",
					Port: "8125",
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			assert.Error(t, c.subject.Validate())
		})
	}
}
