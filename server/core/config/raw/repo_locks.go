package raw

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/runatlantis/atlantis/server/core/config/valid"
)

type RepoLocks struct {
	Mode *valid.RepoLocksMode `yaml:"mode,omitempty"`
}

func (a RepoLocks) ToValid() *valid.RepoLocks {
	var v valid.RepoLocks

	if a.Mode != nil {
		v.Mode = *a.Mode
	} else {
		v.Mode = valid.DefaultRepoLocksMode
	}

	return &v
}

func (a RepoLocks) Validate() error {
	res := validation.ValidateStruct(&a,
		// If a.Mode is nil, this should still pass validation.
		validation.Field(&a.Mode, validation.In(valid.RepoLocksDisabledMode, valid.RepoLocksOnPlanMode, valid.RepoLocksOnApplyMode)),
	)
	return res
}
