// Package common is used to share common code between all VCS clients without
// running into circular dependency issues.
package common

import (
	"fmt"
	"math"
)

// AutomergeCommitMsg returns the commit message to use when automerging.
func AutomergeCommitMsg(pullNum int) string {
	return fmt.Sprintf("[Atlantis] Automatically merging after successful apply: PR #%d", pullNum)
}

/*
SplitComment splits comment into a slice of comments that are under maxSize.
- It appends sepEnd to all comments that have a following comment.
- It prepends sepStart to all comments that have a preceding comment.
- If maxCommentsPerCommand is non-zero, it never returns more than maxCommentsPerCommand
comments, and it truncates the beginning of the comment to preserve the end of the comment string,
which usually contains more important information, such as warnings, errors, and the plan summary.
- SplitComment appends the truncationHeader to the first comment if it would have produced more comments.
*/
func SplitComment(comment string, maxSize int, sepEnd string, sepStart string, maxCommentsPerCommand int, truncationHeader string) []string {
	if len(comment) <= maxSize {
		return []string{comment}
	}

	// No comment contains both sepEnd and truncationHeader, so we only have to count their max.
	maxWithSep := maxSize - max(len(sepEnd), len(truncationHeader)) - len(sepStart)
	var comments []string
	numPotentialComments := int(math.Ceil(float64(len(comment)) / float64(maxWithSep)))
	var numComments int
	if maxCommentsPerCommand == 0 {
		numComments = numPotentialComments
	} else {
		numComments = min(numPotentialComments, maxCommentsPerCommand)
	}
	isTruncated := numComments < numPotentialComments
	upTo := len(comment)
	for len(comments) < numComments {
		downFrom := max(0, upTo-maxWithSep)
		portion := comment[downFrom:upTo]
		if len(comments)+1 != numComments {
			portion = sepStart + portion
		} else if len(comments)+1 == numComments && isTruncated {
			portion = truncationHeader + sepStart + portion
		}
		if len(comments) != 0 {
			portion = portion + sepEnd
		}
		comments = append([]string{portion}, comments...)
		upTo = downFrom
	}
	return comments
}
