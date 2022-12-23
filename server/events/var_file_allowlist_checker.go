package events

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

// VarFileAllowlistChecker implements checking if paths are allowlisted to be used with
// this Atlantis.
type VarFileAllowlistChecker struct {
	rules []string
}

// NewVarFileAllowlistChecker constructs a new checker and validates that the
// allowlist isn't malformed.
func NewVarFileAllowlistChecker(allowlist string) (*VarFileAllowlistChecker, error) {
	var rules []string
	paths := strings.Split(allowlist, ",")
	if paths[0] != "" {
		for _, path := range paths {
			absPath, err := filepath.Abs(path)
			if err != nil {
				return nil, errors.Wrap(err, fmt.Sprintf("converting allowlist %q to absolute path", path))
			}
			rules = append(rules, absPath)
		}
	}
	return &VarFileAllowlistChecker{
		rules: rules,
	}, nil
}

func (p *VarFileAllowlistChecker) Check(flags []string) error {
	for i, flag := range flags {
		var path string
		if i < len(flags)-1 && flag == "-var-file" {
			// Flags are in the format of []{"-var-file", "my-file.tfvars"}
			path = flags[i+1]
		} else {
			flagSplit := strings.Split(flag, "=")
			// Flags are in the format of []{"-var-file=my-file.tfvars"}
			if len(flagSplit) == 2 && flagSplit[0] == "-var-file" {
				path = flagSplit[1]
			}
		}

		if path != "" && !p.isAllowedPath(path) {
			return fmt.Errorf("var file path %s is not allowed by the current allowlist: [%s]",
				path, strings.Join(p.rules, ", "))
		}
	}
	return nil
}

func (p *VarFileAllowlistChecker) isAllowedPath(path string) bool {
	path = filepath.Clean(path)

	// If the path is within the repo directory, return true without checking the rules.
	if !filepath.IsAbs(path) {
		if !strings.HasPrefix(path, "..") && !strings.HasPrefix(path, "~") {
			return true
		}
	}

	// Check the path against the rules.
	for _, rule := range p.rules {
		rel, err := filepath.Rel(rule, path)
		if err == nil && !strings.HasPrefix(rel, "..") {
			return true
		}
	}

	return false
}
