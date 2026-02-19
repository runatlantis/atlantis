// Copyright 2017 HootSuite Media Inc.
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// Modified hereafter by contributors to runatlantis/atlantis.

package common_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/runatlantis/atlantis/server/events/vcs/common"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

// Test dynamic separator functionality with table-driven tests
func TestSplitComment_DynamicSeparators(t *testing.T) {
	logger := logging.NewNoopLogger(t)

	// Override separators for testing with shorter, predictable values
	originalGenerateSeparators := common.GenerateSeparatorsFunc
	common.GenerateSeparatorsFunc = func(command string) map[common.ClosureType]common.SeparatorSet {
		logger.Debug("Generating separators for command: %s", command)
		baseStart := "<!-- START -->"
		if command != "" {
			baseStart = fmt.Sprintf("<!-- START %s -->", command)
		}
		return map[common.ClosureType]common.SeparatorSet{
			common.NoClosure: {
				SepEnd:           "<!-- END -->",
				SepStart:         baseStart,
				TruncationHeader: "<!-- TRUNCATED -->",
			},
			common.CodeBlock: {
				SepEnd:           "```\n<!-- END -->",
				SepStart:         fmt.Sprintf("```diff\n%s", baseStart),
				TruncationHeader: "```diff\n<!-- TRUNCATED -->",
			},
			common.DetailsBlock: {
				SepEnd:           "</details>\n<!-- END -->",
				SepStart:         fmt.Sprintf("<details><summary>Show Output</summary>\n%s", baseStart),
				TruncationHeader: "<details><summary>Show Output</summary>\n<!-- TRUNCATED -->",
			},
			common.CodeInDetails: {
				SepEnd:           "```\n</details>\n<!-- END -->",
				SepStart:         fmt.Sprintf("```diff\n%s", baseStart),
				TruncationHeader: "```diff\n<!-- TRUNCATED -->",
			},
			common.InlineCode: {
				SepEnd:           "`\n<!-- END -->",
				SepStart:         fmt.Sprintf("%s`", baseStart),
				TruncationHeader: "<!-- TRUNCATED -->`",
			},
		}
	}
	defer func() { common.GenerateSeparatorsFunc = originalGenerateSeparators }()

	tests := []struct {
		name             string
		comment          string
		maxSize          int
		maxComments      int
		command          string
		expectedCount    int
		expectedComments []string
	}{
		{
			name:          "UnderMax - Comment under max size",
			comment:       "comment under max size",
			maxSize:       50,
			maxComments:   0,
			command:       "plan",
			expectedCount: 1,
			expectedComments: []string{
				"comment under max size",
			},
		},
		{
			name:          "TwoComments - Split into exactly 2 comments",
			comment:       strings.Repeat("a", 1000),
			maxSize:       999,
			maxComments:   0,
			command:       "plan",
			expectedCount: 2,
			expectedComments: []string{
				strings.Repeat("a", 20) + "<!-- END -->",
				"<!-- START plan -->" + strings.Repeat("a", 980),
			},
		},
		{
			name:          "FourComments - Split into multiple comments",
			comment:       strings.Repeat("a", 1000),
			maxSize:       300,
			maxComments:   0,
			command:       "plan",
			expectedCount: 4,
			expectedComments: []string{
				strings.Repeat("a", 181) + "<!-- END -->",
				"<!-- START plan -->" + strings.Repeat("a", 269) + "<!-- END -->",
				"<!-- START plan -->" + strings.Repeat("a", 269) + "<!-- END -->",
				"<!-- START plan -->" + strings.Repeat("a", 281),
			},
		},
		{
			name:          "Limited - Truncation with comment limit",
			comment:       strings.Repeat("a", 1000),
			maxSize:       300,
			maxComments:   2,
			command:       "plan",
			expectedCount: 2,
			expectedComments: []string{
				"<!-- TRUNCATED -->" + strings.Repeat("a", 270) + "<!-- END -->",
				"<!-- START plan -->" + strings.Repeat("a", 281),
			},
		},
		{
			name:          "NoClosure - Basic text splitting",
			comment:       "This is a long comment that will be split. " + strings.Repeat("This is additional content to make the comment longer so it will be split. ", 5),
			maxSize:       200,
			maxComments:   0,
			command:       "plan",
			expectedCount: 3,
			expectedComments: []string{
				"This is a long comment that will be split. This is additional conten<!-- END -->",
				"<!-- START plan -->t to make the comment longer so it will be split. This is additional content to make the comment longer so it will be split. This is additional content to make the comme<!-- END -->",
				"<!-- START plan -->nt longer so it will be split. This is additional content to make the comment longer so it will be split. This is additional content to make the comment longer so it will be split. ",
			},
		},
		{
			name:          "CodeBlock - Splitting within code block",
			comment:       "Here's some code:\n```\nterraform plan\noutput here\n```\nAnd more text. " + strings.Repeat("This is additional content to make the comment longer so it will be split. ", 3),
			maxSize:       200,
			maxComments:   0,
			command:       "plan",
			expectedCount: 2,
			expectedComments: []string{
				"Here's some code:\n```\nterraform plan\noutput here\n```\nAnd more text. This is additional content to make the comme<!-- END -->",
				"<!-- START plan -->nt longer so it will be split. This is additional content to make the comment longer so it will be split. This is additional content to make the comment longer so it will be split. ",
			},
		},
		{
			name:          "DetailsBlock - Splitting within details block",
			comment:       "<details><summary>Show Output</summary>\n\nSome details content here. " + strings.Repeat("This is additional content to make the comment longer so it will be split. ", 4) + "\n</details>",
			maxSize:       200,
			maxComments:   0,
			command:       "plan",
			expectedCount: 3,
			expectedComments: []string{
				"<details><summary>Show Output</summary>\n\nSome details content here. This is addi</details>\n<!-- END -->",
				"<details><summary>Show Output</summary>\n<!-- START plan -->tional content to make the comment longer so it will be split. This is additional content to make the comment longer s</details>\n<!-- END -->",
				"<!-- START plan -->o it will be split. This is additional content to make the comment longer so it will be split. This is additional content to make the comment longer so it will be split. \n</details>",
			},
		},
		{
			name:          "CodeInDetails - Splitting within code block inside details",
			comment:       "<details><summary>Show Output</summary>\n\n```\nterraform apply\nsome output\n```\n</details>\nMore content. " + strings.Repeat("This is additional content to make the comment longer so it will be split. ", 3),
			maxSize:       200,
			maxComments:   0,
			command:       "apply",
			expectedCount: 2,
			expectedComments: []string{
				"<details><summary>Show Output</summary>\n\n```\nterraform apply\nsome output\n```\n</details>\nMore content. This is additional content to make the commen<!-- END -->",
				"<!-- START apply -->t longer so it will be split. This is additional content to make the comment longer so it will be split. This is additional content to make the comment longer so it will be split. ",
			},
		},
		{
			name:          "InlineCode - Splitting within inline code block",
			comment:       "Here is some text with a very long inline code: `" + strings.Repeat("some_very_long_function_name_", 15) + "` and more text after.",
			maxSize:       200,
			maxComments:   0,
			command:       "plan",
			expectedCount: 3,
			expectedComments: []string{
				"Here is some text with a very long inline code: `some_very_long_function_name_some_very_long_function_name_some_very_long_function_name_some_very_long_function`\n<!-- END -->",
				"<!-- START plan -->`_name_some_very_long_function_name_some_very_long_function_name_some_very_long_function_name_some_very_long_function_name_some_very_long_function_name_some_very_long_`\n<!-- END -->",
				"<!-- START plan -->function_name_some_very_long_function_name_some_very_long_function_name_some_very_long_function_name_some_very_long_function_name_some_very_long_function_name_` and more text after.",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			split := common.SplitComment(logger, tt.comment, tt.maxSize, tt.maxComments, tt.command)

			Assert(t, len(split) == tt.expectedCount, "Expected %d comments, got %d", tt.expectedCount, len(split))

			// Compare each comment with expected output
			for i, expected := range tt.expectedComments {
				if i < len(split) {
					Assert(t, split[i] == expected,
						"Comment %d mismatch:\nExpected: %s\nGot:      %s",
						i, expected, split[i])
				}
				// Verify comment doesn't exceed maxSize
				if len(split[i]) > tt.maxSize {
					t.Errorf("Comment %d exceeds maxSize! Length: %d > %d", i, len(split[i]), tt.maxSize)
				}

			}
		})
	}
}

func TestAutomergeCommitMsg(t *testing.T) {
	tests := []struct {
		name    string
		pullNum int
		want    string
	}{
		{
			name:    "Atlantis PR commit message should include PR number",
			pullNum: 123,
			want:    "[Atlantis] Automatically merging after successful apply: PR #123",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := common.AutomergeCommitMsg(tt.pullNum); got != tt.want {
				t.Errorf("AutomergeCommitMsg() = %v, want %v", got, tt.want)
			}
		})
	}
}
