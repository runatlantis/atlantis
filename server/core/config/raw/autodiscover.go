package raw

import (
	"errors"
	"fmt"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/core/config/valid"
)

var DefaultAutoDiscoverMode = valid.AutoDiscoverAutoMode

type AutoDiscover struct {
	Mode        *valid.AutoDiscoverMode `yaml:"mode,omitempty"`
	IgnorePaths []string                `yaml:"ignore_paths,omitempty"`
}

func (a AutoDiscover) ToValid() *valid.AutoDiscover {
	var v valid.AutoDiscover

	if a.Mode != nil {
		v.Mode = *a.Mode
	} else {
		v.Mode = DefaultAutoDiscoverMode
	}

	v.IgnorePaths = a.IgnorePaths

	return &v
}

func (a AutoDiscover) Validate() error {

	ignoreValid := func(value interface{}) error {
		strSlice := value.([]string)
		if strSlice == nil {
			return nil
		}
		for _, ignore := range strSlice {
			// A beginning slash isn't necessary since they are specifying a relative path, not an absolute one.
			// Rejecting `/...` also allows us to potentially use `/.*/` as regexes in the future
			if strings.HasPrefix(ignore, "/") {
				return errors.New("pattern must not begin with a slash '/'")
			}

			if !doublestar.ValidatePattern(ignore) {
				return fmt.Errorf("invalid pattern: %s", ignore)
			}

		}
		return nil
	}

	res := validation.ValidateStruct(&a,
		// If a.Mode is nil, this should still pass validation.
		validation.Field(&a.Mode, validation.In(valid.AutoDiscoverAutoMode, valid.AutoDiscoverDisabledMode, valid.AutoDiscoverEnabledMode)),
		validation.Field(&a.IgnorePaths, validation.By(ignoreValid)),
	)
	return res
}

func DefaultAutoDiscover() *valid.AutoDiscover {
	return &valid.AutoDiscover{
		Mode:        DefaultAutoDiscoverMode,
		IgnorePaths: nil,
	}
}
