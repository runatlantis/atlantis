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

// RepoWhitelistChecker implements checking if repos are whitelisted to be used with
// this Atlantis.
type RepoWhitelistChecker struct {
	rules []string
}

// NewRepoWhitelistChecker constructs a new checker and validates that the
// whitelist isn't malformed.
func NewRepoWhitelistChecker(whitelist string) (*RepoWhitelistChecker, error) {
	rules := strings.Split(whitelist, ",")
	for _, rule := range rules {
		if strings.Contains(rule, "://") {
			return nil, fmt.Errorf("whitelist %q contained ://", rule)
		}
	}
	return &RepoWhitelistChecker{
		rules: rules,
	}, nil
}

// IsWhitelisted returns true if this repo is in our whitelist and false
// otherwise.
func (r *RepoWhitelistChecker) IsWhitelisted(repoFullName string, vcsHostname string) bool {
	candidate := fmt.Sprintf("%s/%s", vcsHostname, repoFullName)
	for _, rule := range r.rules {
		if r.matchesRule(rule, candidate) {
			return true
		}
	}
	return false
}

func (r *RepoWhitelistChecker) matchesRule(rule string, candidate string) bool {
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

	// Finally we can use the wildcard. Substring both so they're comparing before the wildcard. Example:
	// candidate: abcd
	// rule: abc*
	// substr(candidate): abc
	// substr(rule): abc
	return candidate[:wildcardIdx] == rule[:wildcardIdx]
}
