package raw

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/runatlantis/atlantis/server/core/config/valid"
)

type RevisionSetter struct {
	URL       string    `yaml:"url" json:"url"`
	BasicAuth BasicAuth `yaml:"basic_auth" json:"basic_auth"`
}

func (p RevisionSetter) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.URL, validation.Required, is.URL),
		validation.Field(&p.BasicAuth),
	)
}

func (p *RevisionSetter) ToValid() valid.RevisionSetter {
	return valid.RevisionSetter{
		BasicAuth: p.BasicAuth.ToValid(),
		URL:       p.URL,
	}
}

type BasicAuth struct {
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`
}

func (b BasicAuth) Validate() error {
	return validation.ValidateStruct(&b,
		validation.Field(&b.Username, validation.Required),
		validation.Field(&b.Password, validation.Required),
	)
}

func (b BasicAuth) ToValid() valid.BasicAuth {
	return valid.BasicAuth{
		Username: b.Username,
		Password: b.Password,
	}
}
