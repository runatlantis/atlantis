// Package common is used to share common code between all VCS clients without
// running into circular dependency issues.
package common

import (
	"math"
)

// AutomergeCommitMsg is the commit message Atlantis will use when automatically
// merging pull requests.
const AutomergeCommitMsg = "[Atlantis] Automatically merging after successful apply"

// SepStartComment is the first line of comment when a plan needs to be extended to
// multiple comments.
const SepStartComment string = "Continued from previous comment.\n<details><summary>Show Output</summary>\n\n```diff\n"

// SepStartComment is the last line of the first comment when a plan needs to be
// extended to multiple comments.
const SepEndComment string = "\n```\n</details>" +
	"\n<br>\n\n**Warning**: Output length greater than max comment size. Continued in next comment."

// SplitComment splits comment into a slice of comments that are under maxSize.
// It appends SepEndComment to all comments that have a following comment.
// It prepends SepStartComment to all comments that have a preceding comment.
func SplitComment(comment string, maxSize int) []string {
	if len(comment) <= maxSize {
		return []string{comment}
	}

	maxWithSep := maxSize - len(SepEndComment) - len(SepStartComment)
	var comments []string
	numComments := int(math.Ceil(float64(len(comment)) / float64(maxWithSep)))
	for i := 0; i < numComments; i++ {
		upTo := min(len(comment), (i+1)*maxWithSep)
		portion := comment[i*maxWithSep : upTo]
		if i < numComments-1 {
			portion += SepEndComment
		}
		if i > 0 {
			portion = SepStartComment + portion
		}
		comments = append(comments, portion)
	}
	return comments
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
