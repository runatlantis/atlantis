package models

import (
	"github.com/hashicorp/go-version"
)

type policySetSource string

const (
	LocalPolicySet policySetSource = "Local"
)

type PolicySets struct {
	Version    *version.Version
	PolicySets []PolicySet
}

type PolicySet struct {
	Path   string
	Source policySetSource
	Name   string
	Owners []string
}
