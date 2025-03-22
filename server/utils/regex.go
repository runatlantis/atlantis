package utils

import (
	"regexp"
)

// ParseRegex validates and returns a [Regexp] object
func ParseRegex(pattern string) (*regexp.Regexp, error) {
	if pattern == "" {
		return nil, nil
	}

	return regexp.Compile(pattern)
}
