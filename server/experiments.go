// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"fmt"
	"slices"
	"strings"
)

// ExperimentsDocumentationURL includes experiments not yet in a stable release.
const ExperimentsDocumentationURL = "https://www.runatlantis.io/versions/main/docs/experiments.html"

const (
	apiExperiment   = "api"
	driftExperiment = "drift-detection"
)

var knownExperiments = []string{apiExperiment, driftExperiment}

// ValidateExperiments checks permissions and prerequisites before startup.
func (c UserConfig) ValidateExperiments() error {
	for _, name := range c.Experiments {
		if !slices.Contains(knownExperiments, name) {
			return fmt.Errorf("unknown server experiment %q in experiments configuration; supported names are %s; see %s", name, strings.Join(knownExperiments, ", "), ExperimentsDocumentationURL)
		}
	}
	api := slices.Contains(c.Experiments, apiExperiment)
	drift := slices.Contains(c.Experiments, driftExperiment)
	for _, setting := range []struct {
		name    string
		enabled bool
	}{
		{"enable-drift-detection", c.EnableDriftDetection},
		{"enable-drift-remediation", c.EnableDriftRemediation},
	} {
		if setting.enabled && !drift {
			return fmt.Errorf("%s requires the %q experiment; add it to experiments or disable %s; see %s", setting.name, driftExperiment, setting.name, ExperimentsDocumentationURL)
		}
	}
	// Both route groups use the secret; either experiment may authorize it.
	if c.APISecret != "" && !api && !drift {
		return fmt.Errorf("api-secret requires the %q experiment for command/lock endpoints or %q for drift endpoints; add the matching name to experiments or remove api-secret; see %s", apiExperiment, driftExperiment, ExperimentsDocumentationURL)
	}
	if c.EnableDriftRemediation && !c.EnableDriftDetection {
		return fmt.Errorf("--enable-drift-remediation requires --enable-drift-detection; see %s", ExperimentsDocumentationURL)
	}
	if c.EnableDriftDetection && c.APISecret == "" {
		return fmt.Errorf("enable-drift-detection requires api-secret for authentication; configure api-secret or disable enable-drift-detection; see %s", ExperimentsDocumentationURL)
	}
	return nil
}
