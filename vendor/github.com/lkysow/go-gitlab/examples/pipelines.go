package main

import (
	"log"

	"github.com/xanzy/go-gitlab"
)

func pipelineExample() {
	git := gitlab.NewClient(nil, "yourtokengoeshere")
	git.SetBaseURL("https://gitlab.com/api/v4")

	pipelines, _, err := git.Pipelines.ListProjectPipelines(2743054)
	if err != nil {
		log.Fatal(err)
	}

	for _, pipeline := range pipelines {
		log.Printf("Found pipeline: %v", pipeline)
	}
}
