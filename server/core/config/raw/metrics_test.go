package raw_test

import (
	"encoding/json"
	"testing"

	"github.com/runatlantis/atlantis/server/core/config/raw"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestMetrics_Unmarshal(t *testing.T) {
	t.Run("yaml", func(t *testing.T) {

		rawYaml := `
statsd:
  host: 127.0.0.1
  port: 8125
prometheus:
  endpoint: /metrics
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
	},
	"prometheus": {
		"endpoint": "/metrics"
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
			description: "success with stats config",
			subject: raw.Metrics{
				Statsd: &raw.Statsd{
					Host: "127.0.0.1",
					Port: "8125",
				},
			},
		},
		{
			description: "success with stats config using hostname",
			subject: raw.Metrics{
				Statsd: &raw.Statsd{
					Host: "localhost",
					Port: "8125",
				},
			},
		},
		{
			description: "missing stats",
		},
		{
			description: "success with prometheus config",
			subject: raw.Metrics{
				Prometheus: &raw.Prometheus{
					Endpoint: "/metrics",
				},
			},
		},
		{
			description: "success with both configs",
			subject: raw.Metrics{
				Statsd: &raw.Statsd{
					Host: "127.0.0.1",
					Port: "8125",
				},
				Prometheus: &raw.Prometheus{
					Endpoint: "/metrics",
				},
			},
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
			description: "invalid endpoint",
			subject: raw.Metrics{
				Prometheus: &raw.Prometheus{
					Endpoint: "",
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
