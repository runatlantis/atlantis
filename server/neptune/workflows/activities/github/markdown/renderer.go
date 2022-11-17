package markdown

import (
	"bytes"
	_ "embed" //embedding files
	"fmt"
	"html/template"

	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/terraform/state"
)

//go:embed templates/checkrun.tmpl
var checkrunTemplateStr string

// panics if we can't read the template
var checkrunTemplate = template.Must(template.New("").Parse(checkrunTemplateStr))

type checkrunTemplateData struct {
	PlanStatus    string
	PlanLogURL    string
	ApplyStatus   string
	ApplyLogURL   string
	InternalError bool
	TimedOut      bool
}

func RenderWorkflowStateTmpl(workflowState *state.Workflow) string {
	planStatus, planLogURL := getJobStatusAndOutput(workflowState.Plan)
	applyStatus, applyLogURL := getJobStatusAndOutput(workflowState.Apply)

	// we can probably pass in the completion reason but i like doing all the boolean
	// checking here if we can instead of in the template.
	internalError := workflowState.Result.Reason == state.InternalServiceError
	timedOut := workflowState.Result.Reason == state.TimedOutError

	return renderTemplate(checkrunTemplate, checkrunTemplateData{
		PlanStatus:    planStatus,
		PlanLogURL:    planLogURL,
		ApplyStatus:   applyStatus,
		ApplyLogURL:   applyLogURL,
		InternalError: internalError,
		TimedOut:      timedOut,
	})
}

func getJobStatusAndOutput(jobState *state.Job) (string, string) {
	var status string
	var output string

	if jobState == nil {
		return status, output
	}

	return string(jobState.Status), jobState.Output.URL.String()
}

func renderTemplate(tmpl *template.Template, data interface{}) string {
	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, data); err != nil {
		return fmt.Sprintf("Failed to render template, this is a bug: %v. Dumping the current data object as is: %s", err, data)
	}
	return buf.String()
}
