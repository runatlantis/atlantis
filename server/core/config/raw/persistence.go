package raw

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/graymeta/stow"
	stow_s3 "github.com/graymeta/stow/s3"
	"github.com/runatlantis/atlantis/server/core/config/valid"
)

type Persistence struct {
	DefaultStore DataStore `yaml:"default_store" json:"default_store"`

	DeploymentStorePrefix string `yaml:"deployment_store_prefix" json:"deployment_store_prefix"`
	JobStorePrefix        string `yaml:"job_store_prefix" json:"job_store_prefix"`
}

func (p Persistence) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.DefaultStore),
	)
}

func (p Persistence) ToValid() valid.PersistenceConfig {
	return valid.PersistenceConfig{
		Deployments: buildValidStore(p.DefaultStore, p.DeploymentStorePrefix),
		Jobs:        buildValidStore(p.DefaultStore, p.JobStorePrefix),
	}
}

func buildValidStore(dataStore DataStore, prefix string) valid.StoreConfig {
	var validStore valid.StoreConfig

	// Serially checks for non-nil supported backends
	switch {
	case dataStore.S3 != nil:
		validStore = valid.StoreConfig{
			ContainerName: dataStore.S3.BucketName,
			BackendType:   valid.S3Backend,
			Prefix:        prefix,
			// Hard coding iam auth type since we only support this for now
			Config: stow.ConfigMap{
				stow_s3.ConfigAuthType: "iam",
			},
		}
	}
	return validStore
}

type DataStore struct {
	S3 *S3 `yaml:"s3" json:"s3"`

	// Add other supported data stores in the future
}

func (ds DataStore) Validate() error {
	return validation.ValidateStruct(&ds, validation.Field(&ds.S3))
}

type S3 struct {
	BucketName string `yaml:"bucket-name" json:"bucket-name"`
}

func (s S3) Validate() error {
	return validation.ValidateStruct(&s,
		validation.Field(&s.BucketName, validation.Required),
	)
}
