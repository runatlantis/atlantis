// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"fmt"
	"strings"
	"unicode"
)

// CommentNamespace identifies comments created by one Atlantis instance when
// multiple instances share the same VCS identity.
type CommentNamespace struct {
	value string
}

// NewCommentNamespace returns a comment namespace for value. An empty value
// disables comment namespacing and preserves legacy comment behavior.
func NewCommentNamespace(value string) CommentNamespace {
	return CommentNamespace{value: value}
}

// Enabled reports whether comments should be tagged and matched by namespace.
func (n CommentNamespace) Enabled() bool {
	return n.value != ""
}

// Marker returns the hidden marker for command, or an empty string when
// namespacing is disabled.
func (n CommentNamespace) Marker(command string) string {
	if !n.Enabled() {
		return ""
	}
	return fmt.Sprintf("<!-- atlantis-comment:v1 namespace=%s command=%s -->", n.value, normalizeCommentCommand(command))
}

// Tag appends the namespace marker as the final line of body. It returns body
// unchanged when namespacing is disabled or body already has the marker.
func (n CommentNamespace) Tag(body, command string) string {
	marker := n.Marker(command)
	if marker == "" {
		return body
	}
	if markerCommand, ok := n.markerCommand(body); ok && markerCommand == normalizeCommentCommand(command) {
		return body
	}
	if strings.HasSuffix(body, "\n") {
		return body + marker
	}
	return body + "\n" + marker
}

// Owns reports whether body's final marker belongs to this namespace. A
// disabled namespace imposes no ownership check so providers retain legacy
// matching.
func (n CommentNamespace) Owns(body string) bool {
	if !n.Enabled() {
		return true
	}
	_, ok := n.markerCommand(body)
	return ok
}

// MatchesCommand reports whether body belongs to command. A non-empty marker
// command is authoritative. Empty marker commands and disabled namespaces use
// the legacy first-line match.
func (n CommentNamespace) MatchesCommand(body, command string) bool {
	if n.Enabled() {
		markerCommand, ok := n.markerCommand(body)
		if !ok {
			return false
		}
		if markerCommand != "" {
			return markerCommand == normalizeCommentCommand(command)
		}
	}
	return firstLineMatchesCommand(body, command)
}

// ContentLimit reserves enough space to append a marker and newline while
// keeping a comment within maxLength. It returns maxLength when disabled.
func (n CommentNamespace) ContentLimit(maxLength int, command string) int {
	marker := n.Marker(command)
	if marker == "" {
		return maxLength
	}
	return maxLength - len(marker) - 1
}

func (n CommentNamespace) markerCommand(body string) (string, bool) {
	trimmedBody := strings.TrimRight(body, "\r\n")
	lastLineStart := strings.LastIndexByte(trimmedBody, '\n') + 1
	markerLine := trimmedBody[lastLineStart:]

	payload, ok := strings.CutPrefix(markerLine, "<!-- atlantis-comment:v1 namespace=")
	if !ok {
		return "", false
	}
	payload, ok = strings.CutSuffix(payload, " -->")
	if !ok {
		return "", false
	}
	namespace, command, ok := strings.Cut(payload, " command=")
	if !ok || namespace != n.value {
		return "", false
	}
	return command, true
}

func firstLineMatchesCommand(body, command string) bool {
	firstLine, _, _ := strings.Cut(body, "\n")
	return strings.Contains(strings.ToLower(firstLine), strings.ToLower(command))
}

func normalizeCommentCommand(command string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			return unicode.ToLower(r)
		}
		return -1
	}, command)
}
