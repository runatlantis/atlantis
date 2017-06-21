package server

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

// GithubCommentRenderer renders responses as GitHub comments
type GithubCommentRenderer struct{}

// CompiledTemplate represents a single template with both its source text and its compiled
// template. We compile the templates in init()
type CompiledTemplate struct {
	text     string
	Template *template.Template
}

// PathResultRendered is used as an intermediary data container. We render the individual path results
// into Render and then pass around this struct to be rendered into the main templates
type PathResultRendered struct {
	PathResult
	Rendered string
}

// todo: once this is in its own package don't need to append Tmpl
var (
	// If you add a template here, be sure to add it to init() so it gets compiled
	SetupFailure *CompiledTemplate = &CompiledTemplate{
		text: "**{{.Command}} Failed**:\n{{.Output}}\n" + logTmpl,
	}
	SinglePath *CompiledTemplate = &CompiledTemplate{
		text: "{{ range $result := .Results }}{{$result.Rendered}}{{end}}\n" + logTmpl,
	}
	MultiPath *CompiledTemplate = &CompiledTemplate{
		text: "Ran {{.Command}} in {{ len .Results }} directories:\n" +
			"{{ range $path, $result := .Results }}" +
			" * `{{$path}}`\n" + //todo: add result status
			"{{end}}\n" +
			"{{ range $path, $result := .Results }}" +
			"Terraform {{$.Command}} for `{{$path}}`:\n" +
			"{{$result.Rendered}}\n\n" +
			"---\n{{end}}" +
			logTmpl,
	}
	PlanSuccessTmpl *CompiledTemplate = &CompiledTemplate{
		text: "```diff\n" +
			"{{.TerraformOutput}}\n" +
			"```\n\n" +
			"* To **discard** this plan click [here]({{.LockURL}}).",
	}
	RunLockedFailureTmpl *CompiledTemplate = &CompiledTemplate{
		text: "This plan is currently locked by #{{.LockingPullNum}}\n" +
			"The locking plan must be applied or discarded before future plans can execute.",
	}
	TerraformFailureTmpl *CompiledTemplate = &CompiledTemplate{
		text: "**Atlantis encountered an error while running...**\n" +
			"```\n" +
			"$ {{.Command}}\n" +
			"{{.Output}}\n" +
			"```",
	}
	EnvironmentFileNotFoundFailureTmpl *CompiledTemplate = &CompiledTemplate{
		text: "Environment file did not exist {{.Filename}}",
	}
	EnvironmentErrorTmpl *CompiledTemplate = &CompiledTemplate{
		text: "Please specify environment variable while running plan\n" +
			"For example: `atlantis plan {environment_name}`\n" +
			"*Environments that are available can be found under the `env/` folder of the terraform stack.*",
	}
	ApplySuccessTmpl *CompiledTemplate = &CompiledTemplate{
		text: "```diff\n" +
			"{{.Output}}\n" +
			"```",
	}
	ApplyFailureTmpl *CompiledTemplate = &CompiledTemplate{
		text: "**Apply Failed**:\n" +
			"```bash\n" +
			"$ {{.Command}}\n" +
			"{{.Output}}\n" +
			"{{.ErrorMessage}}\n" +
			"```",
	}
	PullNotApprovedFailureTmpl *CompiledTemplate = &CompiledTemplate{
		text: "Pull Request must be **Approved** before running apply.",
	}
	NoPlansFailureTmpl *CompiledTemplate = &CompiledTemplate{
		text: "0 plans found",
	}
	ErrorTmpl *CompiledTemplate = &CompiledTemplate{
		text: "**Atlantis encountered an error:**\n" +
			"```\n" +
			"{{.Error}}\n" +
			"```\n" +
			"Log:\n" +
			"```\n" +
			"{{.Log}}```",
	}
	GeneralErrorTmpl *CompiledTemplate = &CompiledTemplate{
		text: "{{.Error}}",
	}
)

var logTmpl = "{{if .Verbose}}\nAtlantis Log:\n```\n{{.Log}}```{{end}}\n"

func init() {
	// compile the templates
	for _, t := range []*CompiledTemplate{
		SetupFailure,
		SinglePath,
		MultiPath,
		PlanSuccessTmpl,
		RunLockedFailureTmpl,
		TerraformFailureTmpl,
		EnvironmentFileNotFoundFailureTmpl,
		EnvironmentErrorTmpl,
		ApplySuccessTmpl,
		ApplyFailureTmpl,
		PullNotApprovedFailureTmpl,
		NoPlansFailureTmpl,
		ErrorTmpl,
		GeneralErrorTmpl,
	} {
		t.Template = template.Must(template.New("").Parse(t.text))
	}
}

func (g *GithubCommentRenderer) render(res ExecutionResult, log string, verbose bool) string {
	commandStr := strings.Title(res.Command.String())
	if res.SetupError != nil {
		renderedError := g.renderTemplate(res.SetupError.Template().Template, res.SetupError)
		return g.renderTemplate(ErrorTmpl.Template, struct {
			Error string
			Log   string
		}{renderedError, log})
	} else if res.SetupFailure != nil {
		renderedFailure := g.renderTemplate(res.SetupFailure.Template().Template, res.SetupFailure)
		return g.renderTemplate(SetupFailure.Template, struct {
			Command string
			Output  string
			Log     string
			Verbose bool
		}{commandStr, renderedFailure, log, verbose})
	} else {
		hasErrors := false
		for _, res := range res.PathResults {
			if res.Status == Error {
				hasErrors = true
			}
		}
		return g.renderPathOutputs(res.PathResults, commandStr, log, hasErrors || verbose)
	}
}

func (g *GithubCommentRenderer) renderPathOutputs(pathResults []PathResult, command string, log string, verbose bool) string {
	renderedOutputs := map[string]PathResultRendered{}
	for _, result := range pathResults {
		renderedOutputs[result.Path] = PathResultRendered{result, g.renderTemplate(result.Result.Template().Template, result.Result)}
	}

	var tmpl *template.Template
	if len(renderedOutputs) == 1 {
		tmpl = SinglePath.Template
	} else {
		tmpl = MultiPath.Template
	}
	return g.renderTemplate(tmpl, struct {
		Results map[string]PathResultRendered
		Log     string
		Verbose bool
		Command string
	}{renderedOutputs, log, verbose, command})
}

func (g *GithubCommentRenderer) renderTemplate(tmpl *template.Template, data interface{}) string {
	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, data); err != nil {
		return fmt.Sprintf("Failed to render template, this is a bug: %v", err)
	}
	return buf.String()
}
