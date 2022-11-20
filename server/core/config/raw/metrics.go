package raw

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/runatlantis/atlantis/server/core/config/valid"
)

type Metrics struct {
	Statsd     *Statsd     `yaml:"statsd" json:"statsd"`
	Prometheus *Prometheus `yaml:"prometheus" json:"prometheus"`
}

type Prometheus struct {
	Endpoint string `yaml:"endpoint" json:"endpoint"`
}

func (p *Prometheus) Validate() error {
	return validation.ValidateStruct(p, validation.Field(&p.Endpoint, validation.Required))
}

type Statsd struct {
	Port string `yaml:"port" json:"port"`
	Host string `yaml:"host" json:"host"`
}

func (s *Statsd) Validate() error {
	return validation.ValidateStruct(s,
		validation.Field(&s.Host, validation.Required),
		validation.Field(&s.Port, validation.Required),
		validation.Field(&s.Host, is.Host),
		validation.Field(&s.Port, is.Int))
}

func (m Metrics) Validate() error {
	res := validation.ValidateStruct(&m,
		validation.Field(&m.Statsd, validation.NilOrNotEmpty),
		validation.Field(&m.Prometheus, validation.NilOrNotEmpty),
	)
	return res
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
	if m.Prometheus != nil {
		return valid.Metrics{
			Prometheus: &valid.Prometheus{
				Endpoint: m.Prometheus.Endpoint,
			},
		}
	}
	return valid.Metrics{}
}
