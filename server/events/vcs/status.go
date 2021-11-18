package vcs

import (
	"fmt"
	"strings"
)

type StatusTitleMatcher struct {
	TitlePrefix string
}

func (m StatusTitleMatcher) MatchesCommand(title string, command string) bool {
	return strings.HasPrefix(title, fmt.Sprintf("%s/%s", m.TitlePrefix, command))
}

type StatusTitleBuilder struct {
	TitlePrefix string
}

type StatusTitleOptions struct {
	ProjectName string
}

func (b StatusTitleBuilder) Build(command string, options ...StatusTitleOptions) string {
	src := fmt.Sprintf("%s/%s", b.TitlePrefix, command)

	var projectName string
	for _, opt := range options {
		if opt.ProjectName != "" {
			projectName = opt.ProjectName
		}
	}

	if projectName != "" {
		src = fmt.Sprintf("%s: %s", src, projectName)
	}

	return src
}
