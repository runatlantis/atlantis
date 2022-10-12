package raw

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/graymeta/stow"
	stow_s3 "github.com/graymeta/stow/s3"
	"github.com/runatlantis/atlantis/server/core/config/valid"
)

type Jobs struct {
	StorageBackend *StorageBackend `yaml:"storage-backend" json:"storage-backend"`
}

type StorageBackend struct {
	S3 *S3 `yaml:"s3" json:"s3"`
}

func (j Jobs) Validate() error {
	return validation.ValidateStruct(&j,
		validation.Field(&j.StorageBackend),
	)
}

func (s StorageBackend) Validate() error {
	return validation.ValidateStruct(&s,
		validation.Field(&s.S3),
	)
}

// Switch through all possible storage backends and set the non-nil one as the
// confgured backend
func (s *StorageBackend) ToValid() valid.StorageBackend {
	switch {
	case s.S3 != nil:
		return valid.StorageBackend{
			BackendConfig: &valid.S3{
				BucketName: s.S3.BucketName,
			},
		}
	default:
		return valid.StorageBackend{}
	}
}

func (j *Jobs) ToValid() valid.Jobs {
	if j.StorageBackend == nil {
		return valid.Jobs{}
	}

	storageBackend := j.StorageBackend.ToValid()
	return valid.Jobs{
		StorageBackend: &storageBackend,
	}
}

func (j *Jobs) ToStoreConfig() valid.StoreConfig {
	return valid.StoreConfig{
		ContainerName: j.StorageBackend.S3.BucketName,
		BackendType:   valid.S3Backend,
		Config: stow.ConfigMap{
			stow_s3.ConfigAuthType: "iam",
		},
		Prefix: "ouptut",
	}
}
