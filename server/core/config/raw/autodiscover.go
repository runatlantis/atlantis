package raw

import (
	"regexp"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/core/config/valid"
)

var DefaultAutoDiscoverMode = valid.AutoDiscoverAutoMode

type AutoDiscover struct {
	Mode   *valid.AutoDiscoverMode `yaml:"mode,omitempty"`
	Ignore *string                 `yaml:"ignore,omitempty"`
}

func (a AutoDiscover) ToValid() *valid.AutoDiscover {
	var v valid.AutoDiscover

	if a.Mode != nil {
		v.Mode = *a.Mode
	} else {
		v.Mode = DefaultAutoDiscoverMode
	}

	if a.Ignore != nil {
		ignore := *a.Ignore
		withoutSlashes := ignore[1 : len(ignore)-1]
		// Safe to use MustCompile because we test it in Validate().
		v.Ignore = regexp.MustCompile(withoutSlashes)
	}

	return &v
}

func (a AutoDiscover) Validate() error {

	ignoreValid := func(value interface{}) error {
		strPtr := value.(*string)
		if strPtr == nil {
			return nil
		}
		ignore := *strPtr
		if !strings.HasPrefix(ignore, "/") || !strings.HasSuffix(ignore, "/") {
			return errors.New("regex must begin and end with a slash '/'")
		}
		withoutSlashes := ignore[1 : len(ignore)-1]
		_, err := regexp.Compile(withoutSlashes)
		return errors.Wrapf(err, "parsing: %s", ignore)
	}

	res := validation.ValidateStruct(&a,
		// If a.Mode is nil, this should still pass validation.
		validation.Field(&a.Mode, validation.In(valid.AutoDiscoverAutoMode, valid.AutoDiscoverDisabledMode, valid.AutoDiscoverEnabledMode)),
		validation.Field(&a.Ignore, validation.By(ignoreValid)),
	)
	return res
}

func DefaultAutoDiscover() *valid.AutoDiscover {
	return &valid.AutoDiscover{
		Mode:   DefaultAutoDiscoverMode,
		Ignore: nil,
	}
}
