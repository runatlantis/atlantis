package external

import validation "github.com/go-ozzo/ozzo-validation"

type Repo struct {
	Owner string
	Name  string
}

func (r Repo) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Owner, validation.Required),
		validation.Field(&r.Name, validation.Required),
	)
}

type DeployRequest struct {
	Roots []string
	Repo  Repo
}

func (r DeployRequest) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Roots, validation.Required, validation.Length(1, 0)),
		validation.Field(&r.Repo, validation.Required),
	)
}
