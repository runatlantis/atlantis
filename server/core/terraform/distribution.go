package terraform

import (
	"context"
	"fmt"
	"sort"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/opentofu/tofudl"
)

type Distribution interface {
	BinName() string
	Downloader() Downloader
	// ResolveConstraint gets the latest version for the given constraint
	ResolveConstraint(context.Context, string) (*version.Version, error)
}

func NewDistribution(distribution string) Distribution {
	tfDistribution := NewDistributionTerraform()
	if distribution == "opentofu" {
		tfDistribution = NewDistributionOpenTofu()
	}
	return tfDistribution
}

type DistributionOpenTofu struct {
	downloader Downloader
}

func NewDistributionOpenTofu() Distribution {
	return &DistributionOpenTofu{
		downloader: &TofuDownloader{},
	}
}

func NewDistributionOpenTofuWithDownloader(downloader Downloader) Distribution {
	return &DistributionOpenTofu{
		downloader: downloader,
	}
}

func (*DistributionOpenTofu) BinName() string {
	return "tofu"
}

func (d *DistributionOpenTofu) Downloader() Downloader {
	return d.downloader
}

func (*DistributionOpenTofu) ResolveConstraint(ctx context.Context, constraintStr string) (*version.Version, error) {
	dl, err := tofudl.New()
	if err != nil {
		return nil, err
	}

	vc, err := version.NewConstraint(constraintStr)
	if err != nil {
		return nil, fmt.Errorf("error parsing constraint string: %s", err)
	}

	allVersions, err := dl.ListVersions(ctx)
	if err != nil {
		return nil, fmt.Errorf("error listing OpenTofu versions: %s", err)
	}

	var versions []*version.Version
	for _, ver := range allVersions {
		v, err := version.NewVersion(string(ver.ID))
		if err != nil {
			return nil, err
		}

		if vc.Check(v) {
			versions = append(versions, v)
		}
	}
	sort.Sort(version.Collection(versions))

	if len(versions) == 0 {
		return nil, fmt.Errorf("no OpenTofu versions found for constraints %s", constraintStr)
	}

	// We want to select the highest version that satisfies the constraint.
	version := versions[len(versions)-1]

	// Get the Version object from the versionDownloader.
	return version, nil
}

type DistributionTerraform struct {
	downloader Downloader
}

func NewDistributionTerraform() Distribution {
	return &DistributionTerraform{
		downloader: &TerraformDownloader{},
	}
}

func NewDistributionTerraformWithDownloader(downloader Downloader) Distribution {
	return &DistributionTerraform{
		downloader: downloader,
	}
}

func (*DistributionTerraform) BinName() string {
	return "terraform"
}

func (d *DistributionTerraform) Downloader() Downloader {
	return d.downloader
}

func (*DistributionTerraform) ResolveConstraint(ctx context.Context, constraintStr string) (*version.Version, error) {
	vc, err := version.NewConstraint(constraintStr)
	if err != nil {
		return nil, fmt.Errorf("error parsing constraint string: %s", err)
	}

	constrainedVersions := &releases.Versions{
		Product:     product.Terraform,
		Constraints: vc,
	}

	installCandidates, err := constrainedVersions.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("error listing available versions: %s", err)
	}
	if len(installCandidates) == 0 {
		return nil, fmt.Errorf("no Terraform versions found for constraints %s", constraintStr)
	}

	// We want to select the highest version that satisfies the constraint.
	versionDownloader := installCandidates[len(installCandidates)-1]

	// Get the Version object from the versionDownloader.
	return versionDownloader.(*releases.ExactVersion).Version, nil
}
