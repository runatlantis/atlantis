package main

import (
	"log"

	"github.com/xanzy/go-gitlab"
)

func repositoryFileExample() {
	git := gitlab.NewClient(nil, "yourtokengoeshere")

	// Create a new repository file
	cf := &gitlab.CreateFileOptions{
		Branch:        gitlab.String("master"),
		Content:       gitlab.String("My file contents"),
		CommitMessage: gitlab.String("Adding a test file"),
	}
	file, _, err := git.RepositoryFiles.CreateFile("myname/myproject", "file.go", cf)
	if err != nil {
		log.Fatal(err)
	}

	// Update a repository file
	uf := &gitlab.UpdateFileOptions{
		Branch:        gitlab.String("master"),
		Content:       gitlab.String("My file content"),
		CommitMessage: gitlab.String("Fixing typo"),
	}
	_, _, err = git.RepositoryFiles.UpdateFile("myname/myproject", file.FilePath, uf)
	if err != nil {
		log.Fatal(err)
	}

	gf := &gitlab.GetFileOptions{
		Ref: gitlab.String("master"),
	}
	f, _, err := git.RepositoryFiles.GetFile("myname/myproject", file.FilePath, gf)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("File contains: %s", f.Content)
}
