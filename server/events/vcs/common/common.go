// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

// Package common is used to share common code between all VCS clients without
// running into circular dependency issues.
package common

import (
	"crypto/tls"
	"fmt"
	"github.com/runatlantis/atlantis/server/logging"
	"math"
	"net/http"
)

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
		SepStart:         fmt.Sprintf("%s```diff\n", baseStart),
		TruncationHeader: fmt.Sprintf("%s```diff\n", baseTruncation),
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
	// Track whether we're inside code blocks and details blocks
	inCodeBlock := false
	detailsBlockCount := 0
	inInlineCode := false

	// Look at the text up to the position
	text := comment[:position]

	// Process character by character to handle inline code properly
	i := 0
	for i < len(text) {
		char := text[i]

		// Check for triple backticks (code blocks)
		if i <= len(text)-3 && text[i:i+3] == "```" {
			inCodeBlock = !inCodeBlock
			i += 3
			continue
		}

		// Check for single backticks (inline code) - only if not in a code block
		if char == '`' && !inCodeBlock {
			inInlineCode = !inInlineCode
		}

		// Check for details block markers
		if char == '<' {
			if i <= len(text)-9 && text[i:i+9] == "<details>" {
				detailsBlockCount++
				i += 9
				continue
			}
			if i <= len(text)-10 && text[i:i+10] == "</details>" {
				detailsBlockCount--
				i += 10
				continue
			}
		}

		i++
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
- SplitComment appends the appropriate TruncationHeader to the first comment if it would have produced more comments.
*/
func SplitComment(logger logging.SimpleLogging, comment string, maxSize int, maxCommentsPerCommand int, command string) []string {
	if len(comment) <= maxSize {
		return []string{comment}
	}

	// Generate separators for different closure types
	separators := GenerateSeparatorsFunc(command)

	// Calculate initial estimate for number of comments using a more accurate separator length
	// We'll refine this as we go with per-split calculation
	estimatedSepLength := 30 // More accurate estimate based on typical separator lengths
	maxContentSize := maxSize - estimatedSepLength
	if maxContentSize <= 0 {
		return []string{comment}
	}

	var comments []string
	numPotentialComments := int(math.Ceil(float64(len(comment)) / float64(maxContentSize)))
	var numComments int
	if maxCommentsPerCommand == 0 {
		numComments = numPotentialComments
	} else {
		numComments = min(numPotentialComments, maxCommentsPerCommand)
	}
	isTruncated := numComments < numPotentialComments

	upTo := len(comment)

	for len(comments) < numComments {
		// Detect closure type at the split position
		closureType := detectClosureType(comment, upTo)
		sepSet := separators[closureType]

		// Determine what separators this comment will need based on final array position
		currentCommentIndex := len(comments)
		isFirstCommentInArray := (currentCommentIndex + 1) == numComments // This portion becomes the first comment in final array
		isLastCommentInArray := currentCommentIndex == 0                  // This portion becomes the last comment in final array

		var startSepLength, endSepLength int
		// Calculate startSepLength
		switch {
		case isFirstCommentInArray && isTruncated:
			startSepLength = len(sepSet.TruncationHeader)
		case !isFirstCommentInArray:
			startSepLength = len(sepSet.SepStart)
		default:
			startSepLength = 0
		}

		// Calculate endSepLength
		if isLastCommentInArray {
			endSepLength = 0
		} else {
			endSepLength = len(sepSet.SepEnd)
		}

		// Calculate split position with exact separator lengths
		totalSepLength := startSepLength + endSepLength
		maxContentSize := maxSize - totalSepLength

		if maxContentSize <= 0 {
			return []string{comment}
		}

		downFrom := max(0, upTo-maxContentSize)

		// Skip empty portions
		if downFrom >= upTo {
			break
		}

		portion := comment[downFrom:upTo]

		// Apply the separators we calculated

		// Apply separators in a clear order: start, then end
		switch {
		case isFirstCommentInArray && isTruncated:
			portion = sepSet.TruncationHeader + portion
		case !isFirstCommentInArray:
			portion = sepSet.SepStart + portion
		}

		if !isLastCommentInArray {
			portion += sepSet.SepEnd
		}

		comments = append([]string{portion}, comments...)
		upTo = downFrom
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
