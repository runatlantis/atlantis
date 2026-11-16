// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package events

import (
	"fmt"
	"strings"
)

// Wildcard matches 0-n of all characters except commas.
const Wildcard = "*"

// RepoAllowlistChecker implements checking if repos are allowlisted to be used with
// this Atlantis.
type RepoAllowlistChecker struct {
	includeRules []string
	omitRules    []string
}

// NewRepoAllowlistChecker constructs a new checker and validates that the
// allowlist isn't malformed.
func NewRepoAllowlistChecker(allowlist string) (*RepoAllowlistChecker, error) {
	includeRules := make([]string, 0)
	omitRules := make([]string, 0)
	for rule := range strings.SplitSeq(allowlist, ",") {
		if strings.Contains(rule, "://") {
			return nil, fmt.Errorf("allowlist %q contained ://", rule)
		}
		if len(rule) > 1 && rule[0] == '!' {
			omitRules = append(omitRules, rule[1:])
		} else {
			includeRules = append(includeRules, rule)
		}
	}
	return &RepoAllowlistChecker{
		includeRules: includeRules,
		omitRules:    omitRules,
	}, nil
}

// IsAllowlisted returns true if this repo is in our allowlist and false
// otherwise.
func (r *RepoAllowlistChecker) IsAllowlisted(repoFullName string, vcsHostname string) bool {
	candidate := fmt.Sprintf("%s/%s", vcsHostname, repoFullName)
	shouldInclude := r.matchesAtLeastOneRule(r.includeRules, candidate)
	shouldOmit := r.matchesAtLeastOneRule(r.omitRules, candidate)
	return shouldInclude && !shouldOmit
}

func (r *RepoAllowlistChecker) matchesAtLeastOneRule(rules []string, candidate string) bool {
	for _, rule := range rules {
		if r.matchesRule(rule, candidate) {
			return true
		}
	}
	return false
}

func (r *RepoAllowlistChecker) matchesRule(rule string, candidate string) bool {
	// Case insensitive compare.
	return globMatch(strings.ToLower(rule), strings.ToLower(candidate))
}

// globMatch reports whether candidate matches pattern, where Wildcard stands
// for a run of zero or more characters. The match is anchored at both ends, so
// every literal part of the pattern must appear, in order, including the part
// before the first wildcard.
//
// Anchoring the prefix matters: a rule such as "github.com/myorg/*-prod" must
// not admit "github.com/other-org/anything-prod". Matching only the text after
// the wildcard would allow any repository, in any organisation, whose name
// happens to end the right way.
func globMatch(pattern string, candidate string) bool {
	var (
		p, c          int
		starPat       = -1
		starCandidate int
	)
	for c < len(candidate) {
		switch {
		case p < len(pattern) && pattern[p] == candidate[c]:
			p++
			c++
		case p < len(pattern) && pattern[p] == Wildcard[0]:
			// Record the wildcard position so we can backtrack, and start by
			// matching it against the empty string.
			starPat = p
			starCandidate = c
			p++
		case starPat != -1:
			// Backtrack: let the last wildcard consume one more character.
			p = starPat + 1
			starCandidate++
			c = starCandidate
		default:
			return false
		}
	}
	// Any pattern left over must be wildcards, which can match the empty string.
	for p < len(pattern) && pattern[p] == Wildcard[0] {
		p++
	}
	return p == len(pattern)
}
