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

// SplitComment splits comment into a slice of comments that are under maxSize.
// It appends sepEnd to all comments that have a following comment.
// It prepends sepStart to all comments that have a preceding comment.
// If maxCommentsPerCommand is non-zero, it never returns more than maxCommentsPerCommand
// comments, and it appends truncationFooter to the final comment if it would have
// produced more comments.
func SplitComment(comment string, maxSize int, sepEnd string, sepStart string, maxCommentsPerCommand int, truncationFooter string) []string {
	if len(comment) <= maxSize {
		return []string{comment}
	}

	// No comment contains both sepEnd and truncationFooter, so we only have to count their max.
	maxWithSep := maxSize - max(len(sepEnd), len(truncationFooter)) - len(sepStart)
	var comments []string
	numPotentialComments := int(math.Ceil(float64(len(comment)) / float64(maxWithSep)))
	var numComments int
	if maxCommentsPerCommand == 0 {
		numComments = numPotentialComments
	} else {
		numComments = min(numPotentialComments, maxCommentsPerCommand)
	}

	for i := 0; i < numComments; i++ {
		upTo := min(len(comment), (i+1)*maxWithSep)
		portion := comment[i*maxWithSep : upTo]
		if i < numComments-1 {
			portion += sepEnd
		} else if i == numComments-1 && numComments < numPotentialComments {
			portion += truncationFooter
		}
		if i > 0 {
			portion = sepStart + portion
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

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
