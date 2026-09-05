// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package common_test

import (
	"strings"
	"testing"

	"github.com/runatlantis/atlantis/server/events/vcs/common"
	. "github.com/runatlantis/atlantis/testing"
)

func TestCommentNamespaceDisabled(t *testing.T) {
	namespace := common.NewCommentNamespace("")
	body := "Ran Plan for 2 projects:"

	Assert(t, !namespace.Enabled(), "expected empty namespace to be disabled")
	Equals(t, "", namespace.Marker("plan"))
	Equals(t, body, namespace.Tag(body, "plan"))
	Assert(t, namespace.Owns(body), "disabled namespace should impose no ownership check")
	Assert(t, namespace.MatchesCommand(body, "plan"), "disabled namespace should preserve legacy command matching")
	Assert(t, !namespace.MatchesCommand(body, "apply"), "disabled namespace should reject a different command")
	Equals(t, 100, namespace.ContentLimit(100, "plan"))
}

func TestCommentNamespaceMarkerNormalizesCommand(t *testing.T) {
	namespace := common.NewCommentNamespace("prod-us-east-1")

	Equals(t, namespace.Marker("approve_policies"), namespace.Marker("Approve Policies"))
	Equals(t, namespace.Marker("plan"), namespace.Marker("Plan"))
}

func TestCommentNamespaceTagsOwnsAndMatches(t *testing.T) {
	namespace := common.NewCommentNamespace("prod-us-east-1")
	body := "Ran Plan for 2 projects:"
	tagged := namespace.Tag(body, "plan")

	Assert(t, namespace.Enabled(), "expected namespace to be enabled")
	Assert(t, strings.HasPrefix(tagged, body+"\n"), "expected marker on a new final line")
	Assert(t, namespace.Owns(tagged), "expected namespace to own tagged comment")
	Assert(t, namespace.MatchesCommand(tagged, "Plan"), "expected title-form command to match canonical marker")
	Assert(t, !namespace.MatchesCommand(tagged, "apply"), "non-empty marker command should be authoritative")
	Equals(t, tagged, namespace.Tag(tagged, "plan"))
	Assert(t, !common.NewCommentNamespace("prod-us-west-2").Owns(tagged), "wrong namespace must not own comment")
	Assert(t, !namespace.Owns(body), "unmarked comment must not be claimed")
}

func TestCommentNamespaceEmptyCommandFallsBackToFirstLine(t *testing.T) {
	namespace := common.NewCommentNamespace("prod-us-east-1")
	tagged := namespace.Tag("Ran Plan for 2 projects:", "")

	Assert(t, namespace.Owns(tagged), "expected namespace to own empty-command marker")
	Assert(t, namespace.MatchesCommand(tagged, "plan"), "expected first-line command match")
	Assert(t, !namespace.MatchesCommand(tagged, "apply"), "expected first-line command mismatch")
}

func TestCommentNamespaceMarkerCommandOverridesFirstLine(t *testing.T) {
	namespace := common.NewCommentNamespace("prod-us-east-1")
	tagged := namespace.Tag("Ran Apply for 2 projects:", "plan")

	Assert(t, namespace.MatchesCommand(tagged, "plan"), "expected marker command match")
	Assert(t, !namespace.MatchesCommand(tagged, "apply"), "expected marker to override first-line prose")
}

func TestCommentNamespaceTagRequiresExactMarker(t *testing.T) {
	namespace := common.NewCommentNamespace("prod-us-east-1")
	applyTagged := namespace.Tag("Ran Apply for 2 projects:", "apply")
	planTagged := namespace.Tag(applyTagged, "plan")

	Assert(t, planTagged != applyTagged, "different command marker must not make tagging idempotent")
	Assert(t, strings.HasSuffix(planTagged, namespace.Marker("plan")), "expected requested marker to be final")
}

func TestCommentNamespaceTagRejectsInlineMarkerSuffix(t *testing.T) {
	namespace := common.NewCommentNamespace("prod-us-east-1")
	marker := namespace.Marker("plan")
	body := "command output ending with " + marker
	tagged := namespace.Tag(body, "plan")

	Assert(t, tagged != body, "inline marker suffix must not make tagging idempotent")
	Assert(t, strings.HasSuffix(tagged, "\n"+marker), "expected marker on its own final line")
	Assert(t, namespace.Owns(tagged), "tagged body must be recognized as owned")
}

func TestCommentNamespaceTagsBodyEndingInNewline(t *testing.T) {
	namespace := common.NewCommentNamespace("prod")
	body := "Ran Plan\n"

	Equals(t, body+namespace.Marker("plan"), namespace.Tag(body, "plan"))
}

func TestCommentNamespaceContentLimit(t *testing.T) {
	namespace := common.NewCommentNamespace("prod")
	marker := namespace.Marker("plan")

	Equals(t, 100-len(marker)-1, namespace.ContentLimit(100, "plan"))
}
