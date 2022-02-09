package raw_test

import (
	"encoding/json"
	"testing"

	"github.com/runatlantis/atlantis/server/core/config/raw"
	"github.com/stretchr/testify/assert"
	yaml "gopkg.in/yaml.v2"
)

func TestJobs_Unmarshal(t *testing.T) {
	t.Run("yaml", func(t *testing.T) {

		rawYaml := `
storage-backend:
  s3:
    bucket-name: atlantis-test
`

		var result raw.Jobs

		err := yaml.UnmarshalStrict([]byte(rawYaml), &result)
		assert.NoError(t, err)
	})

	t.Run("json", func(t *testing.T) {
		rawJSON := `
	{
		"storage-backend": {
			"s3": {
				"bucket-name": "atlantis-test"
			}
		}
	}
	`

		var result raw.Jobs

		err := json.Unmarshal([]byte(rawJSON), &result)
		assert.NoError(t, err)
	})
}

func TestJobs_Validate_Success(t *testing.T) {
	cases := []struct {
		description string
		subject     raw.Jobs
	}{
		{
			description: "success",
			subject: raw.Jobs{
				StorageBackend: &raw.StorageBackend{
					S3: &raw.S3{
						BucketName: "test-bucket",
					},
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

func TestJobs_ValidateError(t *testing.T) {
	cases := []struct {
		description string
		subject     raw.Jobs
	}{
		{
			description: "bucket name not specified",
			subject: raw.Jobs{
				StorageBackend: &raw.StorageBackend{
					S3: &raw.S3{},
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
