package main

import (
	"log"

	"github.com/xanzy/go-gitlab"
)

// This example shows how to create a client with username and password.
func basicAuthExample() {
	git, err := gitlab.NewBasicAuthClient(nil, "https://gitlab.company.com", "svanharmelen", "password")
	if err != nil {
		log.Fatal(err)
	}

	// List all projects
	projects, _, err := git.Projects.ListProjects(nil)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Found %d projects", len(projects))
}
