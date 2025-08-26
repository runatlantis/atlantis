package events_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
)

func TestTemplateOverridesLoading(t *testing.T) {
	tmpDir := t.TempDir()
	
	tests := []struct {
		name         string
		setupFiles   map[string]string
		expectedLogs []string
		expectError  bool
	}{
		{
			name: "successful override",
			setupFiles: map[string]string{
				"singleProjectPlanSuccess.tmpl": `{{ define "singleProjectPlanSuccess" }}TEST CUSTOM TEMPLATE{{ end }}`,
			},
			expectedLogs: []string{
				"found 1 template files matching pattern",
				"successfully parsed template overrides",
				"template overrides applied successfully",
			},
		},
		{
			name: "syntax error in template",
			setupFiles: map[string]string{
				"bad_template.tmpl": `{{ define "bad" }}{{ .MissingField {{ end }}`,
			},
			expectedLogs: []string{
				"found 1 template files matching pattern",
				"parsing template overrides",
				"continuing with built-in templates only",
			},
		},
		{
			name: "empty directory",
			setupFiles: map[string]string{},
			expectedLogs: []string{
				"no template files found in override directory",
			},
		},
		{
			name: "multiple templates",
			setupFiles: map[string]string{
				"singleProjectPlanSuccess.tmpl": `{{ define "singleProjectPlanSuccess" }}CUSTOM PLAN{{ end }}`,
				"planSuccessUnwrapped.tmpl":     `{{ define "planSuccessUnwrapped" }}CUSTOM UNWRAPPED{{ end }}`,
			},
			expectedLogs: []string{
				"found 2 template files matching pattern",
				"successfully parsed template overrides",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create template files
			templateDir := filepath.Join(tmpDir, tt.name)
			err := os.MkdirAll(templateDir, 0755)
			if err != nil {
				t.Fatal(err)
			}
			
			for filename, content := range tt.setupFiles {
				err := os.WriteFile(filepath.Join(templateDir, filename), []byte(content), 0644)
				if err != nil {
					t.Fatal(err)
				}
			}
			
			// Create fresh logger for each test
			logger := logging.NewNoopLogger(t).WithHistory()
			
			// Create renderer with overrides
			renderer := events.NewMarkdownRenderer(
				false, false, false, false, false, false,
				templateDir,
				"atlantis",
				false, false,
				logger,
			)
			
			// Check that renderer was created
			if renderer == nil {
				t.Fatal("Expected renderer to be created")
			}
			
			// Check logs contain expected messages
			history := logger.GetHistory()
			for _, expectedLog := range tt.expectedLogs {
				if !strings.Contains(history, expectedLog) {
					t.Errorf("Expected log message containing '%s' not found in history: %s", expectedLog, history)
				}
			}
		})
	}
}

func TestTemplateOverridesExecution(t *testing.T) {
	tmpDir := t.TempDir()
	logger := logging.NewNoopLogger(t).WithHistory()
	
	// Create a custom template with a test marker
	templateDir := filepath.Join(tmpDir, "templates")
	err := os.MkdirAll(templateDir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	
	customTemplate := `{{ define "singleProjectPlanSuccess" -}}
CUSTOM TEMPLATE WORKING
{{ $result := index .Results 0 -}}
Ran {{ .Command }} for {{ if $result.ProjectName }}project: {{ $result.ProjectName }} {{ end }}dir: {{ $result.RepoRelDir }} workspace: {{ $result.Workspace }}
{{ end }}`
	
	err = os.WriteFile(filepath.Join(templateDir, "singleProjectPlanSuccess.tmpl"), []byte(customTemplate), 0644)
	if err != nil {
		t.Fatal(err)
	}
	
	// Create renderer with overrides
	renderer := events.NewMarkdownRenderer(
		false, false, false, false, false, false,
		templateDir,
		"atlantis",
		false, false,
		logger,
	)
	
	// Test execution with a plan result
	ctx := &command.Context{
		Log: logger,
		Pull: models.PullRequest{
			BaseRepo: models.Repo{FullName: "owner/repo"},
		},
	}
	
	result := command.Result{
		ProjectResults: []command.ProjectResult{
			{
				Command:     command.Plan,
				RepoRelDir:  ".",
				Workspace:   "default",
				ProjectName: "test-project",
				PlanSuccess: &models.PlanSuccess{
					TerraformOutput: "No changes. Infrastructure is up-to-date.",
				},
			},
		},
	}
	
	cmd := &events.CommentCommand{
		Name: command.Plan,
	}
	
	// Execute template
	output := renderer.Render(ctx, result, cmd)
	
	// Check that our custom template was used
	if !strings.Contains(output, "CUSTOM TEMPLATE WORKING") {
		t.Errorf("Expected output to contain 'CUSTOM TEMPLATE WORKING', got: %s", output)
	}
	
	// Check that template execution was logged
	history := logger.GetHistory()
	if !strings.Contains(history, "executing template") {
		t.Errorf("Expected log message about executing template not found in history: %s", history)
	}
}

func TestTemplateOverridesNonExistentDirectory(t *testing.T) {
	logger := logging.NewNoopLogger(t).WithHistory()
	
	// Create renderer with non-existent directory
	renderer := events.NewMarkdownRenderer(
		false, false, false, false, false, false,
		"/nonexistent/path",
		"atlantis",
		false, false,
		logger,
	)
	
	// Check that renderer was created (should fall back to built-in templates)
	if renderer == nil {
		t.Fatal("Expected renderer to be created even with non-existent override directory")
	}
	
	// Check logs contain expected warning
	history := logger.GetHistory()
	if !strings.Contains(history, "template override directory does not exist") {
		t.Errorf("Expected warning about non-existent directory not found in history: %s", history)
	}
}

func TestTemplateOverridesOriginalBugReproduction(t *testing.T) {
	// This test reproduces the original bug where templates are never used
	tmpDir := t.TempDir()
	logger := logging.NewNoopLogger(t).WithHistory()
	
	// Create template directory
	templateDir := filepath.Join(tmpDir, "templates")
	err := os.MkdirAll(templateDir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	
	// Create custom template with a clear test marker like in the bug report
	customTemplate := `{{ define "singleProjectPlanSuccess" -}}
TEST TEMPLATE WORKING
{{ $result := index .Results 0 -}}
Ran {{ .Command }} for {{ if $result.ProjectName }}project: ` + "`" + `{{ $result.ProjectName }}` + "`" + ` {{ end }}dir: ` + "`" + `{{ $result.RepoRelDir }}` + "`" + ` workspace: ` + "`" + `{{ $result.Workspace }}` + "`" + `

{{ $result.Rendered }}
{{ end }}`
	
	err = os.WriteFile(filepath.Join(templateDir, "singleProjectPlanSuccess.tmpl"), []byte(customTemplate), 0644)
	if err != nil {
		t.Fatal(err)
	}
	
	// Create renderer with overrides (this should now work!)
	renderer := events.NewMarkdownRenderer(
		false, false, false, false, false, false,
		templateDir,
		"atlantis",
		false, false,
		logger,
	)
	
	// Simulate the exact scenario from the bug report
	ctx := &command.Context{
		Log: logger,
		Pull: models.PullRequest{
			BaseRepo: models.Repo{FullName: "owner/repo"},
		},
	}
	
	result := command.Result{
		ProjectResults: []command.ProjectResult{
			{
				Command:     command.Plan,
				RepoRelDir:  ".",
				Workspace:   "default",
				ProjectName: "test-project",
				PlanSuccess: &models.PlanSuccess{
					TerraformOutput: "No changes. Infrastructure is up-to-date.",
				},
			},
		},
	}
	
	cmd := &events.CommentCommand{
		Name: command.Plan,
	}
	
	// Execute template
	output := renderer.Render(ctx, result, cmd)
	
	// The bug was that "TEST TEMPLATE WORKING" never appeared
	// With our fix, it should now appear!
	if !strings.Contains(output, "TEST TEMPLATE WORKING") {
		t.Errorf("BUG REPRODUCTION FAILED: Expected output to contain 'TEST TEMPLATE WORKING' (the test marker from the original bug report), but it was not found. This means custom templates are still not working. Output was: %s", output)
	}
	
	// Verify logging shows successful override loading
	history := logger.GetHistory()
	if !strings.Contains(history, "template overrides applied successfully") {
		t.Errorf("Expected log message about template overrides being applied successfully. History: %s", history)
	}
}