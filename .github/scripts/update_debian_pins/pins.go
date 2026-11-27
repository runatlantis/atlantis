// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"log"
	"regexp"
	"slices"

	"pault.ag/go/debian/version"
)

// pinPattern matches the same Renovate annotation the customManagers regex in
// .github/renovate.json5 reads, restricted to the deb datasource. Group 1 is
// the package name, group 2 the pinned version.
var pinPattern = regexp.MustCompile(`(?m)^#\s*renovate: datasource=deb depName=(\S+)[^\n]*\n(?:ARG|ENV) \w+_VERSION="?([^"\s]+)"?`)

// pin is one Debian package version pinned in the Dockerfile.
type pin struct {
	Package string
	Current string
	// start and end are the byte offsets of Current in the Dockerfile.
	start, end int
}

// update is a pin that should move to a newer version.
type update struct {
	Package string
	From    string
	To      string
}

func findPins(dockerfile []byte) []pin {
	var pins []pin
	for _, m := range pinPattern.FindAllSubmatchIndex(dockerfile, -1) {
		pins = append(pins, pin{
			Package: string(dockerfile[m[2]:m[3]]),
			Current: string(dockerfile[m[4]:m[5]]),
			start:   m[4],
			end:     m[5],
		})
	}
	return pins
}

// applyUpdates returns a copy of dockerfile with each pinned package set to
// the version in updates. Packages missing from updates keep their pin.
func applyUpdates(dockerfile []byte, updates map[string]string) []byte {
	out := make([]byte, 0, len(dockerfile))
	last := 0
	for _, p := range findPins(dockerfile) {
		to, ok := updates[p.Package]
		if !ok {
			continue
		}
		out = append(out, dockerfile[last:p.start]...)
		out = append(out, to...)
		last = p.end
	}
	return append(out, dockerfile[last:]...)
}

// planUpdates picks, for each pin, the highest version that every
// architecture can install, and returns the pins where that is newer than the
// current one. available maps architecture to package to the versions found
// across all suites, so this matches what apt resolves inside the image, where
// trixie and trixie-security are both configured.
func planUpdates(pins []pin, available map[string]map[string][]string) ([]update, error) {
	var updates []update
	for _, p := range pins {
		current, err := version.Parse(p.Current)
		if err != nil {
			return nil, fmt.Errorf("parsing pinned version %q of %q: %w", p.Current, p.Package, err)
		}
		best, found, err := highestCommonVersion(p.Package, available)
		if err != nil {
			return nil, err
		}
		if !found {
			log.Printf("no version of %q is available on every architecture, leaving it at %s", p.Package, p.Current)
			continue
		}
		if version.Compare(best, current) > 0 {
			updates = append(updates, update{Package: p.Package, From: p.Current, To: best.String()})
		}
	}
	return updates, nil
}

// highestCommonVersion returns the highest version of pkg present on every
// architecture in available. A binNMU (for example "+b3") can exist on one
// architecture only, and pinning it would break the image build for the rest.
func highestCommonVersion(pkg string, available map[string]map[string][]string) (version.Version, bool, error) {
	var common []string
	first := true
	for _, packages := range available {
		if first {
			common = slices.Clone(packages[pkg])
			first = false
			continue
		}
		common = slices.DeleteFunc(common, func(v string) bool {
			return !slices.Contains(packages[pkg], v)
		})
	}

	var best version.Version
	found := false
	for _, s := range common {
		v, err := version.Parse(s)
		if err != nil {
			return version.Version{}, false, fmt.Errorf("parsing index version %q of %q: %w", s, pkg, err)
		}
		if !found || version.Compare(v, best) > 0 {
			best, found = v, true
		}
	}
	return best, found, nil
}
