// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package terraform

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/opentofu/tofudl"
)

// APIAuth holds optional HTTP credentials for a custom tf-download-url mirror.
type APIAuth struct {
	Username    string
	Password    string
	BearerToken string
}

// transport builds an http.RoundTripper wrapper that injects these
// credentials into every outgoing request, suitable for
// releases.Versions.Transport / releases.ExactVersion.Transport. It returns
// nil when no credentials are configured, so hc-install's default transport
// is used unmodified.
func (a APIAuth) transport() func(http.RoundTripper) http.RoundTripper {
	if a.BearerToken == "" && a.Username == "" {
		return nil
	}
	return func(next http.RoundTripper) http.RoundTripper {
		return &apiAuthRoundTripper{next: next, auth: a}
	}
}

type apiAuthRoundTripper struct {
	next http.RoundTripper
	auth APIAuth
}

func (rt *apiAuthRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	switch {
	case rt.auth.BearerToken != "":
		req.Header.Set("Authorization", "Bearer "+rt.auth.BearerToken)
	case rt.auth.Username != "":
		req.SetBasicAuth(rt.auth.Username, rt.auth.Password)
	}
	return rt.next.RoundTrip(req)
}

type Distribution interface {
	BinName() string
	Downloader() Downloader
	// ResolveConstraint gets the latest version for the given constraint
	ResolveConstraint(context.Context, string) (*version.Version, error)
}

// NewDistribution returns the distribution implementation for Atlantis.
// tfDownloadBaseURL is used for Terraform release listing and installs when
// distribution is terraform (e.g. --tf-download-url); it is ignored for OpenTofu.
// apiAuth carries optional credentials for a custom mirror; a zero value
// means no auth (the hc-install default).
func NewDistribution(distribution string, tfDownloadBaseURL string, apiAuth APIAuth) Distribution {
	if distribution == "opentofu" {
		return NewDistributionOpenTofu()
	}
	return &DistributionTerraform{
		downloader:      &TerraformDownloader{apiAuth: apiAuth},
		downloadBaseURL: strings.TrimSpace(tfDownloadBaseURL),
		apiAuth:         apiAuth,
	}
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
	downloader      Downloader
	downloadBaseURL string
	apiAuth         APIAuth
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

func (d *DistributionTerraform) ResolveConstraint(ctx context.Context, constraintStr string) (*version.Version, error) {
	vc, err := version.NewConstraint(constraintStr)
	if err != nil {
		return nil, fmt.Errorf("error parsing constraint string: %s", err)
	}

	constrainedVersions := &releases.Versions{
		Product:     product.Terraform,
		Constraints: vc,
		Transport:   d.apiAuth.transport(),
	}
	if d.downloadBaseURL != "" {
		constrainedVersions.ApiBaseURL = d.downloadBaseURL
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
