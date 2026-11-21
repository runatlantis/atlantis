// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package valid

import (
	"fmt"
	"regexp"
	"strings"
)

const APIAllowVarsKey = "api_allow_vars"

var variableNameRegex = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_-]*$`)

// ValidateVariableName accepts ASCII Terraform identifiers.
func ValidateVariableName(name string) error {
	if !variableNameRegex.MatchString(name) {
		return fmt.Errorf("%q is not a valid variable name", name)
	}
	return nil
}

// ValidateAPIAllowVarPattern checks one api_allow_vars entry: a variable
// name, optionally ending in "*" to match every name with that prefix, or a
// lone "*" to match any name.
func ValidateAPIAllowVarPattern(pattern string) error {
	prefix, wildcard := strings.CutSuffix(pattern, "*")
	if wildcard && prefix == "" {
		return nil
	}
	if strings.Contains(prefix, "*") {
		return fmt.Errorf("invalid %s pattern %q: '*' is only allowed at the end", APIAllowVarsKey, pattern)
	}
	if !variableNameRegex.MatchString(prefix) {
		return fmt.Errorf("invalid %s pattern %q: must be a variable name, optionally ending in '*'", APIAllowVarsKey, pattern)
	}
	return nil
}

// APIVarAllowed reports whether name matches any of patterns, which must have
// passed ValidateAPIAllowVarPattern.
func APIVarAllowed(patterns []string, name string) bool {
	for _, pattern := range patterns {
		if prefix, wildcard := strings.CutSuffix(pattern, "*"); wildcard {
			if strings.HasPrefix(name, prefix) {
				return true
			}
		} else if name == pattern {
			return true
		}
	}
	return false
}
