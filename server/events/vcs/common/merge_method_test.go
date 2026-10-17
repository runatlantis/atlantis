// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package common_test

import (
	"slices"
	"testing"

	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs/common"
	. "github.com/runatlantis/atlantis/testing"
)

func TestSupportedMergeMethods(t *testing.T) {
	// Every host Atlantis supports advertises at least one merge method, and
	// every advertised method is one of the normalised values.
	known := map[models.MergeMethod]bool{
		models.MergeMethodMerge:       true,
		models.MergeMethodRebase:      true,
		models.MergeMethodSquash:      true,
		models.MergeMethodFastForward: true,
	}
	for _, host := range []models.VCSHostType{
		models.Github,
		models.Gitea,
		models.Gitlab,
		models.BitbucketCloud,
		models.BitbucketServer,
		models.AzureDevops,
	} {
		methods := common.SupportedMergeMethods(host)
		Assert(t, len(methods) > 0, "expected %s to support at least one merge method", host)
		for _, method := range methods {
			Assert(t, known[method], "%s advertises unknown merge method %q", host, method)
		}
	}
}

func TestIsMergeMethodValidForHost(t *testing.T) {
	cases := []struct {
		host   models.VCSHostType
		method models.MergeMethod
		valid  bool
	}{
		{models.Github, models.MergeMethodSquash, true},
		{models.Github, models.MergeMethodFastForward, false},
		{models.Gitea, models.MergeMethodFastForward, true},
		{models.Gitea, models.MergeMethodRebase, true},
		{models.Gitlab, models.MergeMethodSquash, true},
		{models.Gitlab, models.MergeMethodRebase, false},
		{models.BitbucketCloud, models.MergeMethodFastForward, true},
		{models.BitbucketCloud, models.MergeMethodRebase, false},
		{models.BitbucketServer, models.MergeMethodFastForward, true},
		{models.AzureDevops, models.MergeMethodRebase, true},
		{models.AzureDevops, models.MergeMethodFastForward, false},
		{models.Gitea, "", false},
	}
	for _, c := range cases {
		Equals(t, c.valid, common.IsMergeMethodValidForHost(c.host, c.method))
	}
}

func TestAllSupportedMergeMethods(t *testing.T) {
	all := common.AllSupportedMergeMethods()

	// The union is sorted and de-duplicated even though several methods are
	// supported by more than one host.
	Assert(t, sortedAndUnique(all), "expected the union to be sorted and de-duplicated, got %v", all)

	// Every host's methods appear in the union.
	for _, host := range []models.VCSHostType{models.Github, models.Gitea, models.AzureDevops} {
		for _, method := range common.SupportedMergeMethods(host) {
			Assert(t, slices.Contains(all, method), "expected the union to contain %q", method)
		}
	}
}

func TestFormatMergeMethods(t *testing.T) {
	Equals(t, "merge, squash", common.FormatMergeMethods(common.SupportedMergeMethods(models.Gitlab)))
}

func sortedAndUnique(s []models.MergeMethod) bool {
	for i := 1; i < len(s); i++ {
		if s[i-1] >= s[i] {
			return false
		}
	}
	return true
}
