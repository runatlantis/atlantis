package models

import (
	paths "path"
	"time"
)

type Repo struct {
	FullName string
	Owner    string
	Name     string
	SSHURL   string
}

type PullRequest struct {
	Num        int
	HeadCommit string
	BaseCommit string
	URL        string
	Branch     string
	Author     string
}

type User struct {
	Username string
}

type ProjectLock struct {
	Project Project
	Pull    PullRequest
	User    User
	Env     string
	Time    time.Time
}

// Project represents a Terraform project.
// Since there may be multiple Terraform projects in a single repo we also include Path
type Project struct {
	RepoFullName string
	// Path to project root in the repo.
	// If "." then project is at root.
	// Never ends in "/".
	Path string
}

type Plan struct {
	Project Project
	// LocalPath is the path to the plan on disk
	LocalPath string
}

func NewProject(repoFullName string, path string) Project {
	path = paths.Clean(path)
	if path == "/" {
		path = "."
	}
	return Project{
		RepoFullName: repoFullName,
		Path:         path,
	}
}
