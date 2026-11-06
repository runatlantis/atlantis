// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

// Command update_debian_pins bumps the Debian package pins in the Dockerfile to
// the newest version apt would install, including trixie-security.
//
// Renovate's deb datasource only reads Packages.gz, and the Debian security and
// updates suites only publish Packages.xz, so Renovate never sees security
// builds. When one lands for a library that a pinned package depends on at an
// exact version (curl needs libcurl4t64 (= same version)), the debian image
// build fails with unmet dependencies. Remove this once Renovate can read
// Packages.xz (https://github.com/renovatebot/renovate/pull/35865).
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ulikunitz/xz"
)

func main() {
	dockerfile := flag.String("dockerfile", "Dockerfile", "path to the Dockerfile to update")
	mirror := flag.String("mirror", "https://deb.debian.org", "Debian mirror base URL")
	suites := flag.String("suites", "debian/trixie,debian-security/trixie-security", "comma-separated archive/suite pairs to read")
	archs := flag.String("archs", "amd64,arm64,armhf", "comma-separated architectures the image is built for")
	summary := flag.String("summary", "", "optional path for a markdown summary of the updates")
	list := flag.Bool("list", false, "print the current pins as package=version, one per line, and exit")
	flag.Parse()

	if *list {
		if err := printPins(*dockerfile); err != nil {
			log.Fatal(err)
		}
		return
	}
	if err := run(*dockerfile, *mirror, strings.Split(*suites, ","), strings.Split(*archs, ","), *summary); err != nil {
		log.Fatal(err)
	}
}

func run(dockerfilePath, mirror string, suites, archs []string, summaryPath string) error {
	content, err := os.ReadFile(dockerfilePath)
	if err != nil {
		return fmt.Errorf("reading dockerfile: %w", err)
	}
	pins := findPins(content)
	if len(pins) == 0 {
		return fmt.Errorf("no deb pins found in %s", dockerfilePath)
	}

	wanted := make(map[string]bool, len(pins))
	for _, p := range pins {
		wanted[p.Package] = true
	}

	client := &http.Client{Timeout: 5 * time.Minute}
	available := make(map[string]map[string][]string, len(archs))
	for _, arch := range archs {
		available[arch] = make(map[string][]string)
		for _, suite := range suites {
			url := fmt.Sprintf("%s/%s/main/binary-%s/Packages.xz", mirror, suitePath(suite), arch)
			versions, err := fetchVersions(client, url, wanted)
			if err != nil {
				return err
			}
			for pkg, vs := range versions {
				available[arch][pkg] = append(available[arch][pkg], vs...)
			}
		}
	}

	updates, err := planUpdates(pins, available)
	if err != nil {
		return err
	}
	if len(updates) == 0 {
		log.Printf("all %d deb pins are current", len(pins))
		return nil
	}

	byPackage := make(map[string]string, len(updates))
	for _, u := range updates {
		log.Printf("bumping %s from %s to %s", u.Package, u.From, u.To)
		byPackage[u.Package] = u.To
	}
	if err := os.WriteFile(dockerfilePath, applyUpdates(content, byPackage), 0o644); err != nil {
		return fmt.Errorf("writing dockerfile: %w", err)
	}
	if summaryPath != "" {
		if err := os.WriteFile(summaryPath, []byte(renderSummary(updates)), 0o644); err != nil {
			return fmt.Errorf("writing summary: %w", err)
		}
	}
	return nil
}

// printPins prints each pin as an apt-get install argument, so the workflow
// can check that the pins resolve together inside the base image.
func printPins(dockerfilePath string) error {
	content, err := os.ReadFile(dockerfilePath)
	if err != nil {
		return fmt.Errorf("reading dockerfile: %w", err)
	}
	for _, p := range findPins(content) {
		fmt.Printf("%s=%s\n", p.Package, p.Current)
	}
	return nil
}

// suitePath turns "debian-security/trixie-security" into
// "debian-security/dists/trixie-security".
func suitePath(suite string) string {
	archive, name, _ := strings.Cut(suite, "/")
	return archive + "/dists/" + name
}

func fetchVersions(client *http.Client, url string, wanted map[string]bool) (map[string][]string, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("downloading %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("downloading %s: unexpected status %s", url, resp.Status)
	}
	r, err := xz.NewReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("decompressing %s: %w", url, err)
	}
	versions, err := readVersions(r, wanted)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", url, err)
	}
	return versions, nil
}

func renderSummary(updates []update) string {
	var b strings.Builder
	b.WriteString("| Package | From | To |\n|---|---|---|\n")
	for _, u := range updates {
		fmt.Fprintf(&b, "| `%s` | `%s` | `%s` |\n", u.Package, u.From, u.To)
	}
	return b.String()
}
