package raw

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/runatlantis/atlantis/server/core/config/valid"
)

type Temporal struct {
	Port string `yaml:"port" json:"port"`
	Host string `yaml:"host" json:"host"`
}

func (t *Temporal) Validate() error {
	return validation.ValidateStruct(t,
		validation.Field(&t.Host, validation.Required),
		validation.Field(&t.Port, validation.Required),
		validation.Field(&t.Host, is.IP),
		validation.Field(&t.Port, is.Int))
}

func (t *Temporal) ToValid() valid.Temporal {
	return valid.Temporal{
		Host: t.Host,
		Port: t.Port,
	}
}
