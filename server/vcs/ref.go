package vcs

import "fmt"

type RefType string

const (
	Unknown   RefType = "unknown"
	BranchRef RefType = "branch"
	TagRef    RefType = "tag"
)

type Ref struct {
	Type RefType
	Name string
}

func FromGithubRefType(t string) (RefType, error) {
	switch t {
	case "heads":
		return BranchRef, nil
	case "tags":
		return TagRef, nil
	}

	return Unknown, fmt.Errorf("unknown ref type %s", t)
}