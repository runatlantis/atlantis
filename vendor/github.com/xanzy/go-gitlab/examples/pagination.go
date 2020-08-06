package main

import (
	"fmt"
	"log"

	"github.com/xanzy/go-gitlab"
)

func pagination() {
	git := gitlab.NewClient(nil, "yourtokengoeshere")

	opt := &gitlab.ListProjectsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 10,
			Page:    1,
		},
	}

	for {
		// Get the first page with projects.
		ps, resp, err := git.Projects.ListProjects(opt)
		if err != nil {
			log.Fatal(err)
		}

		// List all the projects we've found so far.
		for _, p := range ps {
			fmt.Printf("Found project: %s", p.Name)
		}

		// Exit the loop when we've seen all pages.
		if resp.CurrentPage >= resp.TotalPages {
			break
		}

		// Update the page number to get the next page.
		opt.Page = resp.NextPage
	}
}
