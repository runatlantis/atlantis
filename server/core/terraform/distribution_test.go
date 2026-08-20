// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package terraform_test

import (
	"context"
	"net/http"
	"net/http/httptest"
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

const mirrorIndexBody = `{
  "name": "terraform",
  "versions": {
    "1.9.2": {"name":"terraform","version":"1.9.2","builds":[]},
    "1.9.3": {"name":"terraform","version":"1.9.3","builds":[]}
  }
}`

func TestResolveTerraformVersions_CustomDownloadBaseURL(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/terraform/index.json" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("content-type", "application/json")
		_, _ = w.Write([]byte(mirrorIndexBody))
	}))
	t.Cleanup(srv.Close)

	d := terraform.NewDistribution("terraform", srv.URL, terraform.APIAuth{})
	version, err := d.ResolveConstraint(context.Background(), "= 1.9.3")
	Ok(t, err)
	Equals(t, version.String(), "1.9.3")
}

func TestResolveTerraformVersions_MirrorBearerAuth(t *testing.T) {
	const wantToken = "my-mirror-token"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+wantToken {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		w.Header().Set("content-type", "application/json")
		_, _ = w.Write([]byte(mirrorIndexBody))
	}))
	t.Cleanup(srv.Close)

	d := terraform.NewDistribution("terraform", srv.URL, terraform.APIAuth{BearerToken: wantToken})
	version, err := d.ResolveConstraint(context.Background(), "= 1.9.3")
	Ok(t, err)
	Equals(t, version.String(), "1.9.3")
}

func TestResolveTerraformVersions_MirrorBasicAuth(t *testing.T) {
	const (
		wantUser = "user"
		wantPass = "pass"
	)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		if !ok || u != wantUser || p != wantPass {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		w.Header().Set("content-type", "application/json")
		_, _ = w.Write([]byte(mirrorIndexBody))
	}))
	t.Cleanup(srv.Close)

	d := terraform.NewDistribution("terraform", srv.URL, terraform.APIAuth{Username: wantUser, Password: wantPass})
	version, err := d.ResolveConstraint(context.Background(), "= 1.9.3")
	Ok(t, err)
	Equals(t, version.String(), "1.9.3")
}
