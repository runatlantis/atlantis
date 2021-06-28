// Package common is used to share common code between all VCS clients without
// running into circular dependency issues.
package common

import "strings"

// AutomergeCommitMsg is the commit message Atlantis will use when automatically
// merging pull requests.
const AutomergeCommitMsg = "[Atlantis] Automatically merging after successful apply"

// SplitComment splits comment into a slice of comments that are under maxSize.
// It appends sepEnd to all comments that have a following comment.
// It prepends sepStart to all comments that have a preceding comment.
func SplitComment(comment string, maxSize int, sepEnd string, sepStart string) []string {
	if len(comment) <= maxSize {
		return []string{comment}
	}

	maxWithSep := maxSize - len(sepEnd) - len(sepStart)
	var comments []string
	head := 0
	for head < len(comment) {
		tail := min(len(comment), head+maxWithSep)
		portion := comment[head:tail]
		lastLF := strings.LastIndex(portion, "\n")
		if lastLF > 0 {
			portion = portion[:lastLF]
		}

		rawPortionLength := len(portion)
		portion = strings.TrimSpace(portion)

		if len(portion) > 0 {
			if head > 0 {
				portion = sepStart + portion
			}
			head += rawPortionLength
			if head < len(comment) {
				portion += sepEnd
			}
			comments = append(comments, portion)
		} else {
			head += rawPortionLength
		}
	}
	return comments
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
