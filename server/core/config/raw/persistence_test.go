package raw_test

import (
	"encoding/json"
	"testing"

	"github.com/runatlantis/atlantis/server/core/config/raw"
	"github.com/stretchr/testify/assert"
	yaml "gopkg.in/yaml.v2"
)

func TestPersistence_Unmarshal(t *testing.T) {
	t.Run("yaml", func(t *testing.T) {

		rawYaml := `
job_store_prefix: jobs
deployment_store_prefix: deployments
default_store:
  s3:
    bucket-name: atlantis-test
`

		var result raw.Persistence

		err := yaml.UnmarshalStrict([]byte(rawYaml), &result)
		assert.NoError(t, err)
	})

	t.Run("json", func(t *testing.T) {
		rawJSON := `
	{
		"job_store_prefix": "jobs",
		"deployment_store–prefix": "deployments",
		"default_store": {
			"s3": {
				"bucket-name": "test-bucket"
			}
		}
	}
	`
		var result raw.Persistence

		err := json.Unmarshal([]byte(rawJSON), &result)
		assert.NoError(t, err)
	})

}

func TestPersistence_ValidateSuccess(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		assert.NoError(t, raw.Persistence{
			JobStorePrefix:        "jobs",
			DeploymentStorePrefix: "deployments",
			DefaultStore: raw.DataStore{
				S3: &raw.S3{
					BucketName: "test-bucket",
				},
			},
		}.Validate())
	})

}

func TestPersistence_ValidateError(t *testing.T) {
	t.Run("bucket name not configured", func(t *testing.T) {
		assert.Error(t, raw.Persistence{
			JobStorePrefix:        "jobs",
			DeploymentStorePrefix: "deployment",
			DefaultStore: raw.DataStore{
				S3: &raw.S3{},
			},
		}.Validate())
	})
}
