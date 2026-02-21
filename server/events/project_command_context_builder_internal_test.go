// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"os/exec"
	"strings"
	"testing"

	. "github.com/runatlantis/atlantis/testing"
)

// TestEscapeArgs_ASCII verifies that ASCII shell metacharacters are properly
// escaped so they cannot be interpreted as shell syntax when the resulting
// string is passed to "sh -c".
func TestEscapeArgs_ASCII(t *testing.T) {
	cases := []struct {
		desc     string
		args     []string
		expected []string
	}{
		{
			desc:     "plain alphanumeric",
			args:     []string{"arg1", "arg2"},
			expected: []string{`\a\r\g\1`, `\a\r\g\2`},
		},
		{
			desc:     "command substitution via dollar-paren",
			args:     []string{"-var=$(touch /tmp/pwned)"},
			expected: []string{`\-\v\a\r\=\$\(\t\o\u\c\h\ \/\t\m\p\/\p\w\n\e\d\)`},
		},
		{
			desc:     "backtick command substitution",
			args:     []string{"`id`"},
			expected: []string{"\\`\\i\\d\\`"},
		},
		{
			desc:     "semicolon separator",
			args:     []string{"-- ;echo bad"},
			expected: []string{`\-\-\ \;\e\c\h\o\ \b\a\d`},
		},
		{
			desc:     "pipe",
			args:     []string{"foo|bar"},
			expected: []string{`\f\o\o\|\b\a\r`},
		},
		{
			desc:     "ampersand",
			args:     []string{"foo&&bar"},
			expected: []string{`\f\o\o\&\&\b\a\r`},
		},
		{
			desc:     "redirection",
			args:     []string{"foo>bar"},
			expected: []string{`\f\o\o\>\b\a\r`},
		},
		{
			desc:     "newline in arg (unlikely but possible)",
			args:     []string{"foo\nbar"},
			expected: []string{"\\f\\o\\o\\\n\\b\\a\\r"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			got := escapeArgs(tc.args)
			Equals(t, tc.expected, got)
		})
	}
}

// TestEscapeArgs_MultibyteSafety verifies that multibyte UTF-8 characters are
// fully escaped (every constituent byte receives a backslash prefix) so that
// no byte of a multibyte sequence is silently dropped or corrupted.
func TestEscapeArgs_MultibyteSafety(t *testing.T) {
	// U+00E9 LATIN SMALL LETTER E WITH ACUTE ("é"), UTF-8: 0xC3 0xA9
	// Correct byte-level escaping must produce a backslash before each of
	// the two bytes, giving: '\' 0xC3 '\' 0xA9.
	input := "é" // two-byte UTF-8 sequence
	result := escapeArgs([]string{input})
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	escaped := result[0]

	// The escaped string must have exactly 4 bytes: \<0xC3>\<0xA9>
	if len(escaped) != 4 {
		t.Errorf("expected 4 bytes in escaped output, got %d; escaped=%q", len(escaped), escaped)
	}

	// Every byte must be preceded by a backslash.
	rawInput := []byte(input) // 0xC3 0xA9
	for i, b := range rawInput {
		offset := i * 2
		if offset+1 >= len(escaped) {
			t.Errorf("escaped output too short at byte index %d", i)
			break
		}
		if escaped[offset] != '\\' {
			t.Errorf("byte %d: expected backslash at escaped[%d], got %q", i, offset, escaped[offset])
		}
		if escaped[offset+1] != b {
			t.Errorf("byte %d: expected 0x%02X at escaped[%d], got 0x%02X", i, b, offset+1, escaped[offset+1])
		}
	}
}

// TestEscapeArgs_ShellSafety uses the real shell to confirm that an escaped
// argument is never evaluated as a shell command substitution.
// The test only runs when "sh" is available on the system.
func TestEscapeArgs_ShellSafety(t *testing.T) {
	shPath, err := exec.LookPath("sh")
	if err != nil {
		t.Skip("sh not found, skipping shell-safety test")
	}

	injectionAttempts := []string{
		"$(touch /tmp/atlantis_pwned)",
		"`touch /tmp/atlantis_pwned`",
		"$(id)",
		"; id",
		"| id",
		"&& id",
	}

	for _, attempt := range injectionAttempts {
		t.Run(attempt, func(t *testing.T) {
			escaped := escapeArgs([]string{attempt})
			if len(escaped) != 1 {
				t.Fatalf("expected 1 escaped arg, got %d", len(escaped))
			}

			// Build a shell command that echoes the escaped arg.
			// If the escaping is correct, the output should equal the literal
			// attempt string (with leading/trailing whitespace trimmed) rather
			// than the output of any injected command.
			shCmd := "echo " + escaped[0]
			out, err := exec.Command(shPath, "-c", shCmd).Output()
			if err != nil {
				t.Fatalf("shell command failed: %v", err)
			}
			got := strings.TrimRight(string(out), "\n")
			if got != attempt {
				t.Errorf("shell evaluated injected content!\nwanted literal: %q\ngot:            %q", attempt, got)
			}
		})
	}
}
