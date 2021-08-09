package events

import ( 
	"github.com/runatlantis/atlantis/server/events/models" 
)

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_job_url_generator.go JobURLGenerator

type JobURLGenerator interface {
	GenerateJobURL(pull models.PullRequest) string
}
