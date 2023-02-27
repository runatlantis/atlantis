package utils

import (
	"github.com/agext/levenshtein"
)

// IsSimilarWord calculates "The Levenshtein Distance" between two strings which
// represents the minimum total cost of edits that would convert the first string
// into the second. If the distance is less than 3, the word is considered misspelled.
func IsSimilarWord(given string, suggestion string) bool {
	dist := levenshtein.Distance(given, suggestion, nil)
	if dist > 0 && dist < 3 {
		return true
	}

	return false
}
