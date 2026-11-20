// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package terraform_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/go-version"
	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/cmd"
	"github.com/runatlantis/atlantis/server/core/terraform"
	. "github.com/runatlantis/atlantis/testing"
)

func TestTerraformInstall(t *testing.T) {
	d := &terraform.TerraformDownloader{}
	RegisterMockTestingT(t)
	binDir := t.TempDir()

	v, _ := version.NewVersion("1.8.1")

	newPath, err := d.Install(context.Background(), binDir, cmd.DefaultTFDownloadURL, v)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if _, err := os.Stat(newPath); os.IsNotExist(err) {
		t.Errorf("Binary not found at %s", newPath)
	}
}

func TestOpenTofuInstall(t *testing.T) {
	d := &terraform.TofuDownloader{}
	RegisterMockTestingT(t)
	binDir := t.TempDir()

	v, _ := version.NewVersion("1.8.0")

	newPath, err := d.Install(context.Background(), binDir, cmd.DefaultTFDownloadURL, v)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if _, err := os.Stat(newPath); os.IsNotExist(err) {
		t.Errorf("Binary not found at %s", newPath)
	}
}

// Like TestTerraformInstall, this uses a signed upstream release so the full
// installer, including checksum/signature verification and extraction, is tested.
func TestTerraformInstall_AuthenticatedMirrorWithProjectOverride(t *testing.T) {
	upstream, err := url.Parse("https://releases.hashicorp.com")
	Ok(t, err)
	for _, defaultName := range []string{"terraform", "opentofu"} {
		for _, auth := range []struct {
			name   string
			config terraform.APIAuth
			header string
		}{
			{"bearer", terraform.APIAuth{BearerToken: "mirror-token"}, "Bearer mirror-token"},
			{"basic", terraform.APIAuth{Username: "user", Password: "pass"}, "Basic dXNlcjpwYXNz"},
		} {
			t.Run(defaultName+"/"+auth.name, func(t *testing.T) {
				var paths sync.Map
				proxy := &httputil.ReverseProxy{
					Rewrite: func(request *httputil.ProxyRequest) {
						request.SetURL(upstream)
						request.Out.Header.Del("Authorization")
					},
				}
				mirror := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.Header.Get("Authorization") != auth.header {
						t.Error("mirror download request was not authenticated")
						http.Error(w, "unauthorized", http.StatusUnauthorized)
						return
					}
					if !strings.HasPrefix(r.UserAgent(), "hc-install/") {
						t.Errorf("hc-install User-Agent missing: %q", r.UserAgent())
					}
					paths.Store(r.URL.Path, true)
					proxy.ServeHTTP(w, r)
				}))
				t.Cleanup(mirror.Close)
				configured := terraform.NewDistribution(defaultName, mirror.URL, auth.config)
				projectName := "terraform"
				distribution := terraform.SelectDistribution(configured, &projectName)
				v, err := version.NewVersion("1.8.1")
				Ok(t, err)
				ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
				defer cancel()
				path, err := distribution.Downloader().Install(ctx, t.TempDir(), mirror.URL, v)
				Ok(t, err)
				info, err := os.Stat(path)
				Ok(t, err)
				Assert(t, info.Size() > 0, "installed Terraform binary must not be empty")
				for _, suffix := range []string{"/index.json", "_SHA256SUMS", ".sig", ".zip"} {
					found := false
					paths.Range(func(key, _ any) bool {
						found = found || strings.HasSuffix(key.(string), suffix)
						return !found
					})
					Assert(t, found, "mirror must authenticate download stage %s", suffix)
				}
			})
		}
	}
}
