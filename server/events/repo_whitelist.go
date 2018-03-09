package events

import (
	"fmt"
	"strings"
)

const Wildcard = "*"

type RepoWhitelist struct {
	// Whitelist is a comma separated list of rules with wildcards '*' allowed.
	Whitelist string
}

// IsWhitelisted returns true if this repo is in our whitelist and false
// otherwise.
func (r *RepoWhitelist) IsWhitelisted(repoFullName string, vcsHostname string) bool {
	candidate := fmt.Sprintf("%s/%s", vcsHostname, repoFullName)
	rules := strings.Split(r.Whitelist, ",")
	for _, rule := range rules {
		if r.matchesRule(rule, candidate) {
			return true
		}
	}
	return false
}

func (r *RepoWhitelist) matchesRule(rule string, candidate string) bool {
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
