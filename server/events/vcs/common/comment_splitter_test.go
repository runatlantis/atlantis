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
	"strings"
	"testing"

	"github.com/runatlantis/atlantis/server/events/vcs/common"

	. "github.com/runatlantis/atlantis/testing"
)

// If under the maximum number of chars, we shouldn't split the comments.
func TestSplitComment_UnderMax(t *testing.T) {
	comment := "comment under max size"
	split := common.SplitComment(comment, len(comment)+1, "sepEnd", "sepStart")
	Equals(t, []string{comment}, split)
}

// If the comment needs to be split into 2 we should do the split and add the
// separators properly.
func TestSplitComment_TwoComments(t *testing.T) {
	comment := strings.Repeat("a", 1000)
	sepEnd := "-sepEnd"
	sepStart := "-sepStart"
	split := common.SplitComment(comment, len(comment)-1, sepEnd, sepStart)

	expCommentLen := len(comment) - len(sepEnd) - len(sepStart) - 1
	expFirstComment := comment[:expCommentLen]
	expSecondComment := comment[expCommentLen:]
	Equals(t, 2, len(split))
	Equals(t, expFirstComment+sepEnd, split[0])
	Equals(t, sepStart+expSecondComment, split[1])
}

// If the comment needs to be split into 4 we should do the split and add the
// separators properly.
func TestSplitComment_FourComments(t *testing.T) {
	comment := strings.Repeat("a", 1000)
	sepEnd := "-sepEnd"
	sepStart := "-sepStart"
	max := (len(comment) / 4) + len(sepEnd) + len(sepStart)
	split := common.SplitComment(comment, max, sepEnd, sepStart)

	expMax := len(comment) / 4
	Equals(t, []string{
		comment[:expMax] + sepEnd,
		sepStart + comment[expMax:expMax*2] + sepEnd,
		sepStart + comment[expMax*2:expMax*3] + sepEnd,
		sepStart + comment[expMax*3:]}, split)
}
