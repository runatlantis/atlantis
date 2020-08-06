package main

import (
	"log"

	"github.com/xanzy/go-gitlab"
)

func applicationsExample() {
	git := gitlab.NewClient(nil, "yourtokengoeshere")

	// Create an application
	opts := &gitlab.CreateApplicationOptions{
		Name:        gitlab.String("Travis"),
		RedirectURI: gitlab.String("http://example.org"),
		Scopes:      gitlab.String("api"),
	}
	created, _, err := git.Applications.CreateApplication(opts)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Last created application : %v", created)

	// List all applications
	applications, _, err := git.Applications.ListApplications(&gitlab.ListApplicationsOptions{})
	if err != nil {
		log.Fatal(err)
	}

	for _, app := range applications {
		log.Printf("Found app : %v", app)
	}

	// Delete an application
	resp, err := git.Applications.DeleteApplication(created.ID)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Status code response : %d", resp.StatusCode)
}
