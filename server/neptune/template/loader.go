package template

import (
	"bytes"
	"fmt"
	"html/template"
	"os"

	_ "embed" // embedding files

	"github.com/Masterminds/sprig/v3"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events/models"
)

type Key string

// input for the template to be loaded
// add fields here as necessary
type Input struct {
}

// list of all valid template ids
const (
	LegacyApplyComment = Key("legacyApply")
)

var defaultTemplates = map[Key]string{
	LegacyApplyComment: legacyApplyTemplate,
}

//go:embed templates/legacyApply.tmpl
var legacyApplyTemplate string

type Loader[T any] struct {
	GlobalCfg valid.GlobalCfg
}

func NewLoader[T any](globalCfg valid.GlobalCfg) Loader[T] {
	return Loader[T]{
		GlobalCfg: globalCfg,
	}
}

type Template struct{}

func (l Loader[T]) Load(id Key, repo models.Repo, data T) (string, error) {
	tmpl := template.Must(l.getTemplate(id, repo))

	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, data); err != nil {
		return "", fmt.Errorf("Failed to render template: %v", err)
	}
	return buf.String(), nil
}

func (l Loader[T]) getTemplate(id Key, repo models.Repo) (*template.Template, error) {
	var templateOverrides map[string]string

	repoCfg := l.GlobalCfg.MatchingRepo(repo.ID())
	if repoCfg != nil {
		templateOverrides = repoCfg.TemplateOverrides
	}

	template := template.New("").Funcs(sprig.TxtFuncMap())
	if fileName, ok := templateOverrides[string(id)]; ok {
		if content, err := os.ReadFile(fileName); err == nil {
			return template.Parse(string(content))
		}
	}

	return template.Parse(defaultTemplates[id])
}
