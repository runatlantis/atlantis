// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

// Package common is used to share common code between all VCS clients without
// running into circular dependency issues.
package common

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/runatlantis/atlantis/server/logging"
)

// estimateBuffer is subtracted when estimating the start position for closure-type
// detection. It compensates for the difference between the estimated and actual
// start-separator lengths so the initial guess lands inside (not past) the chunk.
const estimateBuffer = 50

// ClosureType represents the type of markdown closure at a given position
type ClosureType int

const (
	// NoClosure means no special closure is needed
	NoClosure ClosureType = iota
	// CodeBlock means we're inside a code block (```)
	CodeBlock
	// DetailsBlock means we're inside a details block (<details>)
	DetailsBlock
	// CodeInDetails means we're inside a code block within a details block
	CodeInDetails
	// InlineCode means we're inside an inline code block (`)
	InlineCode
)

// closureRegex matches code blocks (3+ backticks), inline code (1 backtick), and details tags (with optional attributes)
var closureRegex = regexp.MustCompile("(`{3,})|(`)|(<details(?:\\s[^>]*)?>)|(</details>)")

// SeparatorSet contains the separators for a specific closure type
type SeparatorSet struct {
	SepEnd           string
	SepStart         string
	TruncationHeader string
}

// GenerateSeparatorsFunc is a variable that holds the separator generation function
// This allows it to be overridden for testing
var GenerateSeparatorsFunc = GenerateSeparators

// GenerateSeparators creates separator sets for different closure types
func GenerateSeparators(command string) map[ClosureType]SeparatorSet {
	separators := make(map[ClosureType]SeparatorSet)

	// Base separators
	baseEnd := "\n<br>\n\n**Warning**: Output length greater than max comment size. Continued in next comment."
	baseStart := "Continued from previous comment.\n"
	if command != "" {
		baseStart = fmt.Sprintf("Continued %s output from previous comment.\n", command)
	}
	baseTruncation := "> [!WARNING]\n> **Warning**: Command output is larger than the maximum number of comments per command. Output truncated.\n"

	// NoClosure separators
	separators[NoClosure] = SeparatorSet{
		SepEnd:           baseEnd,
		SepStart:         baseStart,
		TruncationHeader: baseTruncation,
	}

	// CodeBlock separators
	separators[CodeBlock] = SeparatorSet{
		SepEnd:           fmt.Sprintf("\n```\n%s", baseEnd),
		SepStart:         fmt.Sprintf("%s```diff\n", baseStart),
		TruncationHeader: fmt.Sprintf("%s```diff\n", baseTruncation),
	}

	// DetailsBlock separators
	separators[DetailsBlock] = SeparatorSet{
		SepEnd:           fmt.Sprintf("\n</details>\n%s", baseEnd),
		SepStart:         fmt.Sprintf("%s<details><summary>Show Output</summary>\n\n```diff\n", baseStart),
		TruncationHeader: fmt.Sprintf("%s<details><summary>Show Output</summary>\n\n```diff\n", baseTruncation),
	}

	// CodeInDetails separators
	separators[CodeInDetails] = SeparatorSet{
		SepEnd:           fmt.Sprintf("\n```\n</details>\n%s", baseEnd),
		SepStart:         fmt.Sprintf("%s<details><summary>Show Output</summary>\n\n```diff\n", baseStart),
		TruncationHeader: fmt.Sprintf("%s<details><summary>Show Output</summary>\n\n```diff\n", baseTruncation),
	}

	// InlineCode separators
	separators[InlineCode] = SeparatorSet{
		SepEnd:           fmt.Sprintf("`\n%s", baseEnd),
		SepStart:         fmt.Sprintf("%s`", baseStart),
		TruncationHeader: fmt.Sprintf("%s`", baseTruncation),
	}

	return separators
}

// detectClosureType determines what type of closure is needed at a given position in the comment
func detectClosureType(comment string, position int) ClosureType {
	// Look at the text up to the position
	text := comment[:position]

	// Find all matches for our closure tokens
	matches := closureRegex.FindAllStringIndex(text, -1)

	// Track state
	inCodeBlock := false
	inInlineCode := false
	detailsBlockCount := 0

	for _, loc := range matches {
		token := text[loc[0]:loc[1]]

		if strings.HasPrefix(token, "```") {
			// Code block fence (3+ backticks)
			// Only toggle if we're not inside inline code
			if !inInlineCode {
				inCodeBlock = !inCodeBlock
			}
		} else if token == "`" {
			// Inline code delimiter
			// Only toggle if we're not inside a code block
			if !inCodeBlock {
				inInlineCode = !inInlineCode
			}
		} else if strings.HasPrefix(token, "<details") {
			// Opening details tag
			// Only count if not inside code block or inline code
			if !inCodeBlock && !inInlineCode {
				detailsBlockCount++
			}
		} else if token == "</details>" {
			// Closing details tag
			// Only count if not inside code block or inline code
			if !inCodeBlock && !inInlineCode {
				// Prevent count from going negative (e.g. if we missed an opening tag or saw stray closing tag)
				if detailsBlockCount > 0 {
					detailsBlockCount--
				}
			}
		}
	}

	// Determine closure type based on current state
	if detailsBlockCount > 0 && inCodeBlock {
		return CodeInDetails
	} else if inCodeBlock {
		return CodeBlock
	} else if detailsBlockCount > 0 {
		return DetailsBlock
	} else if inInlineCode {
		return InlineCode
	}

	return NoClosure
}

// AutomergeCommitMsg returns the commit message to use when automerging.
func AutomergeCommitMsg(pullNum int) string {
	return fmt.Sprintf("[Atlantis] Automatically merging after successful apply: PR #%d", pullNum)
}

/*
SplitComment splits comment into a slice of comments that are under maxSize.
- It appends appropriate SepEnd to all comments that have a following comment based on closure type.
- It prepends appropriate SepStart to all comments that have a preceding comment based on closure type.
- If maxCommentsPerCommand is non-zero, it never returns more than maxCommentsPerCommand
comments, and it truncates the beginning of the comment to preserve the end of the comment string,
which usually contains more important information, such as warnings, errors, and the plan summary.
- SplitComment prepends the appropriate TruncationHeader to the first comment if it would have produced more comments.
*/
func SplitComment(logger logging.SimpleLogging, comment string, maxSize int, maxCommentsPerCommand int, command string) []string {
	if len(comment) <= maxSize {
		return []string{comment}
	}

	// Generate separators for different closure types
	separators := GenerateSeparatorsFunc(command)

	var comments []string
	upTo := len(comment)

	for upTo > 0 {
		// Check if this is the last allowed comment (which becomes the first comment in the array)
		// because we are limited by maxCommentsPerCommand
		isLastAllowed := maxCommentsPerCommand > 0 && len(comments) == maxCommentsPerCommand-1

		// 1. Determine End Separator (needed if there are comments after this one)
		closureAtEnd := detectClosureType(comment, upTo)
		endSepSet := separators[closureAtEnd]
		endSepLength := 0
		if len(comments) > 0 {
			endSepLength = len(endSepSet.SepEnd)
		}

		// 2. Determine Start Separator (needed if this is NOT the start of string)
		// We estimate downFrom to guess the closure type at the start
		estimatedDownFrom := min(upTo, max(0, upTo-(maxSize-endSepLength-estimateBuffer)))
		closureAtStart := detectClosureType(comment, estimatedDownFrom)
		startSepSet := separators[closureAtStart]

		startSepLength := len(startSepSet.SepStart)
		if isLastAllowed {
			startSepLength = len(startSepSet.TruncationHeader)
		}

		maxContentSize := maxSize - endSepLength - startSepLength
		if maxContentSize <= 0 {
			// Separators are too large for maxSize. Return as is (best effort).
			return []string{comment}
		}

		downFrom := max(0, upTo-maxContentSize)

		// Optimization: Check if we can fit everything from 0 to upTo without SepStart
		// (Only if we are not forced to truncate)
		if !isLastAllowed && downFrom > 0 {
			// If we assume no Start separator, can we fit from 0?
			if upTo <= (maxSize - endSepLength) {
				downFrom = 0
			}
		}

		// 3. Adjust split point for line boundaries or spaces
		// Prefer splitting at a line boundary. Search forward from downFrom.
		// If no newline found, look for a space.
		// This shrinks the current comment (safe) but leaves more for previous ones.
		if downFrom > 0 && downFrom < len(comment) {
			// whitespaceForwardSearchWindow limits how far ahead we scan when adjusting
			// splits to align with whitespace or newlines. 500 bytes is a compromise
			// between usually finding a natural boundary near the ideal split point
			// and avoiding excessive scanning on very large comments.
			const whitespaceForwardSearchWindow = 500

			searchLimit := min(upTo-1, downFrom+whitespaceForwardSearchWindow)
			foundSpace := -1
			foundNewline := -1

			for p := downFrom; p < searchLimit; p++ {
				if comment[p] == '\n' {
					foundNewline = p
					break
				}
				if comment[p] == ' ' && foundSpace == -1 {
					foundSpace = p
				}
			}

			if foundNewline != -1 {
				downFrom = foundNewline + 1
			} else if foundSpace != -1 {
				downFrom = foundSpace + 1
			}
		}

		// 4. Re-calculate closure at actual downFrom and ensure we fit maxSize
		// If closure type changed or we underestimated header size, we might overflow.
		// Iterate to stabilize.
		for {
			closureAtStart = detectClosureType(comment, downFrom)
			startSepSet = separators[closureAtStart]

			actualStartSepLength := 0
			if downFrom > 0 {
				if isLastAllowed {
					actualStartSepLength = len(startSepSet.TruncationHeader)
				} else {
					actualStartSepLength = len(startSepSet.SepStart)
				}
			}

			currentTotalSize := (upTo - downFrom) + actualStartSepLength + endSepLength
			if currentTotalSize <= maxSize {
				break
			}

			// We overflowed. Shrink the chunk by increasing downFrom.
			overflow := currentTotalSize - maxSize
			downFrom = min(upTo, downFrom+overflow)
		}

		// 5. Construct the comment portion
		portion := comment[downFrom:upTo]

		if downFrom > 0 {
			if isLastAllowed {
				portion = startSepSet.TruncationHeader + portion
			} else {
				// Re-verify if we really needed SepStart (logic above handled downFrom=0)
				// If downFrom > 0, we definitely need it.
				portion = startSepSet.SepStart + portion
			}
		}

		if len(comments) > 0 {
			portion += endSepSet.SepEnd
		}

		comments = append([]string{portion}, comments...)
		upTo = downFrom

		// If we hit the limit and still have content left (downFrom > 0), stop.
		// The current comment (comments[0]) already has TruncationHeader.
		if isLastAllowed && downFrom > 0 {
			break
		}
	}

	return comments
}

// disableSSLVerification disables ssl verification for the global http client
// and returns a function to be called in a defer that will re-enable it.
func DisableSSLVerification() func() {
	orig := http.DefaultTransport.(*http.Transport).TLSClientConfig
	// nolint: gosec
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	return func() {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = orig
	}
}
