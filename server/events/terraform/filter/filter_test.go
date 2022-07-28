package filter_test

import (
	"github.com/runatlantis/atlantis/server/events/terraform/filter"
	"github.com/stretchr/testify/assert"
	"regexp"
	"testing"
)

func TestLogFilter_ShouldFilter(t *testing.T) {
	regex := regexp.MustCompile("abc*")
	filter := filter.LogFilter{
		Regexes: []*regexp.Regexp{regex},
	}
	assert.True(t, filter.ShouldFilterLine("abcd"))
}

func TestLogFilter_ShouldNotFilter(t *testing.T) {
	regex := regexp.MustCompile("abc*")
	filter := filter.LogFilter{
		Regexes: []*regexp.Regexp{regex},
	}
	assert.False(t, filter.ShouldFilterLine("efg"))
}
