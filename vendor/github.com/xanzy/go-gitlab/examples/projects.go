package main

import (
	"log"

	"github.com/xanzy/go-gitlab"
)

func projectExample() {
	git := gitlab.NewClient(nil, "yourtokengoeshere")

	// Create new project
	p := &gitlab.CreateProjectOptions{
		Name:                 gitlab.String("My Project"),
		Description:          gitlab.String("Just a test project to play with"),
		MergeRequestsEnabled: gitlab.Bool(true),
		SnippetsEnabled:      gitlab.Bool(true),
		Visibility:           gitlab.Visibility(gitlab.PublicVisibility),
	}
	project, _, err := git.Projects.CreateProject(p)
	if err != nil {
		log.Fatal(err)
	}

	// Add a new snippet
	s := &gitlab.CreateProjectSnippetOptions{
		Title:      gitlab.String("Dummy Snippet"),
		FileName:   gitlab.String("snippet.go"),
		Code:       gitlab.String("package main...."),
		Visibility: gitlab.Visibility(gitlab.PublicVisibility),
	}
	_, _, err = git.ProjectSnippets.CreateSnippet(project.ID, s)
	if err != nil {
		log.Fatal(err)
	}

	// List all project snippets
	snippets, _, err := git.ProjectSnippets.ListSnippets(project.PathWithNamespace, &gitlab.ListProjectSnippetsOptions{})
	if err != nil {
		log.Fatal(err)
	}

	for _, snippet := range snippets {
		log.Printf("Found snippet: %s", snippet.Title)
	}
}
