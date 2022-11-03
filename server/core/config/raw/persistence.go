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

func (p Persistence) ToValid(defaultCfg valid.GlobalCfg) valid.PersistenceConfig {
	deployments := buildValidStore(p.DefaultStore, p.DeploymentStorePrefix, defaultCfg.PersistenceConfig.Deployments)
	jobs := buildValidStore(p.DefaultStore, p.JobStorePrefix, defaultCfg.PersistenceConfig.Jobs)

	return valid.PersistenceConfig{
		Deployments: deployments,
		Jobs:        jobs,
	}
}

func buildValidStore(dataStore DataStore, prefix string, defaultCfg valid.StoreConfig) valid.StoreConfig {
	// Serially checks for non-nil supported backends
	switch {
	case dataStore.S3 != nil:
		return valid.StoreConfig{
			ContainerName: dataStore.S3.BucketName,
			BackendType:   valid.S3Backend,
			Prefix:        prefix,
			// Hard coding iam auth type since we only support this for now
			Config: stow.ConfigMap{
				stow_s3.ConfigAuthType: "iam",
			},
		}
	default:
		return defaultCfg
	}
}

type DataStore struct {
	S3 *S3 `yaml:"s3" json:"s3"`
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
