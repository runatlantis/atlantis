// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/ulikunitz/xz"
)

func TestFetchVersions(t *testing.T) {
	var compressed bytes.Buffer
	w, err := xz.NewWriter(&compressed)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte(testIndex)); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/debian-security/dists/trixie-security/main/binary-amd64/Packages.xz" {
			http.NotFound(rw, r)
			return
		}
		rw.Write(compressed.Bytes()) // nolint: errcheck
	}))
	defer server.Close()

	tests := []struct {
		name    string
		suite   string
		want    map[string][]string
		wantErr bool
	}{
		{
			name:  "decompresses and parses the index",
			suite: "debian-security/trixie-security",
			want:  map[string][]string{"curl": {"8.14.1-2+deb13u7", "8.14.1-2+deb13u5"}},
		},
		{
			name:    "fails on a missing index",
			suite:   "debian/trixie-backports",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := server.URL + "/" + suitePath(tt.suite) + "/main/binary-amd64/Packages.xz"
			got, err := fetchVersions(server.Client(), url, map[string]bool{"curl": true})
			if (err != nil) != tt.wantErr {
				t.Fatalf("fetchVersions() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("fetchVersions() = %v, want %v", got, tt.want)
			}
		})
	}
}
