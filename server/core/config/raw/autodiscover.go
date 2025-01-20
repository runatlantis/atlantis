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

	var ignorePaths []*regexp.Regexp
	if a.IgnorePaths != nil {
		for _, ignore := range a.IgnorePaths {
			withoutSlashes := ignore[1 : len(ignore)-1]
			// Safe to use MustCompile because we test it in Validate().
			ignorePaths = append(ignorePaths, regexp.MustCompile(withoutSlashes))
		}
		v.IgnorePaths = ignorePaths
	}

	return &v
}

func (a AutoDiscover) Validate() error {

	ignoreValid := func(value interface{}) error {
		strSlice := value.([]string)
		if strSlice == nil {
			return nil
		}
		for _, ignore := range strSlice {
			if !strings.HasPrefix(ignore, "/") || !strings.HasSuffix(ignore, "/") {
				return errors.New("regex must begin and end with a slash '/'")
			}
			withoutSlashes := ignore[1 : len(ignore)-1]
			_, err := regexp.Compile(withoutSlashes)
			if err != nil {
				return errors.Wrapf(err, "parsing: %s", ignore)
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
