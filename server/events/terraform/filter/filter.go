package filter

import (
	"regexp"
)

type LogFilter struct {
	Regexes []*regexp.Regexp
}

func (l *LogFilter) ShouldFilterLine(message string) bool {
	for _, regex := range l.Regexes {
		if regex.MatchString(message) {
			return true
		}
	}
	return false
}
