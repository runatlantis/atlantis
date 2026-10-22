// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"reflect"
	"testing"
)

const testDockerfile = `FROM debian:13 AS debian-base
# renovate: datasource=deb depName=curl
ENV DEBIAN_CURL_VERSION="8.14.1-2+deb13u5"
# renovate: datasource=deb depName=openssh-server
ENV DEBIAN_OPENSSH_SERVER_VERSION=1:10.0p1-7+deb13u4
# renovate: datasource=apk depName=curl
ENV CURL_VERSION="8.22.0-r0"
# renovate: datasource=deb depName=libcap2-bin versioning=deb
ARG DEBIAN_LIBCAP2_BIN_VERSION="1:2.75-10+deb13u1+b3"
`

func TestFindPins(t *testing.T) {
	got := findPins([]byte(testDockerfile))

	want := []pin{
		{Package: "curl", Current: "8.14.1-2+deb13u5"},
		{Package: "openssh-server", Current: "1:10.0p1-7+deb13u4"},
		{Package: "libcap2-bin", Current: "1:2.75-10+deb13u1+b3"},
	}
	if len(got) != len(want) {
		t.Fatalf("findPins() returned %d pins, want %d: %+v", len(got), len(want), got)
	}
	for i := range want {
		if got[i].Package != want[i].Package || got[i].Current != want[i].Current {
			t.Errorf("pin %d = %s %q, want %s %q", i, got[i].Package, got[i].Current, want[i].Package, want[i].Current)
		}
	}
}

func TestApplyUpdates(t *testing.T) {
	tests := []struct {
		name    string
		updates map[string]string
		want    string
	}{
		{
			name:    "no updates leaves the file unchanged",
			updates: map[string]string{},
			want:    testDockerfile,
		},
		{
			name: "keeps quoting and leaves apk pins alone",
			updates: map[string]string{
				"curl":           "8.14.1-2+deb13u7",
				"openssh-server": "1:10.0p1-7+deb13u5",
			},
			want: `FROM debian:13 AS debian-base
# renovate: datasource=deb depName=curl
ENV DEBIAN_CURL_VERSION="8.14.1-2+deb13u7"
# renovate: datasource=deb depName=openssh-server
ENV DEBIAN_OPENSSH_SERVER_VERSION=1:10.0p1-7+deb13u5
# renovate: datasource=apk depName=curl
ENV CURL_VERSION="8.22.0-r0"
# renovate: datasource=deb depName=libcap2-bin versioning=deb
ARG DEBIAN_LIBCAP2_BIN_VERSION="1:2.75-10+deb13u1+b3"
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := string(applyUpdates([]byte(testDockerfile), tt.updates))
			if got != tt.want {
				t.Errorf("applyUpdates() =\n%s\nwant\n%s", got, tt.want)
			}
		})
	}
}

func TestPlanUpdates(t *testing.T) {
	pins := []pin{
		{Package: "curl", Current: "8.14.1-2+deb13u5"},
		{Package: "openssh-server", Current: "1:10.0p1-7+deb13u4"},
		{Package: "libcap2-bin", Current: "1:2.75-10+deb13u1+b3"},
		{Package: "gone", Current: "1.0-1"},
	}
	// Each architecture's set is already the union of every suite.
	available := map[string]map[string][]string{
		"amd64": {
			"curl":           {"8.14.1-2+deb13u5", "8.14.1-2+deb13u7"},
			"openssh-server": {"1:10.0p1-7+deb13u4", "1:10.0p1-7+deb13u2"},
			"libcap2-bin":    {"1:2.75-10+deb13u1+b3", "1:2.75-10+deb13u2"},
		},
		"armhf": {
			"curl":           {"8.14.1-2+deb13u5", "8.14.1-2+deb13u7"},
			"openssh-server": {"1:10.0p1-7+deb13u4"},
			// The newer build has not reached armhf yet.
			"libcap2-bin": {"1:2.75-10+deb13u1+b3"},
		},
	}

	got, err := planUpdates(pins, available)
	if err != nil {
		t.Fatalf("planUpdates() error = %v", err)
	}

	want := []update{
		{Package: "curl", From: "8.14.1-2+deb13u5", To: "8.14.1-2+deb13u7"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("planUpdates() = %+v, want %+v", got, want)
	}
}

func TestPlanUpdatesRejectsInvalidVersion(t *testing.T) {
	pins := []pin{{Package: "curl", Current: "8.14.1-2+deb13u5"}}
	available := map[string]map[string][]string{
		"amd64": {"curl": {"not a version!"}},
	}

	if _, err := planUpdates(pins, available); err == nil {
		t.Error("planUpdates() error = nil, want an error for an unparsable version")
	}
}
