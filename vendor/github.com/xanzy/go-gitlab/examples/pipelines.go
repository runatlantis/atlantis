package main

import (
	"log"
	"time"

	"github.com/xanzy/go-gitlab"
)

func pipelineExample() {
	git := gitlab.NewClient(nil, "yourtokengoeshere")
	git.SetBaseURL("https://gitlab.com/api/v4")

	opt := &gitlab.ListProjectPipelinesOptions{
		Scope:         gitlab.String("branches"),
		Status:        gitlab.BuildState(gitlab.Running),
		Ref:           gitlab.String("master"),
		YamlErrors:    gitlab.Bool(true),
		Name:          gitlab.String("name"),
		Username:      gitlab.String("username"),
		UpdatedAfter:  gitlab.Time(time.Now().Add(-24 * 365 * time.Hour)),
		UpdatedBefore: gitlab.Time(time.Now().Add(-7 * 24 * time.Hour)),
		OrderBy:       gitlab.String("status"),
		Sort:          gitlab.String("asc"),
	}

	pipelines, _, err := git.Pipelines.ListProjectPipelines(2743054, opt)
	if err != nil {
		log.Fatal(err)
	}

	for _, pipeline := range pipelines {
		log.Printf("Found pipeline: %v", pipeline)
	}
}
