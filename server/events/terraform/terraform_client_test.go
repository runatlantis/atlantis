package terraform_test

import (
	"testing"

	"github.com/atlantisnorth/atlantis/server/events/terraform"
	. "github.com/atlantisnorth/atlantis/testing"
	"github.com/hashicorp/go-version"
)

func TestMustConstraint_PancisOnBadConstraint(t *testing.T) {
	t.Log("MustConstraint should panic on a bad constraint")
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	terraform.MustConstraint("invalid constraint")
}

func TestMustConstraint(t *testing.T) {
	t.Log("MustConstraint should return the constrain")
	c := terraform.MustConstraint(">0.1")
	expectedConstraint, err := version.NewConstraint(">0.1")
	Ok(t, err)
	Equals(t, expectedConstraint.String(), c.String())
}
