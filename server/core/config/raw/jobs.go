package raw

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/runatlantis/atlantis/server/core/config/valid"
)

type Jobs struct {
	StorageBackend *StorageBackend `yaml:"storage-backend" json:"storage-backend"`
}

type StorageBackend struct {
	S3 *S3 `yaml:"s3" json:"s3"`
}

type S3 struct {
	BucketName string `yaml:"bucket-name" json:"bucket-name"`
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

func (s S3) Validate() error {
	return validation.ValidateStruct(&s,
		validation.Field(&s.BucketName, validation.Required),
	)
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
