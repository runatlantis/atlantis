package raw

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/runatlantis/atlantis/server/core/config/valid"
)

type Temporal struct {
	Port            string `yaml:"port" json:"port"`
	Host            string `yaml:"host" json:"host"`
	UseSystemCACert bool   `yaml:"us_system_ca_cert" json:"us_system_ca_cert"`
	Namespace       string `yaml:"namespace" json:"namespace"`
}

func (t *Temporal) Validate() error {
	return validation.ValidateStruct(t,
		validation.Field(&t.Host, validation.Required),
		validation.Field(&t.Port, validation.Required),
		validation.Field(&t.Port, is.Int))
}

func (t *Temporal) ToValid() valid.Temporal {
	return valid.Temporal{
		Host:            t.Host,
		Port:            t.Port,
		UseSystemCACert: t.UseSystemCACert,
		Namespace:       t.Namespace,
	}
}
