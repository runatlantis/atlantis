package vcs_test

import (
	"testing"

	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/stretchr/testify/assert"
)

func TestMatches(t *testing.T) {

	t.Run("in sync with builder", func(t *testing.T) {
		titlePrefix := "atlantis-test"
		command := "apply"
		builder := vcs.StatusTitleBuilder{TitlePrefix: titlePrefix}
		matcher := vcs.StatusTitleMatcher{TitlePrefix: titlePrefix}

		title := builder.Build(command)

		assert.True(t, matcher.MatchesCommand(title, command))
	})

	t.Run("incorrect command", func(t *testing.T) {
		titlePrefix := "atlantis-test"
		builder := vcs.StatusTitleBuilder{TitlePrefix: titlePrefix}
		matcher := vcs.StatusTitleMatcher{TitlePrefix: titlePrefix}

		title := builder.Build("apply")

		assert.False(t, matcher.MatchesCommand(title, "plan"))
	})
}
