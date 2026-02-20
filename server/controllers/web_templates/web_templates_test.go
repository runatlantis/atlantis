// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package web_templates

import (
	"bytes"
	"encoding/json"
	"html/template"
	"io"
	"strings"
	"testing"

	. "github.com/runatlantis/atlantis/testing"
)

func TestProjectJobsErrorTemplate(t *testing.T) {
	err := ProjectJobsErrorTemplate.Execute(io.Discard, ProjectJobsError{
		LayoutData: LayoutData{
			AtlantisVersion: "v0.0.0",
			CleanedBasePath: "/path",
		},
	})
	Ok(t, err)
}

func TestGithubAppSetupTemplate(t *testing.T) {
	err := GithubAppSetupTemplate.Execute(io.Discard, GithubSetupData{
		Target:          "target",
		Manifest:        "manifest",
		ID:              1,
		Key:             "key",
		WebhookSecret:   "webhook secret",
		URL:             "https://example.com",
		CleanedBasePath: "/path",
	})
	Ok(t, err)
}

func TestMustEncodeScriptData(t *testing.T) {
	t.Run("encodes strings with ANSI escape codes", func(t *testing.T) {
		data := map[string]string{
			"output": "\033[1mBold\033[0m \033[32mgreen\033[0m",
		}
		result := MustEncodeScriptData(data)

		// Parse the JSON to verify round-trip
		var decoded map[string]string
		err := json.Unmarshal([]byte(string(result)), &decoded)
		Ok(t, err)
		Equals(t, "\033[1mBold\033[0m \033[32mgreen\033[0m", decoded["output"])
	})

	t.Run("encodes mixed types", func(t *testing.T) {
		data := map[string]any{
			"output":    "hello",
			"startTime": int64(1234567890),
			"endTime":   int64(0),
		}
		result := MustEncodeScriptData(data)

		var decoded map[string]any
		err := json.Unmarshal([]byte(string(result)), &decoded)
		Ok(t, err)
		Equals(t, "hello", decoded["output"])
	})

	t.Run("encodes empty data", func(t *testing.T) {
		result := MustEncodeScriptData(map[string]string{})
		Equals(t, template.HTML("{}"), result)
	})
}

func TestMustEncodeScriptData_ANSIRoundTrip(t *testing.T) {
	// Simulate the full template rendering pipeline to verify ANSI codes
	// survive: Go string → MustEncodeScriptData → template rendering →
	// browser JSON.parse would recover the original string.
	//
	// This is the critical test: html/template's built-in "js" function
	// double-escapes \x1b inside <script> blocks, producing \\u001B which
	// browsers display as literal text. MustEncodeScriptData avoids this
	// by returning template.HTML which bypasses context-aware escaping.
	ansiOutput := "\033[1mTerraform\033[0m will perform the following actions:\n" +
		"  \033[32m+\033[0m resource \"aws_instance\" \"web\" {\n" +
		"      \033[32m+\033[0m ami = \"ami-12345\"\n" +
		"    }\n" +
		"\033[1mPlan:\033[0m 1 to add, 0 to change, 0 to destroy."

	scriptData := MustEncodeScriptData(map[string]string{
		"output": ansiOutput,
		"error":  "",
	})

	// Create a minimal template that mirrors the real template pattern:
	// a hidden <div> used as a data carrier (not <script>, which HTMX strips
	// when allowScriptTags=false).
	tmpl := template.Must(template.New("test").Parse(
		`<div id="output-data" hidden>{{ .ScriptData }}</div>`))

	var buf bytes.Buffer
	err := tmpl.Execute(&buf, struct{ ScriptData template.HTML }{ScriptData: scriptData})
	Ok(t, err)

	rendered := buf.String()

	// The rendered HTML should NOT contain double-escaped sequences like \\u001B
	if strings.Contains(rendered, `\\u001B`) || strings.Contains(rendered, `\\u001b`) {
		t.Errorf("rendered HTML contains double-escaped ANSI codes: %s", rendered)
	}

	// Extract the JSON from the hidden div and parse it
	jsonStr := strings.TrimPrefix(rendered, `<div id="output-data" hidden>`)
	jsonStr = strings.TrimSuffix(jsonStr, `</div>`)

	var decoded map[string]string
	err = json.Unmarshal([]byte(jsonStr), &decoded)
	Ok(t, err)

	// The decoded output should exactly match the original Go string
	Equals(t, ansiOutput, decoded["output"])
	Equals(t, "", decoded["error"])
}

func TestProjectOutputTemplate_ANSIInOutput(t *testing.T) {
	// Verify the actual ProjectOutputTemplate renders ANSI codes correctly
	ansiOutput := "\033[32m+\033[0m resource \"aws_instance\" \"web\" {}"

	data := ProjectOutputData{
		LayoutData: LayoutData{
			AtlantisVersion: "v0.0.0",
			CleanedBasePath: "/test",
			ActiveNav:       "prs",
		},
		RepoFullName: "owner/repo",
		RepoOwner:    "owner",
		RepoName:     "repo",
		PullNum:      1,
		PullURL:      "https://github.com/owner/repo/pull/1",
		Path:         ".",
		Workspace:    "default",
		Status:       "success",
		StatusLabel:  "Planned",
		CommandName:  "plan",
		Output:       ansiOutput,
		OutputScriptData: MustEncodeScriptData(map[string]string{
			"output": ansiOutput,
			"error":  "",
		}),
	}

	var buf bytes.Buffer
	err := GetTemplate(TemplateName_ProjectOutput).Execute(&buf, data)
	Ok(t, err)

	rendered := buf.String()

	// Must NOT contain double-escaped ANSI sequences
	if strings.Contains(rendered, `\\u001B`) || strings.Contains(rendered, `\\u001b`) {
		t.Error("ProjectOutputTemplate contains double-escaped ANSI codes")
	}

	// Must contain the data block with properly encoded ANSI
	if !strings.Contains(rendered, `\u001b[32m`) {
		t.Error("ProjectOutputTemplate missing properly encoded ANSI codes in data block")
	}
}

// extractDataDivJSON extracts JSON content from a rendered HTML string
// by finding the element with the given id and extracting text between > and </div>.
func extractDataDivJSON(rendered, id string) string {
	marker := `id="` + id + `"`
	start := strings.Index(rendered, marker)
	if start < 0 {
		return ""
	}
	contentStart := strings.Index(rendered[start:], ">") + start + 1
	contentEnd := strings.Index(rendered[contentStart:], "</div>") + contentStart
	return rendered[contentStart:contentEnd]
}

func TestProjectOutputPartialTemplate_ContainsTerminalAndData(t *testing.T) {
	// Verify the partial template (used by HTMX swaps) renders the terminal
	// container and output-data div when Output is set.
	ansiOutput := "\033[32m+\033[0m resource \"aws_instance\" \"web\" {}"

	data := ProjectOutputData{
		Status:      "success",
		StatusLabel: "Planned",
		CommandName: "plan",
		Workspace:   "default",
		Output:      ansiOutput,
		OutputScriptData: MustEncodeScriptData(map[string]string{
			"output": ansiOutput,
			"error":  "",
		}),
	}

	var buf bytes.Buffer
	err := GetTemplate(TemplateName_ProjectOutputPartial).Execute(&buf, data)
	Ok(t, err)

	rendered := buf.String()

	// Must contain the terminal container div
	if !strings.Contains(rendered, `id="output-terminal"`) {
		t.Error("partial template missing output-terminal div")
	}

	// Must contain the output-data element (hidden div, not script — HTMX strips scripts)
	if !strings.Contains(rendered, `id="output-data"`) {
		t.Error("partial template missing output-data element")
	}

	// Extract JSON from the output-data div and verify round-trip
	jsonStr := extractDataDivJSON(rendered, "output-data")
	if jsonStr != "" {
		var decoded map[string]string
		err := json.Unmarshal([]byte(jsonStr), &decoded)
		Ok(t, err)
		Equals(t, ansiOutput, decoded["output"])
		Equals(t, "", decoded["error"])
	}

	// Must NOT double-escape ANSI codes
	if strings.Contains(rendered, `\\u001B`) || strings.Contains(rendered, `\\u001b`) {
		t.Error("partial template contains double-escaped ANSI codes")
	}
}

func TestProjectOutputPartialTemplate_ContainsErrorData(t *testing.T) {
	// Verify the partial template includes error content when Error is set.
	errorMsg := "Error: resource not found"

	data := ProjectOutputData{
		Status:      "failed",
		StatusLabel: "Plan Failed",
		CommandName: "plan",
		Workspace:   "default",
		Output:      "some output before error",
		Error:       errorMsg,
		OutputScriptData: MustEncodeScriptData(map[string]string{
			"output": "some output before error",
			"error":  errorMsg,
		}),
	}

	var buf bytes.Buffer
	err := GetTemplate(TemplateName_ProjectOutputPartial).Execute(&buf, data)
	Ok(t, err)

	rendered := buf.String()

	// Must contain the error in the OOB error section
	if !strings.Contains(rendered, errorMsg) {
		t.Error("partial template missing error message")
	}

	// Error must also be in the output-data JSON for JS error styling toggle
	jsonStr := extractDataDivJSON(rendered, "output-data")
	if jsonStr != "" {
		var decoded map[string]string
		err := json.Unmarshal([]byte(jsonStr), &decoded)
		Ok(t, err)
		if decoded["error"] == "" {
			t.Error("output-data JSON missing error field")
		}
	}
}

func TestProjectOutputPartialTemplate_ContainsPolicyOutput(t *testing.T) {
	// Verify the partial template includes policy output for HTMX swaps.
	policyOutput := "PASS - policy/enforce_tags.rego - all resources tagged"

	data := ProjectOutputData{
		Status:         "success",
		StatusLabel:    "Planned",
		CommandName:    "plan",
		Workspace:      "default",
		Output:         "terraform plan output",
		PolicyOutput:   policyOutput,
		PolicyPassed:   true,
		HasPolicyCheck: true,
		OutputScriptData: MustEncodeScriptData(map[string]string{
			"output": "terraform plan output",
			"error":  "",
		}),
	}

	var buf bytes.Buffer
	err := GetTemplate(TemplateName_ProjectOutputPartial).Execute(&buf, data)
	Ok(t, err)

	rendered := buf.String()

	// Must contain the policy output text
	if !strings.Contains(rendered, policyOutput) {
		t.Error("partial template missing policy output - policy section not updating on HTMX swap")
	}

	// Policy section must be an OOB swap element
	if !strings.Contains(rendered, `id="output-policy"`) {
		t.Error("partial template missing output-policy OOB target")
	}
	if !strings.Contains(rendered, `hx-swap-oob="true"`) {
		t.Error("partial template missing hx-swap-oob on policy section")
	}
}

func TestProjectOutputPartialTemplate_NoPolicyOutput(t *testing.T) {
	// Verify the partial template sends an empty policy OOB div when no policy output,
	// so HTMX clears the policy section from a previous run.
	data := ProjectOutputData{
		Status:      "success",
		StatusLabel: "Planned",
		CommandName: "plan",
		Workspace:   "default",
		Output:      "terraform plan output",
		OutputScriptData: MustEncodeScriptData(map[string]string{
			"output": "terraform plan output",
			"error":  "",
		}),
	}

	var buf bytes.Buffer
	err := GetTemplate(TemplateName_ProjectOutputPartial).Execute(&buf, data)
	Ok(t, err)

	rendered := buf.String()

	// Must still have the output-policy OOB element (empty, to clear previous policy)
	if !strings.Contains(rendered, `id="output-policy"`) {
		t.Error("partial template missing output-policy OOB target when no policy output")
	}

	// Must NOT contain policy content
	if strings.Contains(rendered, "Policy Check Output") {
		t.Error("partial template should not show policy section when no policy output")
	}
}

func TestProjectOutputPartialTemplate_EmptyOutput(t *testing.T) {
	// Verify the partial template shows "No Output Available" when output is empty.
	data := ProjectOutputData{
		Status:      "success",
		StatusLabel: "Planned",
		CommandName: "plan",
		Workspace:   "default",
		Output:      "", // Empty output
	}

	var buf bytes.Buffer
	err := GetTemplate(TemplateName_ProjectOutputPartial).Execute(&buf, data)
	Ok(t, err)

	rendered := buf.String()

	// Should NOT contain terminal div (no output to show)
	if strings.Contains(rendered, `id="output-terminal"`) {
		t.Error("partial template should not render terminal when output is empty")
	}

	// Should show empty state
	if !strings.Contains(rendered, "No Output Available") {
		t.Error("partial template missing empty state message")
	}
}

func TestProjectOutputTemplate_HasOOBTargets(t *testing.T) {
	// Verify the main template has OOB swap target elements that the partial
	// template expects to exist for hx-swap-oob updates.
	data := ProjectOutputData{
		LayoutData: LayoutData{
			AtlantisVersion: "v0.0.0",
			CleanedBasePath: "/test",
			ActiveNav:       "prs",
		},
		RepoFullName: "owner/repo",
		RepoOwner:    "owner",
		RepoName:     "repo",
		PullNum:      1,
		PullURL:      "https://github.com/owner/repo/pull/1",
		Path:         ".",
		Workspace:    "default",
		Status:       "success",
		StatusLabel:  "Planned",
		CommandName:  "plan",
		Output:       "some output",
		OutputScriptData: MustEncodeScriptData(map[string]string{
			"output": "some output",
			"error":  "",
		}),
	}

	var buf bytes.Buffer
	err := GetTemplate(TemplateName_ProjectOutput).Execute(&buf, data)
	Ok(t, err)

	rendered := buf.String()

	// Must have OOB targets that the partial template swaps into
	if !strings.Contains(rendered, `id="output-stats"`) {
		t.Error("main template missing output-stats OOB target")
	}
	if !strings.Contains(rendered, `id="output-error"`) {
		t.Error("main template missing output-error OOB target")
	}
	if !strings.Contains(rendered, `id="output-policy"`) {
		t.Error("main template missing output-policy OOB target")
	}
	if !strings.Contains(rendered, `id="output-content"`) {
		t.Error("main template missing output-content swap target")
	}
}

func TestProjectOutputPartialTemplate_DataNotInScriptTag(t *testing.T) {
	// Verify the data carrier uses <div hidden> not <script>, because
	// HTMX's allowScriptTags=false strips all script tags from swapped content.
	data := ProjectOutputData{
		Status:      "success",
		StatusLabel: "Planned",
		CommandName: "plan",
		Workspace:   "default",
		Output:      "some output",
		OutputScriptData: MustEncodeScriptData(map[string]string{
			"output": "some output",
			"error":  "",
		}),
	}

	var buf bytes.Buffer
	err := GetTemplate(TemplateName_ProjectOutputPartial).Execute(&buf, data)
	Ok(t, err)

	rendered := buf.String()

	// output-data must NOT be a script tag (HTMX strips scripts on swap)
	if strings.Contains(rendered, `<script id="output-data"`) {
		t.Error("output-data must not be a <script> tag — HTMX allowScriptTags=false strips it")
	}

	// output-data must be a hidden div
	if !strings.Contains(rendered, `<div id="output-data" hidden>`) {
		t.Error("output-data should be a <div hidden> element")
	}
}

func TestJobDetailTemplate_ANSIInOutput(t *testing.T) {
	// Verify the actual JobDetailTemplate renders ANSI codes correctly
	ansiOutput := "\033[31mError:\033[0m something failed"

	data := JobDetailData{
		LayoutData: LayoutData{
			AtlantisVersion: "v0.0.0",
			CleanedBasePath: "/test",
			ActiveNav:       "jobs",
		},
		JobID:      "test-job-123",
		JobStep:    "plan",
		Status:     "complete",
		BadgeText:  "Planned",
		BadgeStyle: "success",
		Output:     ansiOutput,
		Workspace:  "default",
		TerminalScriptData: MustEncodeScriptData(map[string]any{
			"output":    ansiOutput,
			"badgeText": "Planned",
			"jobStep":   "plan",
			"status":    "complete",
			"startTime": int64(0),
			"endTime":   int64(0),
		}),
	}

	var buf bytes.Buffer
	err := GetTemplate(TemplateName_JobDetail).Execute(&buf, data)
	Ok(t, err)

	rendered := buf.String()

	// Must NOT contain double-escaped ANSI sequences
	if strings.Contains(rendered, `\\u001B`) || strings.Contains(rendered, `\\u001b`) {
		t.Error("JobDetailTemplate contains double-escaped ANSI codes")
	}

	// Must contain properly encoded ANSI in the data block
	if !strings.Contains(rendered, `\u001b[31m`) {
		t.Error("JobDetailTemplate missing properly encoded ANSI codes in data block")
	}
}
