package github

import (
	"fmt"
)

type Commit struct {
	Ref    Ref
	Author User
}

type Ref struct {
	Name string
	Type string
}

func (r Ref) String() (string, error) {
	if r.Type == "branch" {
		return fmt.Sprintf("refs/heads/%s", r.Name), nil
	} else if r.Type == "tag" {
		return fmt.Sprintf("refs/tags/%s", r.Name), nil
	}
	return "", fmt.Errorf("unknown ref type: %s", r.Type)
}
