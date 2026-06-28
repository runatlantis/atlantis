// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package tfclient

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
)

// Schema matching terraform-config-inspect: terraform block with no labels.
var tofuRootSchema = &hcl.BodySchema{
	Blocks: []hcl.BlockHeaderSchema{
		{Type: "terraform", LabelNames: nil},
	},
}

var tofuTerraformBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{Name: "required_version"},
	},
}

// detectRequiredCoreFromTofu reads .tofu, .tofu.json, .tf, and .tf.json files
// in the given directory, applying OpenTofu file precedence rules:
//   - .tofu overrides same-basename .tf
//   - .tofu.json overrides same-basename .tf.json
//
// Returns constraints found and any diagnostics encountered. When no constraints
// are found and diagnostics exist, returns a non-nil error. When constraints are
// recovered despite diagnostics, returns both so the caller can log warnings.
func detectRequiredCoreFromTofu(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	tofuFiles := make(map[string]bool)
	tofuJSONFiles := make(map[string]bool)

	for _, e := range entries {
		if e.IsDir() || isIgnoredFile(e.Name()) {
			continue
		}
		name := e.Name()
		if base, ok := strings.CutSuffix(name, ".tofu.json"); ok {
			tofuJSONFiles[base] = true
		} else if base, ok := strings.CutSuffix(name, ".tofu"); ok {
			tofuFiles[base] = true
		}
	}

	var filesToParse []string
	for _, e := range entries {
		if e.IsDir() || isIgnoredFile(e.Name()) {
			continue
		}
		name := e.Name()
		switch {
		case strings.HasSuffix(name, ".tofu.json"):
			filesToParse = append(filesToParse, filepath.Join(dir, name))
		case strings.HasSuffix(name, ".tofu"):
			filesToParse = append(filesToParse, filepath.Join(dir, name))
		case strings.HasSuffix(name, ".tf.json"):
			base, _ := strings.CutSuffix(name, ".tf.json")
			if !tofuJSONFiles[base] {
				filesToParse = append(filesToParse, filepath.Join(dir, name))
			}
		case strings.HasSuffix(name, ".tf"):
			base, _ := strings.CutSuffix(name, ".tf")
			if !tofuFiles[base] {
				filesToParse = append(filesToParse, filepath.Join(dir, name))
			}
		}
	}

	var constraints []string
	var parseErrors []error
	parser := hclparse.NewParser()

	for _, path := range filesToParse {
		found, parseErr := extractRequiredVersion(parser, path)
		if parseErr != nil {
			parseErrors = append(parseErrors, parseErr)
		}
		constraints = append(constraints, found...)
	}

	if len(constraints) == 0 && len(parseErrors) > 0 {
		return nil, errors.Join(parseErrors...)
	}

	if len(parseErrors) > 0 {
		return constraints, errors.Join(parseErrors...)
	}

	return constraints, nil
}

// isIgnoredFile returns true if the filename should be skipped (editor swap
// files, hidden files, etc.). Matches terraform-config-inspect behavior.
func isIgnoredFile(name string) bool {
	return strings.HasPrefix(name, ".") ||
		strings.HasSuffix(name, "~") ||
		(strings.HasPrefix(name, "#") && strings.HasSuffix(name, "#"))
}

// extractRequiredVersion parses either native HCL or HCL JSON format files
// using the HCL parser (matching terraform-config-inspect behavior) and
// extracts required_version from zero-label terraform {} blocks.
func extractRequiredVersion(parser *hclparse.Parser, path string) ([]string, error) {
	var file *hcl.File
	var diags hcl.Diagnostics

	if strings.HasSuffix(path, ".json") {
		file, diags = parser.ParseJSONFile(path)
	} else {
		file, diags = parser.ParseHCLFile(path)
	}

	if file == nil {
		if diags.HasErrors() {
			return nil, fmt.Errorf("%s: %s", path, diags.Error())
		}
		return nil, nil
	}

	content, _, contentDiags := file.Body.PartialContent(tofuRootSchema)
	diags = append(diags, contentDiags...)

	if content == nil {
		if diags.HasErrors() {
			return nil, fmt.Errorf("%s: %s", path, diags.Error())
		}
		return nil, nil
	}

	var constraints []string
	var valueErrors []error

	for _, block := range content.Blocks {
		if block.Type != "terraform" {
			continue
		}

		innerContent, _, innerDiags := block.Body.PartialContent(tofuTerraformBlockSchema)
		diags = append(diags, innerDiags...)

		if innerContent == nil {
			continue
		}

		attr, defined := innerContent.Attributes["required_version"]
		if !defined {
			continue
		}

		val, valDiags := attr.Expr.Value(nil)
		if valDiags.HasErrors() {
			valueErrors = append(valueErrors, fmt.Errorf("%s: required_version: %s", path, valDiags.Error()))
			continue
		}
		if val.IsNull() || !val.IsKnown() {
			valueErrors = append(valueErrors, fmt.Errorf("%s: required_version must be a string literal", path))
			continue
		}

		strVal, convErr := convert.Convert(val, cty.String)
		if convErr != nil {
			valueErrors = append(valueErrors, fmt.Errorf("%s: required_version: %s", path, convErr))
			continue
		}
		constraints = append(constraints, strVal.AsString())
	}

	var allErrs []error
	allErrs = append(allErrs, valueErrors...)
	if diags.HasErrors() {
		allErrs = append(allErrs, fmt.Errorf("%s: %s", path, diags.Error()))
	}

	if len(constraints) == 0 && len(allErrs) > 0 {
		return nil, errors.Join(allErrs...)
	}
	if len(allErrs) > 0 {
		return constraints, errors.Join(allErrs...)
	}
	return constraints, nil
}
