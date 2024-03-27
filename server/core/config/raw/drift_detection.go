package raw

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/runatlantis/atlantis/server/core/config/valid"
)

type DriftDetection struct {
	Enabled *bool   `yaml:"enabled,omitempty"`
	Cron    *string `yaml:"cron,omitempty"`
}

func (d DriftDetection) Validate() error {

	return validation.ValidateStruct(&d,
		validation.Field(&d.Cron, validation.By(CronValid)),
	)
}

func (d DriftDetection) ToValid() valid.DriftDetection {
	var v valid.DriftDetection

	if d.Enabled != nil {
		v.Enabled = *d.Enabled
	} else {
		v.Enabled = false
	}

	if d.Cron != nil {
		v.Cron = *d.Cron
	} else {
		v.Cron = *d.Cron
	}

	return v
}

func CronValid(value interface{}) error {
	//TODO: perform validation of cron strings
	return nil
}
