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

	"github.com/hashicorp/go-retryablehttp"
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

// newMirrorHTTPClient scopes credentials to the mirror origin on every request,
// including redirects and retries. With no credentials it returns nil so that
// hc-install uses its default client.
func newMirrorHTTPClient(mirrorURL string, auth APIAuth) *http.Client {
	if auth.BearerToken == "" && auth.Username == "" {
		return nil
	}
	u, err := url.Parse(mirrorURL)
	if err != nil || u.Host == "" {
		return nil
	}
	return httpclient.New(func(client *retryablehttp.Client) {
		// This inner client follows redirects. Wrapping its transport checks
		// each destination while preserving hc-install's retries and User-Agent.
		switch {
		case auth.BearerToken != "":
			client.HTTPClient.Transport = &bearerTokenRoundTripper{
				origin: u,
				token:  auth.BearerToken,
				inner:  client.HTTPClient.Transport,
			}
		case auth.Username != "":
			client.HTTPClient.Transport = &basicAuthRoundTripper{
				origin:   u,
				username: auth.Username,
				password: auth.Password,
				inner:    client.HTTPClient.Transport,
			}
		}
	})
}

// sameMirrorOrigin compares scheme, hostname and effective port. Go's default
// redirect policy permits credentials across ports and to subdomains.
func sameMirrorOrigin(a, b *url.URL) bool {
	return strings.EqualFold(a.Scheme, b.Scheme) &&
		strings.EqualFold(a.Hostname(), b.Hostname()) && urlPort(a) == urlPort(b)
}

func urlPort(u *url.URL) string {
	if port := u.Port(); port != "" {
		return port
	}
	if strings.EqualFold(u.Scheme, "https") {
		return "443"
	}
	if strings.EqualFold(u.Scheme, "http") {
		return "80"
	}
	return ""
}

// bearerTokenRoundTripper scopes bearer auth to the configured mirror origin.
type bearerTokenRoundTripper struct {
	origin *url.URL
	token  string
	inner  http.RoundTripper
}

func (rt *bearerTokenRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	if !sameMirrorOrigin(req.URL, rt.origin) {
		req.Header.Del("Authorization")
		return rt.inner.RoundTrip(req)
	}
	req.Header.Set("Authorization", "Bearer "+rt.token)
	return rt.inner.RoundTrip(req)
}

// basicAuthRoundTripper scopes basic auth to the configured mirror origin.
type basicAuthRoundTripper struct {
	origin   *url.URL
	username string
	password string
	inner    http.RoundTripper
}

func (rt *basicAuthRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	if !sameMirrorOrigin(req.URL, rt.origin) {
		req.Header.Del("Authorization")
		return rt.inner.RoundTrip(req)
	}
	req.SetBasicAuth(rt.username, rt.password)
	return rt.inner.RoundTrip(req)
}

type Distribution interface {
	BinName() string
	Downloader() Downloader
	// ResolveConstraint gets the latest version for the given constraint
	ResolveConstraint(context.Context, string) (*version.Version, error)
}

// NewDistribution configures the default distribution and its project overrides.
// Terraform mirror settings are retained even when the default is OpenTofu.
func NewDistribution(distribution string, tfDownloadBaseURL string, apiAuth APIAuth) Distribution {
	downloadBaseURL := strings.TrimSpace(tfDownloadBaseURL)
	httpClient := newMirrorHTTPClient(downloadBaseURL, apiAuth)
	terraformDistribution := &DistributionTerraform{
		downloader:      &TerraformDownloader{httpClient: httpClient},
		downloadBaseURL: downloadBaseURL,
		httpClient:      httpClient,
	}
	opentofuDistribution := NewDistributionOpenTofu()
	var defaultDistribution Distribution = terraformDistribution
	if distribution == "opentofu" {
		defaultDistribution = opentofuDistribution
	}
	return &configuredDistributions{
		Distribution: defaultDistribution,
		terraform:    terraformDistribution,
		opentofu:     opentofuDistribution,
	}
}

// configuredDistributions exposes the default Distribution while retaining the
// configured alternatives for project-specific selection.
type configuredDistributions struct {
	Distribution
	terraform Distribution
	opentofu  Distribution
}

// SelectDistribution selects a project's distribution without dropping the
// server's mirror configuration. A nil override preserves the default.
func SelectDistribution(defaultDistribution Distribution, override *string) Distribution {
	if override == nil {
		return defaultDistribution
	}
	if configured, ok := defaultDistribution.(*configuredDistributions); ok {
		if *override == "opentofu" {
			return configured.opentofu
		}
		return configured.terraform
	}
	// Standalone distributions can supply custom downloaders. Preserve those
	// instances when the project selects the same distribution.
	binName := "terraform"
	if *override == "opentofu" {
		binName = "tofu"
	}
	if defaultDistribution != nil && defaultDistribution.BinName() == binName {
		return defaultDistribution
	}
	return NewDistribution(*override, "", APIAuth{})
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
