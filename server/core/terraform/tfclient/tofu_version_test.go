// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package tfclient_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/cmd"
	"github.com/runatlantis/atlantis/server/core/terraform"
	"github.com/runatlantis/atlantis/server/core/terraform/mocks"
	"github.com/runatlantis/atlantis/server/core/terraform/tfclient"
	jobmocks "github.com/runatlantis/atlantis/server/jobs/mocks"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func newTestClientForTofu(t *testing.T, downloadsAllowed bool) *tfclient.DefaultClient {
	t.Helper()
	RegisterMockTestingT(t)
	_, binDir, cacheDir := mkSubDirs(t)
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()
	mockDownloader := mocks.NewMockDownloader()
	distribution := terraform.NewDistributionTerraformWithDownloader(mockDownloader)

	c, err := tfclient.NewTestClient(
		logger(t), distribution, binDir, cacheDir,
		"", "", "1.0.0", cmd.DefaultTFVersionFlag, cmd.DefaultTFDownloadURL,
		downloadsAllowed, true, projectCmdOutputHandler,
	)
	Ok(t, err)
	return c
}

func logger(t *testing.T) logging.SimpleLogging {
	t.Helper()
	return logging.NewNoopLogger(t)
}

func opentofuDist(version string) *constraintResolvingDistribution {
	return &constraintResolvingDistribution{
		binName:         "tofu",
		resolvedVersion: version,
	}
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	Ok(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0600))
}

func makeProjectDir(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "project")
	Ok(t, os.MkdirAll(dir, 0700))
	return dir
}

// recordingLogger captures Err log messages for assertion in production-path tests.
type recordingLogger struct {
	logging.SimpleLogging
	errMessages []string
}

func newRecordingLogger(t *testing.T) *recordingLogger {
	return &recordingLogger{SimpleLogging: logging.NewNoopLogger(t)}
}

func (l *recordingLogger) Err(format string, a ...any) {
	l.errMessages = append(l.errMessages, fmt.Sprintf(format, a...))
	l.SimpleLogging.Err(format, a...)
}

func (l *recordingLogger) hasErrContaining(substr string) bool {
	for _, msg := range l.errMessages {
		if strings.Contains(msg, substr) {
			return true
		}
	}
	return false
}

// --- Theme 1: OpenTofu .tofu file detection ---

func TestDetectVersion_OpenTofu_TofuFileOnly(t *testing.T) {
	c := newTestClientForTofu(t, false)
	dir := makeProjectDir(t)
	writeFile(t, dir, "versions.tofu", `terraform { required_version = "1.12.1" }`)

	detected := c.DetectVersion(logger(t), opentofuDist("1.12.1"), dir)
	Assert(t, detected != nil, "expected version detected from .tofu file")
	Equals(t, "1.12.1", detected.String())
}

func TestDetectVersion_Terraform_IgnoresTofuFile(t *testing.T) {
	c := newTestClientForTofu(t, false)
	dir := makeProjectDir(t)
	writeFile(t, dir, "versions.tofu", `terraform { required_version = "1.12.1" }`)

	detected := c.DetectVersion(logger(t), nil, dir)
	Assert(t, detected == nil, "Terraform distribution must not read .tofu files")
}

func TestDetectVersion_Terraform_TfFileUnchanged(t *testing.T) {
	c := newTestClientForTofu(t, false)
	dir := makeProjectDir(t)
	writeFile(t, dir, "versions.tf", `terraform { required_version = "= 1.5.7" }`)

	detected := c.DetectVersion(logger(t), nil, dir)
	Assert(t, detected != nil, "existing .tf behavior should be unchanged")
	Equals(t, "1.5.7", detected.String())
}

func TestDetectVersion_OpenTofu_TfFileAlone(t *testing.T) {
	c := newTestClientForTofu(t, false)
	dir := makeProjectDir(t)
	writeFile(t, dir, "versions.tf", `terraform { required_version = "1.10.0" }`)

	detected := c.DetectVersion(logger(t), opentofuDist("1.10.0"), dir)
	Assert(t, detected != nil, "OpenTofu should still read .tf files")
	Equals(t, "1.10.0", detected.String())
}

// --- Theme 6: Same-basename precedence ---

func TestDetectVersion_OpenTofu_TofuOverridesSameBasenameTf(t *testing.T) {
	c := newTestClientForTofu(t, false)
	dir := makeProjectDir(t)
	writeFile(t, dir, "versions.tf", `terraform { required_version = "1.10.0" }`)
	writeFile(t, dir, "versions.tofu", `terraform { required_version = "1.12.1" }`)

	detected := c.DetectVersion(logger(t), opentofuDist("1.12.1"), dir)
	Assert(t, detected != nil, "expected .tofu to override same-basename .tf")
	Equals(t, "1.12.1", detected.String())
}

func TestDetectVersion_OpenTofu_TofuJSON_OverridesSameBasenameTfJSON(t *testing.T) {
	c := newTestClientForTofu(t, false)
	dir := makeProjectDir(t)
	writeFile(t, dir, "versions.tf.json", `{"terraform":{"required_version":"1.10.0"}}`)
	writeFile(t, dir, "versions.tofu.json", `{"terraform":{"required_version":"1.12.1"}}`)

	detected := c.DetectVersion(logger(t), opentofuDist("1.12.1"), dir)
	Assert(t, detected != nil, "expected .tofu.json to override same-basename .tf.json")
	Equals(t, "1.12.1", detected.String())
}

func TestDetectVersion_OpenTofu_UnrelatedTfAndTofu_BothContribute(t *testing.T) {
	c := newTestClientForTofu(t, false)
	dir := makeProjectDir(t)
	// Different basenames: both should contribute, giving 2 constraints → nil
	writeFile(t, dir, "main.tofu", `terraform { required_version = "1.12.1" }`)
	writeFile(t, dir, "other.tf", `terraform { required_version = "1.11.5" }`)

	detected := c.DetectVersion(logger(t), opentofuDist("1.12.1"), dir)
	Assert(t, detected == nil, "multiple constraints from different basenames should return nil")
}

func TestDetectVersion_OpenTofu_SameBasenameSuppressesOnlyThatBasename(t *testing.T) {
	c := newTestClientForTofu(t, false)
	dir := makeProjectDir(t)
	// versions.tofu overrides versions.tf, but other.tf is unrelated and contributes
	writeFile(t, dir, "versions.tf", `terraform { required_version = "1.10.0" }`)
	writeFile(t, dir, "versions.tofu", `terraform { required_version = "1.12.1" }`)
	writeFile(t, dir, "other.tf", `terraform { required_version = "1.11.5" }`)

	// Result: "1.12.1" from versions.tofu + "1.11.5" from other.tf → 2 constraints → nil
	detected := c.DetectVersion(logger(t), opentofuDist("1.12.1"), dir)
	Assert(t, detected == nil, "same-basename suppression should not suppress unrelated .tf files")
}

// --- Theme 1 continued: .tofu.json ---

func TestDetectVersion_OpenTofu_TofuJSON(t *testing.T) {
	c := newTestClientForTofu(t, false)
	dir := makeProjectDir(t)
	writeFile(t, dir, "versions.tofu.json", `{"terraform":{"required_version":"1.12.1"}}`)

	detected := c.DetectVersion(logger(t), opentofuDist("1.12.1"), dir)
	Assert(t, detected != nil, "expected version from .tofu.json")
	Equals(t, "1.12.1", detected.String())
}

// --- Theme 1: Constraint resolution ---

func TestDetectVersion_OpenTofu_ConstraintResolution(t *testing.T) {
	c := newTestClientForTofu(t, true)
	dir := makeProjectDir(t)
	writeFile(t, dir, "versions.tofu", `terraform { required_version = ">= 1.8" }`)

	dist := opentofuDist("1.12.3")
	detected := c.DetectVersion(logger(t), dist, dir)
	Assert(t, detected != nil, "expected constraint resolution")
	Equals(t, "1.12.3", detected.String())
	Equals(t, []string{">= 1.8"}, dist.constraints)
}

func TestDetectVersion_OpenTofu_MultipleConstraints_ReturnsNil(t *testing.T) {
	c := newTestClientForTofu(t, false)
	dir := makeProjectDir(t)
	writeFile(t, dir, "a.tofu", `terraform { required_version = "1.12.0" }`)
	writeFile(t, dir, "b.tofu", `terraform { required_version = "1.12.1" }`)

	detected := c.DetectVersion(logger(t), opentofuDist("1.12.1"), dir)
	Assert(t, detected == nil, "multiple constraints should return nil")
}

// --- Theme 7 + Theme 1 (P1): Crash safety / Invalid values ---

func TestDetectVersion_OpenTofu_NullRequiredVersion(t *testing.T) {
	c := newTestClientForTofu(t, false)
	dir := makeProjectDir(t)
	writeFile(t, dir, "versions.tofu", `terraform { required_version = null }`)

	detected := c.DetectVersion(logger(t), opentofuDist("1.12.1"), dir)
	Assert(t, detected == nil, "null required_version should not panic, returns nil")
}

func TestDetectVersion_OpenTofu_NumericRequiredVersion(t *testing.T) {
	// Numeric values are implicitly converted to string for compatibility
	// with gohcl.DecodeExpression behavior in terraform-config-inspect.
	c := newTestClientForTofu(t, false)
	dir := makeProjectDir(t)
	writeFile(t, dir, "versions.tofu", `terraform { required_version = 1.12 }`)

	detected := c.DetectVersion(logger(t), opentofuDist("1.12"), dir)
	Assert(t, detected != nil, "numeric required_version should be accepted via implicit conversion")
}

func TestDetectVersion_OpenTofu_BoolRequiredVersion(t *testing.T) {
	c := newTestClientForTofu(t, false)
	dir := makeProjectDir(t)
	writeFile(t, dir, "versions.tofu", `terraform { required_version = true }`)

	detected := c.DetectVersion(logger(t), opentofuDist("1.12.1"), dir)
	Assert(t, detected == nil, "bool required_version should not panic")
}

func TestDetectVersion_OpenTofu_ListRequiredVersion(t *testing.T) {
	c := newTestClientForTofu(t, false)
	dir := makeProjectDir(t)
	writeFile(t, dir, "versions.tofu", `terraform { required_version = ["1.12.0"] }`)

	detected := c.DetectVersion(logger(t), opentofuDist("1.12.1"), dir)
	Assert(t, detected == nil, "list required_version should not panic")
}

func TestDetectVersion_OpenTofu_VariableReference(t *testing.T) {
	c := newTestClientForTofu(t, false)
	dir := makeProjectDir(t)
	writeFile(t, dir, "versions.tofu", `
variable "version" { default = "1.12.0" }
terraform { required_version = var.version }
`)

	detected := c.DetectVersion(logger(t), opentofuDist("1.12.1"), dir)
	Assert(t, detected == nil, "variable reference should not panic, returns nil")
}

func TestDetectVersion_OpenTofu_ConditionalNull(t *testing.T) {
	c := newTestClientForTofu(t, false)
	dir := makeProjectDir(t)
	writeFile(t, dir, "versions.tofu", `terraform { required_version = true ? null : "1.12.0" }`)

	detected := c.DetectVersion(logger(t), opentofuDist("1.12.1"), dir)
	Assert(t, detected == nil, "conditional with null should not panic")
}

// --- Theme 7: Malformed files ---

func TestDetectVersion_OpenTofu_MalformedTofuFile(t *testing.T) {
	c := newTestClientForTofu(t, false)
	dir := makeProjectDir(t)
	writeFile(t, dir, "versions.tofu", `this is not valid HCL {{{`)

	detected := c.DetectVersion(logger(t), opentofuDist("1.12.1"), dir)
	Assert(t, detected == nil, "malformed .tofu should not panic")
}

func TestDetectVersion_OpenTofu_MalformedTofuJSON(t *testing.T) {
	c := newTestClientForTofu(t, false)
	dir := makeProjectDir(t)
	writeFile(t, dir, "versions.tofu.json", `{not valid json`)

	detected := c.DetectVersion(logger(t), opentofuDist("1.12.1"), dir)
	Assert(t, detected == nil, "malformed .tofu.json should not panic")
}

func TestDetectVersion_OpenTofu_MalformedPlusValid(t *testing.T) {
	c := newTestClientForTofu(t, false)
	dir := makeProjectDir(t)
	writeFile(t, dir, "bad.tofu", `this is not valid HCL {{{`)
	writeFile(t, dir, "good.tofu", `terraform { required_version = "1.12.1" }`)

	detected := c.DetectVersion(logger(t), opentofuDist("1.12.1"), dir)
	Assert(t, detected != nil, "valid file should still be detected despite malformed sibling")
	Equals(t, "1.12.1", detected.String())
}

func TestDetectVersion_OpenTofu_AllMalformed_LogsError(t *testing.T) {
	c := newTestClientForTofu(t, false)
	dir := makeProjectDir(t)
	writeFile(t, dir, "versions.tofu", `this is not valid HCL {{{`)

	detected := c.DetectVersion(logger(t), opentofuDist("1.12.1"), dir)
	Assert(t, detected == nil, "all-malformed should return nil and log error")
}

func TestDetectVersion_OpenTofu_MissingDirectory(t *testing.T) {
	c := newTestClientForTofu(t, false)
	detected := c.DetectVersion(logger(t), opentofuDist("1.12.1"), "/nonexistent/path/xyz")
	Assert(t, detected == nil, "missing directory should not panic")
}

// --- Theme 4: Ignored files ---

func TestDetectVersion_OpenTofu_IgnoresHiddenFiles(t *testing.T) {
	c := newTestClientForTofu(t, false)
	dir := makeProjectDir(t)
	writeFile(t, dir, ".versions.tofu", `terraform { required_version = "1.12.1" }`)

	detected := c.DetectVersion(logger(t), opentofuDist("1.12.1"), dir)
	Assert(t, detected == nil, "hidden .tofu file should be ignored")
}

func TestDetectVersion_OpenTofu_IgnoresBackupFiles(t *testing.T) {
	c := newTestClientForTofu(t, false)
	dir := makeProjectDir(t)
	writeFile(t, dir, "versions.tofu~", `terraform { required_version = "1.12.1" }`)

	detected := c.DetectVersion(logger(t), opentofuDist("1.12.1"), dir)
	Assert(t, detected == nil, "backup file (~) should be ignored")
}

func TestDetectVersion_OpenTofu_IgnoresEmacsFiles(t *testing.T) {
	c := newTestClientForTofu(t, false)
	dir := makeProjectDir(t)
	writeFile(t, dir, "#versions.tofu#", `terraform { required_version = "1.12.1" }`)

	detected := c.DetectVersion(logger(t), opentofuDist("1.12.1"), dir)
	Assert(t, detected == nil, "emacs temp file should be ignored")
}

func TestDetectVersion_OpenTofu_IgnoredFileDoesNotConflict(t *testing.T) {
	c := newTestClientForTofu(t, false)
	dir := makeProjectDir(t)
	// Ignored file has conflicting version; real file should win
	writeFile(t, dir, ".versions.tofu", `terraform { required_version = "9.9.9" }`)
	writeFile(t, dir, "versions.tofu", `terraform { required_version = "1.12.1" }`)

	detected := c.DetectVersion(logger(t), opentofuDist("1.12.1"), dir)
	Assert(t, detected != nil, "real file should be detected despite ignored conflict")
	Equals(t, "1.12.1", detected.String())
}

func TestDetectVersion_OpenTofu_IgnoresHiddenTfFile(t *testing.T) {
	c := newTestClientForTofu(t, false)
	dir := makeProjectDir(t)
	writeFile(t, dir, ".versions.tf", `terraform { required_version = "1.10.0" }`)

	detected := c.DetectVersion(logger(t), opentofuDist("1.12.1"), dir)
	Assert(t, detected == nil, "hidden .tf file should be ignored in OpenTofu mode")
}

// --- Finding 1: Labeled terraform blocks ---

func TestDetectVersion_OpenTofu_LabeledTerraformBlock(t *testing.T) {
	c := newTestClientForTofu(t, false)
	dir := makeProjectDir(t)
	writeFile(t, dir, "versions.tofu", `terraform "invalid" { required_version = "1.12.0" }`)

	detected := c.DetectVersion(logger(t), opentofuDist("1.12.0"), dir)
	Assert(t, detected == nil, "labeled terraform block should be rejected")
}

func TestDetectVersion_OpenTofu_LabeledBlockIgnored_ValidBlockUsed(t *testing.T) {
	c := newTestClientForTofu(t, false)
	dir := makeProjectDir(t)
	writeFile(t, dir, "versions.tofu", `
terraform "invalid" { required_version = "9.9.9" }
terraform { required_version = "1.12.1" }
`)

	detected := c.DetectVersion(logger(t), opentofuDist("1.12.1"), dir)
	Assert(t, detected != nil, "valid unlabeled block should be used despite labeled sibling")
	Equals(t, "1.12.1", detected.String())
}

// --- Finding 4: Partial file recovery ---

func TestDetectVersion_OpenTofu_PartialFileRecovery(t *testing.T) {
	c := newTestClientForTofu(t, false)
	dir := makeProjectDir(t)
	writeFile(t, dir, "versions.tofu", `
terraform {
  required_version = "1.12.1"
}

this is invalid HCL later in the file
`)

	detected := c.DetectVersion(logger(t), opentofuDist("1.12.1"), dir)
	Assert(t, detected != nil, "should recover required_version from partial file")
	Equals(t, "1.12.1", detected.String())
}

// --- Finding 9: JSON precedence pinned to same basename ---

func TestDetectVersion_OpenTofu_JSONPrecedence_OnlySameBasename(t *testing.T) {
	c := newTestClientForTofu(t, false)
	dir := makeProjectDir(t)
	writeFile(t, dir, "versions.tf.json", `{"terraform":{"required_version":"1.11.5"}}`)
	writeFile(t, dir, "versions.tofu.json", `{"terraform":{"required_version":"1.12.1"}}`)
	writeFile(t, dir, "other.tf.json", `{"terraform":{"required_version":"1.10.0"}}`)

	// versions.tofu.json overrides versions.tf.json, but other.tf.json contributes
	// Result: "1.12.1" + "1.10.0" = 2 constraints → nil
	detected := c.DetectVersion(logger(t), opentofuDist("1.12.1"), dir)
	Assert(t, detected == nil, "json precedence should only suppress same basename, not unrelated")
}

// --- Finding 3: Invalid JSON shapes ---

func TestDetectVersion_OpenTofu_InvalidJSON_TerraformNotObject(t *testing.T) {
	c := newTestClientForTofu(t, false)
	dir := makeProjectDir(t)
	writeFile(t, dir, "versions.tofu.json", `{"terraform": true}`)

	detected := c.DetectVersion(logger(t), opentofuDist("1.12.1"), dir)
	Assert(t, detected == nil, "terraform=true should be invalid shape")
}

func TestDetectVersion_OpenTofu_JSON_NumericVersion(t *testing.T) {
	// HCL JSON semantics support implicit number→string conversion
	// (same as gohcl.DecodeExpression in terraform-config-inspect).
	c := newTestClientForTofu(t, false)
	dir := makeProjectDir(t)
	writeFile(t, dir, "versions.tofu.json", `{"terraform":{"required_version": 123}}`)

	detected := c.DetectVersion(logger(t), opentofuDist("123"), dir)
	Assert(t, detected != nil, "numeric JSON required_version should be accepted via HCL JSON conversion")
}

func TestDetectVersion_OpenTofu_InvalidJSON_VersionNull(t *testing.T) {
	c := newTestClientForTofu(t, false)
	dir := makeProjectDir(t)
	writeFile(t, dir, "versions.tofu.json", `{"terraform":{"required_version": null}}`)

	detected := c.DetectVersion(logger(t), opentofuDist("1.12.1"), dir)
	Assert(t, detected == nil, "null JSON required_version should be rejected")
}

func TestDetectVersion_OpenTofu_InvalidJSON_VersionList(t *testing.T) {
	c := newTestClientForTofu(t, false)
	dir := makeProjectDir(t)
	writeFile(t, dir, "versions.tofu.json", `{"terraform":{"required_version": ["1.12.0"]}}`)

	detected := c.DetectVersion(logger(t), opentofuDist("1.12.1"), dir)
	Assert(t, detected == nil, "list JSON required_version should be rejected")
}

// --- Finding 10: Lower-level helper diagnostics ---

func TestDetectRequiredCoreFromTofu_MalformedOnly_ReturnsError(t *testing.T) {
	dir := makeProjectDir(t)
	writeFile(t, dir, "versions.tofu", `this is not valid HCL {{{`)

	constraints, err := tfclient.DetectRequiredCoreFromTofuForTest(dir)
	Assert(t, err != nil, "malformed-only should return error")
	Assert(t, len(constraints) == 0, "no constraints from malformed file")
	Assert(t, strings.Contains(err.Error(), "versions.tofu"), "error should reference the file")
}

func TestDetectRequiredCoreFromTofu_MalformedJSON_ReturnsError(t *testing.T) {
	dir := makeProjectDir(t)
	writeFile(t, dir, "versions.tofu.json", `{not valid json`)

	constraints, err := tfclient.DetectRequiredCoreFromTofuForTest(dir)
	Assert(t, err != nil, "malformed JSON should return error")
	Assert(t, len(constraints) == 0, "no constraints from malformed JSON")
}

func TestDetectRequiredCoreFromTofu_InvalidValue_ReturnsError(t *testing.T) {
	dir := makeProjectDir(t)
	writeFile(t, dir, "versions.tofu", `terraform { required_version = null }`)

	constraints, err := tfclient.DetectRequiredCoreFromTofuForTest(dir)
	Assert(t, err != nil, "invalid value with no valid constraints should return error")
	Assert(t, len(constraints) == 0, "no constraints from null value")
}

func TestDetectRequiredCoreFromTofu_ValidPlusMalformed_RecoversValid(t *testing.T) {
	dir := makeProjectDir(t)
	writeFile(t, dir, "bad.tofu", `this is not valid HCL {{{`)
	writeFile(t, dir, "good.tofu", `terraform { required_version = "1.12.1" }`)

	constraints, err := tfclient.DetectRequiredCoreFromTofuForTest(dir)
	Equals(t, []string{"1.12.1"}, constraints)
	// Diagnostics are still surfaced even when constraints are recovered
	Assert(t, err != nil, "should surface diagnostics from malformed sibling")
	Assert(t, strings.Contains(err.Error(), "bad.tofu"), "error should reference the malformed file")
}

// --- Finding 1 (JSON): HCL JSON array block form ---

func TestDetectVersion_OpenTofu_JSON_ArrayBlockForm(t *testing.T) {
	c := newTestClientForTofu(t, false)
	dir := makeProjectDir(t)
	writeFile(t, dir, "versions.tofu.json", `{"terraform":[{"required_version":"1.12.1"}]}`)

	detected := c.DetectVersion(logger(t), opentofuDist("1.12.1"), dir)
	Assert(t, detected != nil, "HCL JSON array block form should be accepted")
	Equals(t, "1.12.1", detected.String())
}

func TestDetectVersion_OpenTofu_TfJSON_ArrayBlockForm(t *testing.T) {
	c := newTestClientForTofu(t, false)
	dir := makeProjectDir(t)
	writeFile(t, dir, "versions.tf.json", `{"terraform":[{"required_version":"1.12.1"}]}`)

	detected := c.DetectVersion(logger(t), opentofuDist("1.12.1"), dir)
	Assert(t, detected != nil, "OpenTofu-mode .tf.json array block form should be accepted")
	Equals(t, "1.12.1", detected.String())
}

// --- Finding 4: Cross-format precedence ---

func TestDetectVersion_OpenTofu_CrossFormat_TofuDoesNotSuppressTfJSON(t *testing.T) {
	c := newTestClientForTofu(t, false)
	dir := makeProjectDir(t)
	// .tofu should not suppress .tf.json (different format counterpart)
	writeFile(t, dir, "versions.tofu", `terraform { required_version = "1.12.1" }`)
	writeFile(t, dir, "versions.tf.json", `{"terraform":{"required_version":"1.12.4"}}`)

	// Both contribute → multiple constraints → nil
	detected := c.DetectVersion(logger(t), opentofuDist("1.12.1"), dir)
	Assert(t, detected == nil, ".tofu must not suppress .tf.json; multiple constraints → nil")
}

func TestDetectVersion_OpenTofu_CrossFormat_TofuJSONDoesNotSuppressTf(t *testing.T) {
	c := newTestClientForTofu(t, false)
	dir := makeProjectDir(t)
	// .tofu.json should not suppress .tf (different format counterpart)
	writeFile(t, dir, "versions.tofu.json", `{"terraform":{"required_version":"1.12.1"}}`)
	writeFile(t, dir, "versions.tf", `terraform { required_version = "1.12.4" }`)

	// Both contribute → multiple constraints → nil
	detected := c.DetectVersion(logger(t), opentofuDist("1.12.1"), dir)
	Assert(t, detected == nil, ".tofu.json must not suppress .tf; multiple constraints → nil")
}

// --- Finding 7: Assert exact numeric constraint value ---

func TestDetectRequiredCoreFromTofu_NumericHCL_ExactValue(t *testing.T) {
	dir := makeProjectDir(t)
	writeFile(t, dir, "versions.tofu", `terraform { required_version = 1.12 }`)

	constraints, err := tfclient.DetectRequiredCoreFromTofuForTest(dir)
	Assert(t, err == nil, "no error for numeric value")
	Assert(t, len(constraints) == 1, "expected one constraint")
	Equals(t, "1.12", constraints[0])
}

func TestDetectRequiredCoreFromTofu_NumericJSON_ExactValue(t *testing.T) {
	dir := makeProjectDir(t)
	writeFile(t, dir, "versions.tofu.json", `{"terraform":{"required_version": 1.12}}`)

	constraints, err := tfclient.DetectRequiredCoreFromTofuForTest(dir)
	Assert(t, err == nil, "no error for JSON numeric value")
	Assert(t, len(constraints) == 1, "expected one constraint")
	Equals(t, "1.12", constraints[0])
}

// --- Finding 6: Terraform-mode .tofu.json exclusion ---

func TestDetectVersion_Terraform_IgnoresTofuJSONFile(t *testing.T) {
	c := newTestClientForTofu(t, false)
	dir := makeProjectDir(t)
	writeFile(t, dir, "versions.tofu.json", `{"terraform":{"required_version":"1.12.1"}}`)

	detected := c.DetectVersion(logger(t), nil, dir)
	Assert(t, detected == nil, "Terraform distribution must not read .tofu.json files")
}

func TestDetectVersion_Terraform_TfPlusTofuJSON_IgnoresTofuJSON(t *testing.T) {
	c := newTestClientForTofu(t, false)
	dir := makeProjectDir(t)
	writeFile(t, dir, "versions.tf", `terraform { required_version = "= 1.5.7" }`)
	writeFile(t, dir, "versions.tofu.json", `{"terraform":{"required_version":"1.12.1"}}`)

	detected := c.DetectVersion(logger(t), nil, dir)
	Assert(t, detected != nil, "Terraform should detect only from .tf")
	Equals(t, "1.5.7", detected.String())
}

// --- Finding 1: Same-file diagnostics ---

func TestDetectRequiredCoreFromTofu_PartialFile_ReturnsDiagnostics(t *testing.T) {
	dir := makeProjectDir(t)
	writeFile(t, dir, "versions.tofu", `
terraform {
  required_version = "1.12.1"
}

this is invalid HCL after the valid block
`)

	constraints, err := tfclient.DetectRequiredCoreFromTofuForTest(dir)
	Equals(t, []string{"1.12.1"}, constraints)
	Assert(t, err != nil, "should surface diagnostics even when constraint recovered")
	Assert(t, strings.Contains(err.Error(), "versions.tofu"), "diagnostic should reference file")
}

// --- Finding 3: Production-path DetectVersion diagnostics ---

func TestDetectVersion_OpenTofu_MalformedOnly_LogsDiagnostics(t *testing.T) {
	c := newTestClientForTofu(t, false)
	rl := newRecordingLogger(t)
	dir := makeProjectDir(t)
	writeFile(t, dir, "versions.tofu", `this is not valid HCL {{{`)

	detected := c.DetectVersion(rl, opentofuDist("1.12.1"), dir)
	Assert(t, detected == nil, "malformed-only should return nil")
	Assert(t, rl.hasErrContaining("versions.tofu"), "should log error referencing malformed file")
}

func TestDetectVersion_OpenTofu_MalformedJSON_LogsDiagnostics(t *testing.T) {
	c := newTestClientForTofu(t, false)
	rl := newRecordingLogger(t)
	dir := makeProjectDir(t)
	writeFile(t, dir, "versions.tofu.json", `{not valid json}`)

	detected := c.DetectVersion(rl, opentofuDist("1.12.1"), dir)
	Assert(t, detected == nil, "malformed JSON should return nil")
	Assert(t, rl.hasErrContaining("versions.tofu.json"), "should log error referencing malformed JSON")
}

func TestDetectVersion_OpenTofu_ValidPlusMalformed_LogsDiagnostics(t *testing.T) {
	c := newTestClientForTofu(t, false)
	rl := newRecordingLogger(t)
	dir := makeProjectDir(t)
	writeFile(t, dir, "bad.tofu", `this is not valid HCL {{{`)
	writeFile(t, dir, "good.tofu", `terraform { required_version = "1.12.1" }`)

	detected := c.DetectVersion(rl, opentofuDist("1.12.1"), dir)
	Assert(t, detected != nil, "should recover valid constraint")
	Equals(t, "1.12.1", detected.String())
	Assert(t, rl.hasErrContaining("bad.tofu"), "should log diagnostics for malformed sibling")
}

func TestDetectVersion_OpenTofu_PartialFile_LogsDiagnostics(t *testing.T) {
	c := newTestClientForTofu(t, false)
	rl := newRecordingLogger(t)
	dir := makeProjectDir(t)
	writeFile(t, dir, "versions.tofu", `
terraform {
  required_version = "1.12.1"
}

this is invalid HCL after
`)

	detected := c.DetectVersion(rl, opentofuDist("1.12.1"), dir)
	Assert(t, detected != nil, "should recover constraint from partial file")
	Equals(t, "1.12.1", detected.String())
	Assert(t, rl.hasErrContaining("versions.tofu"), "should log diagnostics from partial parse")
}

// --- Terraform .tf.json regression ---

func TestDetectVersion_Terraform_TfJSON(t *testing.T) {
	c := newTestClientForTofu(t, false)
	dir := makeProjectDir(t)
	writeFile(t, dir, "versions.tf.json", `{"terraform":{"required_version":"1.2.3"}}`)

	detected := c.DetectVersion(logger(t), nil, dir)
	Assert(t, detected != nil, "Terraform distribution must still detect from .tf.json")
	Equals(t, "1.2.3", detected.String())
}
