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
				SepStart:         fmt.Sprintf("<details><summary>Show Output</summary>\n\n```diff\n%s", baseStart),
				TruncationHeader: "<details><summary>Show Output</summary>\n\n```diff\n<!-- TRUNCATED -->",
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
			maxSize:       70,
			maxComments:   0,
			command:       "plan",
			expectedCount: 11,
			expectedComments: []string{
				"This is a long comment that will <!-- END -->",
				"<!-- START plan -->be split. This is additional content <!-- END -->",
				"<!-- START plan -->to make the comment longer so it will <!-- END -->",
				"<!-- START plan -->be split. This is additional content <!-- END -->",
				"<!-- START plan -->to make the comment longer so it will <!-- END -->",
				"<!-- START plan -->be split. This is additional content <!-- END -->",
				"<!-- START plan -->to make the comment longer so it will <!-- END -->",
				"<!-- START plan -->be split. This is additional content <!-- END -->",
				"<!-- START plan -->to make the comment longer so it will <!-- END -->",
				"<!-- START plan -->be split. This is additional content <!-- END -->",
				"<!-- START plan -->to make the comment longer so it will be split. ",
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
				"Here's some code:\n```\nterraform plan\noutput here\n```\nAnd more text. This is additional content to make the comment <!-- END -->",
				"<!-- START plan -->longer so it will be split. This is additional content to make the comment longer so it will be split. This is additional content to make the comment longer so it will be split. ",
			},
		},
		{
			name:          "DetailsBlock - Splitting within details block",
			comment:       "<details><summary>Show Output</summary>\n\nSome details content here. " + strings.Repeat("This is additional content to make the comment longer so it will be split. ", 4) + "\n</details>",
			maxSize:       200,
			maxComments:   0,
			command:       "plan",
			expectedCount: 4,
			expectedComments: []string{
				"<details><summary>Show Output</summary>\n\nSome details content here. This is additional content to make the comment longer so it will be </details>\n<!-- END -->",
				"<details><summary>Show Output</summary>\n<!-- START plan -->split. This is additional content to make the comment longer so it will be split. This is additional content to make </details>\n<!-- END -->",
				"<details><summary>Show Output</summary>\n<!-- START plan -->the comment longer so it will be split. This is additional content to make the comment longer so it will be split. \n</details>\n<!-- END -->",
				"<details><summary>Show Output</summary>\n<!-- START plan --></details>",
			},
		},
		{
			name: "CodeInDetails - Splitting within code block inside details",
			comment: "<details><summary>Show Output</summary>\n\n```diff\n" +
				strings.Repeat("Line of terraform output\n", 20) +
				"```\n</details>",
			maxSize:       200,
			maxComments:   0,
			command:       "plan",
			expectedCount: 5,
			// Split inside code block within details; continuation reopens details + code block
			expectedComments: []string{
				"<details><summary>Show Output</summary>\n\n```diff\nLine of terraform output\nLine of terraform output\nLine of terraform output\nLine of terraform output\n```\n</details>\n<!-- END -->",
				"<details><summary>Show Output</summary>\n\n```diff\n<!-- START plan -->Line of terraform output\nLine of terraform output\nLine of terraform output\nLine of terraform output\n```\n</details>\n<!-- END -->",
				"<details><summary>Show Output</summary>\n\n```diff\n<!-- START plan -->Line of terraform output\nLine of terraform output\nLine of terraform output\nLine of terraform output\n```\n</details>\n<!-- END -->",
				"<details><summary>Show Output</summary>\n\n```diff\n<!-- START plan -->Line of terraform output\nLine of terraform output\nLine of terraform output\nLine of terraform output\n```\n</details>\n<!-- END -->",
				"<details><summary>Show Output</summary>\n\n```diff\n<!-- START plan -->Line of terraform output\nLine of terraform output\nLine of terraform output\nLine of terraform output\n```\n</details>",
			},
		},
		{
			name:          "InlineCode - Splitting within inline code block",
			comment:       "Here is text: `" + strings.Repeat("a", 150) + "` end.",
			maxSize:       100,
			maxComments:   0,
			command:       "plan",
			expectedCount: 4,
			expectedComments: []string{
				"Here is text: `" + strings.Repeat("a", 18) + "`\n<!-- END -->",
				"<!-- START plan -->`" + strings.Repeat("a", 66) + "`\n<!-- END -->",
				"<!-- START plan -->`" + strings.Repeat("a", 66) + "` <!-- END -->",
				"<!-- START plan -->end.",
			},
		},
		{
			name: "CodeInDetails_DeeplyNested - Splitting within nested structures inside details",
			comment: "<details><summary>Show Output</summary>\n\n```diff\n" +
				`  # aws_iam_role.task will be created
+   resource "aws_iam_role" "task" {
+       arn                   = (known after apply)
+       assume_role_policy    = jsonencode(
            {
+               Statement = [
+                   {
+                       Action    = "sts:AssumeRole"
+                       Effect    = "Allow"
+                       Principal = {
+                           Service = "ecs-tasks.amazonaws.com"
                        }
                    },
                ]
+               Version   = "2012-10-17"
            }
        )
+       create_date           = (known after apply)
    }
` + "\n```\n</details>",
			maxSize:       200,
			maxComments:   0,
			command:       "plan",
			expectedCount: 8,
			expectedComments: []string{
				"<details><summary>Show Output</summary>\n\n```diff\n  # aws_iam_role.task will be created\n```\n</details>\n<!-- END -->",
				"<details><summary>Show Output</summary>\n\n```diff\n<!-- START plan -->+   resource \"aws_iam_role\" \"task\" {\n+       arn                   = (known after apply)\n```\n</details>\n<!-- END -->",
				"<details><summary>Show Output</summary>\n\n```diff\n<!-- START plan -->+       assume_role_policy    = jsonencode(\n            {\n+               Statement = [\n```\n</details>\n<!-- END -->",
				"<details><summary>Show Output</summary>\n\n```diff\n<!-- START plan -->+                   {\n+                       Action    = \"sts:AssumeRole\"\n```\n</details>\n<!-- END -->",
				"<details><summary>Show Output</summary>\n\n```diff\n<!-- START plan -->+                       Effect    = \"Allow\"\n+                       Principal = {\n```\n</details>\n<!-- END -->",
				"<details><summary>Show Output</summary>\n\n```diff\n<!-- START plan -->+                           Service = \"ecs-tasks.amazonaws.com\"\n                        }\n```\n</details>\n<!-- END -->",
				"<details><summary>Show Output</summary>\n\n```diff\n<!-- START plan -->                    },\n                ]\n+               Version   = \"2012-10-17\"\n```\n</details>\n<!-- END -->",
				"<details><summary>Show Output</summary>\n\n```diff\n<!-- START plan -->            }\n        )\n+       create_date           = (known after apply)\n    }\n\n```\n</details>",
			},
		},
		{
			name:          "DetailsBlockWithAttributes - Splitting within details block with attributes",
			comment:       "<details open><summary>Show Output</summary>\n\nSome details content here. " + strings.Repeat("This is additional content to make the comment longer so it will be split. ", 4) + "\n</details>",
			maxSize:       200,
			maxComments:   0,
			command:       "plan",
			expectedCount: 4,
			expectedComments: []string{
				"<details open><summary>Show Output</summary>\n\nSome details content here. This is additional content to make the comment longer so it will be </details>\n<!-- END -->",
				"<details><summary>Show Output</summary>\n<!-- START plan -->split. This is additional content to make the comment longer so it will be split. This is additional content to make </details>\n<!-- END -->",
				"<details><summary>Show Output</summary>\n<!-- START plan -->the comment longer so it will be split. This is additional content to make the comment longer so it will be split. \n</details>\n<!-- END -->",
				"<details><summary>Show Output</summary>\n<!-- START plan --></details>",
			},
		},
		{
			name: "DetailsBlockWithAttributes - Splitting within details block with attributes -2",
			comment: strings.ReplaceAll(`# Some header 1
<details open><summary>Show Output 1a.</summary>
$$$diff
Code block within an OPEN details block 1.
$$$
</details>

<details><summary>Show Output 1b.</summary>
$$$diff
Code block within a CLOSED details block.
$$$
</details>

# Some header 2

<details open><summary>Show Output 2a.</summary>
$$$diff
Code block within an OPEN details block 2.
$$$
</details>

<details><summary>Show Output 2b.</summary>
$$$diff
Code block within a CLOSED details block 2.
$$$
</details>

# Some header 3

$some single line code block 1$
$some single line code block 2$
$some single line code block 3$
</details>`, "$", "`"),
			maxSize:       200,
			maxComments:   0,
			command:       "plan",
			expectedCount: 5,
			expectedComments: []string{
				strings.ReplaceAll(`# Some header 1
<details open><summary>Show Output 1a.</summary>
$$$diff
Code block within an OPEN details block 1.
$$$
</details>
<!-- END -->`, "$", "`"),
				strings.ReplaceAll(`<!-- START plan -->
<details><summary>Show Output 1b.</summary>
$$$diff
Code block within a CLOSED details block.
$$$
</details>

# Some header 2
<!-- END -->`, "$", "`"),
				strings.ReplaceAll(`<!-- START plan -->
<details open><summary>Show Output 2a.</summary>
$$$diff
Code block within an OPEN details block 2.
$$$
</details>
<!-- END -->`, "$", "`"),
				strings.ReplaceAll(`<!-- START plan -->
<details><summary>Show Output 2b.</summary>
$$$diff
Code block within a CLOSED details block 2.
$$$
</details>
<!-- END -->`, "$", "`"),
				strings.ReplaceAll(`<!-- START plan -->
# Some header 3

$some single line code block 1$
$some single line code block 2$
$some single line code block 3$
</details>`, "$", "`"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			split := common.SplitComment(logger, tt.comment, tt.maxSize, tt.maxComments, tt.command)

			if len(split) != tt.expectedCount {
				t.Errorf("Expected %d comments, got %d", tt.expectedCount, len(split))
				for i, s := range split {
					t.Logf("Got comment %d: %q", i, s)
				}
			}

			// Compare each comment with expected output
			for i, expected := range tt.expectedComments {
				if i < len(split) {
					Assert(t, split[i] == expected,
						"Comment %d mismatch:\nExpected: %s\nGot:      %s",
						i, expected, split[i])
				}
				if i < len(split) && len(split[i]) > tt.maxSize {
					t.Errorf("Comment %d exceeds maxSize: length %d > %d", i, len(split[i]), tt.maxSize)
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
