package raw

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/runatlantis/atlantis/server/events/yaml/valid"
)

type Metrics struct {
	Statsd *Statsd `yaml:"statsd" json:"statsd"`
}

type Statsd struct {
	Port string `yaml:"port" json:"port"`
	Host string `yaml:"host" json:"host"`
}

func (s *Statsd) Validate() error {
	return validation.ValidateStruct(s,
		validation.Field(&s.Host, validation.Required),
		validation.Field(&s.Port, validation.Required),
		validation.Field(&s.Host, is.IP),
		validation.Field(&s.Port, is.Int))
}

func (m Metrics) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.Statsd),
	)
}

func (m Metrics) ToValid() valid.Metrics {
	// we've already validated at this point
	if m.Statsd != nil {
		return valid.Metrics{
			Statsd: &valid.Statsd{
				Host: m.Statsd.Host,
				Port: m.Statsd.Port,
			},
		}
	}

	return valid.Metrics{}
}
