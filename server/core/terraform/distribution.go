// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package terraform

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/httpclient"
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

// newMirrorHTTPClient returns an http.Client for hc-install which injects
// auth credentials into requests made to the host of mirrorURL only, so they
// are never sent to any other host.
//
// It returns nil when no credentials are configured, or no mirror host can be
// determined, so that hc-install falls back to its default client.
func newMirrorHTTPClient(mirrorURL string, auth APIAuth) *http.Client {
	if auth.BearerToken == "" && auth.Username == "" {
		return nil
	}
	u, err := url.Parse(mirrorURL)
	if err != nil || u.Host == "" {
		return nil
	}

	client := httpclient.New()
	client.Transport = &apiAuthRoundTripper{
		host:  u.Host,
		auth:  auth,
		inner: client.Transport,
	}
	return client
}

type apiAuthRoundTripper struct {
	host  string
	auth  APIAuth
	inner http.RoundTripper
}

func (rt *apiAuthRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Host != rt.host {
		return rt.inner.RoundTrip(req)
	}

	req = req.Clone(req.Context())
	switch {
	case rt.auth.BearerToken != "":
		req.Header.Set("Authorization", "Bearer "+rt.auth.BearerToken)
	case rt.auth.Username != "":
		req.SetBasicAuth(rt.auth.Username, rt.auth.Password)
	}
	return rt.inner.RoundTrip(req)
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
// apiAuth carries optional credentials for the tfDownloadBaseURL mirror; a
// zero value means no auth (the hc-install default).
func NewDistribution(distribution string, tfDownloadBaseURL string, apiAuth APIAuth) Distribution {
	if distribution == "opentofu" {
		return NewDistributionOpenTofu()
	}
	downloadBaseURL := strings.TrimSpace(tfDownloadBaseURL)
	httpClient := newMirrorHTTPClient(downloadBaseURL, apiAuth)
	return &DistributionTerraform{
		downloader:      &TerraformDownloader{httpClient: httpClient},
		downloadBaseURL: downloadBaseURL,
		httpClient:      httpClient,
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
	httpClient      *http.Client
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
		ApiBaseURL:  d.downloadBaseURL,
		HTTPClient:  d.httpClient,
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
