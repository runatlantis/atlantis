# go-gitlab

A GitLab API client enabling Go programs to interact with GitLab in a simple and uniform way

**Documentation:** [![GoDoc](https://godoc.org/github.com/xanzy/go-gitlab?status.svg)](https://godoc.org/github.com/xanzy/go-gitlab)
**Build Status:** [![Build Status](https://travis-ci.org/xanzy/go-gitlab.svg?branch=master)](https://travis-ci.org/xanzy/go-gitlab)

## NOTE

Release v0.6.0 (released on 25-08-2017) no longer supports the older V3 Gitlab API. If
you need V3 support, please use the `f-api-v3` branch. This release contains some backwards
incompatible changes that were needed to fully support the V4 Gitlab API.

## Coverage

This API client package covers **100%** of the existing GitLab API calls! So this
includes all calls to the following services:

- [x] Branches
- [x] Commits
- [x] Deploy Keys
- [x] Environments
- [x] Groups
- [x] Issues
- [x] Labels
- [x] Merge Requests
- [x] Milestones
- [x] Namespaces
- [x] Notes (comments)
- [x] Pipelines
- [x] Project Snippets
- [x] Projects (including setting Webhooks)
- [x] Repositories
- [x] Repository Files
- [x] Services
- [x] Session
- [x] Settings
- [x] System Hooks
- [x] Users
- [x] Version
- [x] Wikis

## Usage

```go
import "github.com/xanzy/go-gitlab"
```

Construct a new GitLab client, then use the various services on the client to
access different parts of the GitLab API. For example, to list all
users:

```go
git := gitlab.NewClient(nil, "yourtokengoeshere")
//git.SetBaseURL("https://git.mydomain.com/api/v3")
users, _, err := git.Users.ListUsers()
```

Some API methods have optional parameters that can be passed. For example,
to list all projects for user "svanharmelen":

```go
git := gitlab.NewClient(nil)
opt := &ListProjectsOptions{Search: gitlab.String("svanharmelen")}
projects, _, err := git.Projects.ListProjects(opt)
```

### Examples

The [examples](https://github.com/xanzy/go-gitlab/tree/master/examples) directory
contains a couple for clear examples, of which one is partially listed here as well:

```go
package main

import (
	"log"

	"github.com/xanzy/go-gitlab"
)

func main() {
	git := gitlab.NewClient(nil, "yourtokengoeshere")

	// Create new project
	p := &gitlab.CreateProjectOptions{
		Name:                 gitlab.String("My Project"),
		Description:          gitlab.String("Just a test project to play with"),
		MergeRequestsEnabled: gitlab.Bool(true),
		SnippetsEnabled:      gitlab.Bool(true),
		Visibility:           gitlab.VisibilityLevel(gitlab.PublicVisibility),
	}
	project, _, err := git.Projects.CreateProject(p)
	if err != nil {
		log.Fatal(err)
	}

	// Add a new snippet
	s := &gitlab.CreateSnippetOptions{
		Title:           gitlab.String("Dummy Snippet"),
		FileName:        gitlab.String("snippet.go"),
		Code:            gitlab.String("package main...."),
		Visibility:      gitlab.VisibilityLevel(gitlab.PublicVisibility),
	}
	_, _, err = git.ProjectSnippets.CreateSnippet(project.ID, s)
	if err != nil {
		log.Fatal(err)
	}
}

```

For complete usage of go-gitlab, see the full [package docs](https://godoc.org/github.com/xanzy/go-gitlab).

## ToDo

- The biggest thing this package still needs is tests :disappointed:

## Issues

- If you have an issue: report it on the [issue tracker](https://github.com/xanzy/go-gitlab/issues)

## Author

Sander van Harmelen (<sander@xanzy.io>)

## License

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at <http://www.apache.org/licenses/LICENSE-2.0>
