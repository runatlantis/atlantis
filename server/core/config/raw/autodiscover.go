package raw

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/runatlantis/atlantis/server/core/config/valid"
)

var DefaultAutoDiscoverMode = valid.AutoDiscoverAutoMode

type AutoDiscover struct {
	Mode *valid.AutoDiscoverMode `yaml:"mode,omitempty"`
}

func (a AutoDiscover) ToValid() *valid.AutoDiscover {
	var v valid.AutoDiscover

	if a.Mode != nil {
		v.Mode = *a.Mode
	} else {
		v.Mode = DefaultAutoDiscoverMode
	}

	return &v
}

func (a AutoDiscover) Validate() error {
	res := validation.ValidateStruct(&a,
		// If a.Mode is nil, this should still pass validation.
		validation.Field(&a.Mode, validation.In(valid.AutoDiscoverAutoMode, valid.AutoDiscoverDisabledMode, valid.AutoDiscoverEnabledMode)),
	)
	return res
}

func DefaultAutoDiscover() *valid.AutoDiscover {
	return &valid.AutoDiscover{
		Mode: DefaultAutoDiscoverMode,
	}
}
