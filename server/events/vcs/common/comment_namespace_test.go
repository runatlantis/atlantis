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
	Assert(t, namespace.Owns(body, "plan"), "disabled namespace should preserve legacy matching")
	Equals(t, 100, namespace.ContentLimit(100, "plan"))
}

func TestCommentNamespaceMarkerNormalizesCommand(t *testing.T) {
	namespace := common.NewCommentNamespace("prod-us-east-1")

	Equals(t, namespace.Marker("approve_policies"), namespace.Marker("Approve Policies"))
	Equals(t, namespace.Marker("plan"), namespace.Marker("Plan"))
}

func TestCommentNamespaceTagsAndMatches(t *testing.T) {
	namespace := common.NewCommentNamespace("prod-us-east-1")
	body := "Ran Plan for 2 projects:"
	tagged := namespace.Tag(body, "plan")

	Assert(t, namespace.Enabled(), "expected namespace to be enabled")
	Assert(t, strings.HasPrefix(tagged, body+"\n"), "expected marker on a new final line")
	Assert(t, namespace.Owns(tagged, "Plan"), "expected title-form command to match canonical marker")
	Equals(t, tagged, namespace.Tag(tagged, "plan"))
	Assert(t, !common.NewCommentNamespace("prod-us-west-2").Owns(tagged, "plan"), "wrong namespace must not own comment")
	Assert(t, !namespace.Owns(tagged, "apply"), "wrong command must not own comment")
	Assert(t, !namespace.Owns(body, "plan"), "unmarked comment must not be claimed")
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
