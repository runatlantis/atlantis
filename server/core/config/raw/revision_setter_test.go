package raw_test

import (
	"encoding/json"
	"testing"

	"github.com/runatlantis/atlantis/server/core/config/raw"
	"github.com/stretchr/testify/assert"
	yaml "gopkg.in/yaml.v2"
)

func TestPRRevision_Unmarshal(t *testing.T) {
	t.Run("yaml", func(t *testing.T) {

		rawYaml := `
url: https://test-url.com
basic_auth:
  username: test-user
  password: tes-password
`
		var result raw.RevisionSetter

		err := yaml.UnmarshalStrict([]byte(rawYaml), &result)
		assert.NoError(t, err)
	})

	t.Run("json", func(t *testing.T) {
		rawJSON := `
	{
		"url": "https://test-url.com",
		"basic_auth": {
			"username": "test-user",
			"password": "test-password"
		}
	}
	`
		var result raw.RevisionSetter

		err := json.Unmarshal([]byte(rawJSON), &result)
		assert.NoError(t, err)
	})

}

func TestPRRevision_Validate_Success(t *testing.T) {
	prRevision := &raw.RevisionSetter{
		URL: "https://test-url.com",
		BasicAuth: raw.BasicAuth{
			Username: "test-username",
			Password: "test-password",
		},
	}
	assert.NoError(t, prRevision.Validate())
}

func TestPRRevision_Validate_Error(t *testing.T) {
	cases := []struct {
		description string
		subject     raw.RevisionSetter
	}{
		{
			description: "missing basic auth",
			subject: raw.RevisionSetter{
				URL: "https://tes-url.com",
			},
		},
		{
			description: "missing password",
			subject: raw.RevisionSetter{
				URL: "https://tes-url.com",
				BasicAuth: raw.BasicAuth{
					Username: "test-username",
				},
			},
		},
		{
			description: "missing username",
			subject: raw.RevisionSetter{
				URL: "https://tes-url.com",
				BasicAuth: raw.BasicAuth{
					Password: "test-password",
				},
			},
		},
		{
			description: "missing url",
			subject: raw.RevisionSetter{
				BasicAuth: raw.BasicAuth{
					Username: "test-username",
					Password: "test-password",
				},
			},
		},
		{
			description: "invalid url",
			subject: raw.RevisionSetter{
				URL: "tes-url&^*&.com",
				BasicAuth: raw.BasicAuth{
					Username: "test-username",
					Password: "test-password",
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
