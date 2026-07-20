// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"slices"
	"strings"

	"github.com/runatlantis/atlantis/server/events/models"
)

// supportedMergeMethods lists, per VCS host, the normalised merge methods that
// host is able to perform. Each client's MergePull translates these onto the
// provider's own merge strategy (see the per-client switch statements). A host
// absent from this map does not support choosing a merge method.
var supportedMergeMethods = map[models.VCSHostType][]models.MergeMethod{
	models.Github:          {models.MergeMethodMerge, models.MergeMethodRebase, models.MergeMethodSquash},
	models.Gitea:           {models.MergeMethodMerge, models.MergeMethodRebase, models.MergeMethodSquash, models.MergeMethodFastForward},
	models.Gitlab:          {models.MergeMethodMerge, models.MergeMethodSquash},
	models.BitbucketCloud:  {models.MergeMethodMerge, models.MergeMethodSquash, models.MergeMethodFastForward},
	models.BitbucketServer: {models.MergeMethodMerge, models.MergeMethodRebase, models.MergeMethodSquash, models.MergeMethodFastForward},
	models.AzureDevops:     {models.MergeMethodMerge, models.MergeMethodRebase, models.MergeMethodSquash},
}

// SupportedMergeMethods returns the merge methods that host is able to perform.
// It returns nil if the host does not support selecting a merge method.
func SupportedMergeMethods(host models.VCSHostType) []models.MergeMethod {
	return supportedMergeMethods[host]
}

// IsMergeMethodValidForHost reports whether host is able to perform method.
func IsMergeMethodValidForHost(host models.VCSHostType, method models.MergeMethod) bool {
	return slices.Contains(supportedMergeMethods[host], method)
}

// AllSupportedMergeMethods returns the sorted union of every merge method any
// VCS host can perform. It is used to validate the server-side
// --automerge-method flag, which is set before the VCS host of an individual
// pull request is known; the host-specific check happens later in the comment
// parser and each client's MergePull.
func AllSupportedMergeMethods() []models.MergeMethod {
	seen := make(map[models.MergeMethod]struct{})
	for _, methods := range supportedMergeMethods {
		for _, method := range methods {
			seen[method] = struct{}{}
		}
	}
	all := make([]models.MergeMethod, 0, len(seen))
	for method := range seen {
		all = append(all, method)
	}
	slices.Sort(all)
	return all
}

// FormatMergeMethods renders methods as a comma-separated list for use in flag
// help and error messages.
func FormatMergeMethods(methods []models.MergeMethod) string {
	names := make([]string, len(methods))
	for i, method := range methods {
		names[i] = method.String()
	}
	return strings.Join(names, ", ")
}
