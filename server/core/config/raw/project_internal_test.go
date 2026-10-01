// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package raw

import (
	"testing"

	"github.com/bmatcuk/doublestar/v4"
)

func TestTerraformProjectIndicators(t *testing.T) {
	for _, indicator := range terraformProjectIndicators {
		if !doublestar.ValidatePattern(indicator) {
			t.Errorf("%s is not a valid pattern", indicator)
		}
	}
}
