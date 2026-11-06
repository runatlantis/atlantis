// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package metrics

import tally "github.com/uber-go/tally/v4"

// PullRequestNumberTag is the tag key that holds the pull request number.
const PullRequestNumberTag = "pr_number"

// WithoutTags returns a scope that removes the given tag keys from every
// Tagged call made on it or on any scope derived from it.
func WithoutTags(scope tally.Scope, keys ...string) tally.Scope {
	if len(keys) == 0 {
		return scope
	}
	drop := make(map[string]struct{}, len(keys))
	for _, k := range keys {
		drop[k] = struct{}{}
	}
	return &tagFilterScope{Scope: scope, drop: drop}
}

type tagFilterScope struct {
	tally.Scope
	drop map[string]struct{}
}

func (s *tagFilterScope) Tagged(tags map[string]string) tally.Scope {
	kept := make(map[string]string, len(tags))
	for k, v := range tags {
		if _, ok := s.drop[k]; !ok {
			kept[k] = v
		}
	}
	return &tagFilterScope{Scope: s.Scope.Tagged(kept), drop: s.drop}
}

func (s *tagFilterScope) SubScope(name string) tally.Scope {
	return &tagFilterScope{Scope: s.Scope.SubScope(name), drop: s.drop}
}
