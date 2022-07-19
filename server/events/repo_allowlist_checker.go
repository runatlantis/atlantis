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
	rules []string
}

// NewRepoAllowlistChecker constructs a new checker and validates that the
// allowlist isn't malformed.
func NewRepoAllowlistChecker(allowlist string) (*RepoAllowlistChecker, error) {
	rules := strings.Split(allowlist, ",")
	for _, rule := range rules {
		if strings.Contains(rule, "://") {
			return nil, fmt.Errorf("allowlist %q contained ://", rule)
		}
	}
	return &RepoAllowlistChecker{
		rules: rules,
	}, nil
}

// IsAllowlisted returns true if this repo is in our allowlist and false
// otherwise.
func (r *RepoAllowlistChecker) IsAllowlisted(repoFullName string, vcsHostname string) bool {
	candidate := fmt.Sprintf("%s/%s", vcsHostname, repoFullName)
	for _, rule := range r.rules {
		if r.matchesRule(rule, candidate) {
			return true
		}
	}
	return false
}

func (r *RepoAllowlistChecker) matchesRule(rule string, candidate string) bool {
	// Case insensitive compare.
	rule = strings.ToLower(rule)
	candidate = strings.ToLower(candidate)

	wildcardIdx := strings.Index(rule, Wildcard)
	if wildcardIdx == -1 {
		// No wildcard so can do a straight up match.
		return candidate == rule
	}

	// If the candidate length is less than where we found the wildcard
	// then it can't be equal. For example:
	//   rule: abc*
	//   candidate: ab
	if len(candidate) < wildcardIdx {
		return false
	}

	// If wildcard is not the last character, substring both to compare what is after the wildcard.  Example:
	// candidate: repo-abc
	// rule: *-abc
	// substr(candidate): -abc
	// substr(rule): -abc
	if wildcardIdx != len(rule)-1 {
		// If the rule substring after wildcard does not exist in the candidate, then it is not a match.
		idx := strings.LastIndex(candidate, rule[wildcardIdx+1:])
		if idx == -1 {
			return false
		}
		return candidate[idx:] == rule[wildcardIdx+1:]
	}

	// If wildcard is last character, substring both so they're comparing before the wildcard. Example:
	// candidate: abcd
	// rule: abc*
	// substr(candidate): abc
	// substr(rule): abc
	return candidate[:wildcardIdx] == rule[:wildcardIdx]
}
