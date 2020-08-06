package main

import (
	"log"

	"github.com/xanzy/go-gitlab"
)

func labelExample() {
	git := gitlab.NewClient(nil, "yourtokengoeshere")

	// Create new label
	l := &gitlab.CreateLabelOptions{
		Name:  gitlab.String("My Label"),
		Color: gitlab.String("#11FF22"),
	}
	label, _, err := git.Labels.CreateLabel("myname/myproject", l)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Created label: %s\nWith color: %s\n", label.Name, label.Color)

	// List all labels
	labels, _, err := git.Labels.ListLabels("myname/myproject", nil)
	if err != nil {
		log.Fatal(err)
	}

	for _, label := range labels {
		log.Printf("Found label: %s", label.Name)
	}
}
