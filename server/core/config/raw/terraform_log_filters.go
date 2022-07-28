package raw

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"regexp"
)

// TerraformLogFilters is the raw schema for repo-level atlantis.yaml config.
type TerraformLogFilters struct {
	Regexes []string `yaml:"regexes,omitempty" json:"regexes,omitempty"`
}

func (t TerraformLogFilters) ToValid() valid.TerraformLogFilters {
	terraformLogFilters := valid.TerraformLogFilters{}
	if len(t.Regexes) > 0 {
		var regexes []*regexp.Regexp
		for _, regexString := range t.Regexes {
			//already validated compile should work
			regex, _ := regexp.Compile(regexString)
			regexes = append(regexes, regex)
		}
		terraformLogFilters.Regexes = regexes
	}
	return terraformLogFilters
}

func (t TerraformLogFilters) Validate() error {
	return validation.ValidateStruct(&t,
		validation.Field(&t.Regexes, validation.By(regexesCompile)))
}

func regexesCompile(value interface{}) error {
	regexStrings := value.([]string)
	for _, regexString := range regexStrings {
		regex, err := regexp.Compile(regexString)
		if err != nil {
			return errors.Wrapf(err, "invalid log filter regex: %s", regex)
		}
	}
	return nil
}
