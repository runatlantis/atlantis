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
	if marker == "" || n.Owns(body, command) {
		return body
	}
	if strings.HasSuffix(body, "\n") {
		return body + marker
	}
	return body + "\n" + marker
}

// Owns reports whether body belongs to this namespace and command. A disabled
// namespace imposes no ownership check so providers retain legacy matching.
func (n CommentNamespace) Owns(body, command string) bool {
	marker := n.Marker(command)
	if marker == "" {
		return true
	}
	return strings.HasSuffix(strings.TrimRight(body, "\r\n"), marker)
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

func normalizeCommentCommand(command string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			return unicode.ToLower(r)
		}
		return -1
	}, command)
}
