package terraform_test

import (
	"context"
	"testing"

	"github.com/runatlantis/atlantis/server/core/terraform"
	. "github.com/runatlantis/atlantis/testing"
)

func TestOpenTofuBinName(t *testing.T) {
	d := terraform.NewDistributionOpenTofu()
	Equals(t, d.BinName(), "tofu")
}

func TestResolveOpenTofuVersions(t *testing.T) {
	d := terraform.NewDistributionOpenTofu()
	version, err := d.ResolveConstraint(context.Background(), "= 1.8.0")
	Ok(t, err)
	Equals(t, version.String(), "1.8.0")
}

func TestTerraformBinName(t *testing.T) {
	d := terraform.NewDistributionTerraform()
	Equals(t, d.BinName(), "terraform")
}

func TestResolveTerraformVersions(t *testing.T) {
	d := terraform.NewDistributionTerraform()
	version, err := d.ResolveConstraint(context.Background(), "= 1.9.3")
	Ok(t, err)
	Equals(t, version.String(), "1.9.3")
}
