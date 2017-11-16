package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/lkysow/go-gitlab"
)

const secretHeader = "X-Gitlab-Token"

//go:generate pegomock generate --use-experimental-model-gen --package mocks -o mocks/mock_gitlab_request_parser.go GitlabRequestParser

// GitlabRequestParser parses and validates GitLab requests.
type GitlabRequestParser interface {
	// Validate validates that the request has a token header matching secret.
	// If the secret does not match it returns an error.
	// If secret is empty it does not check the token header.
	// It then parses the request as a gitlab object depending on the header
	// provided by GitLab identifying the webhook type. If the webhook type
	// is not recognized it will return nil but will not return an error.
	// Usage:
	//	event, err := GitlabRequestParser.Validate(r, secret)
	//	if err != nil {
	//		return
	//	}
	//	switch event := event.(type) {
	//	case gitlab.MergeCommentEvent:
	//		// handle
	//	case gitlab.MergeEvent:
	//		// handle
	//	default:
	//		// unsupported event
	//	}
	Validate(r *http.Request, secret []byte) (interface{}, error)
}

// DefaultGitlabRequestParser parses GitLab requests.
type DefaultGitlabRequestParser struct{}

// Validate returns the JSON payload of the request.
// If secret is not empty, it checks that the request was signed
// by secret and returns an error if it was not.
// If secret is empty, it does not check if the request was signed.
func (d *DefaultGitlabRequestParser) Validate(r *http.Request, secret []byte) (interface{}, error) {
	const mergeEventHeader = "Merge Request Hook"
	const noteEventHeader = "Note Hook"

	// Validate secret if specified.
	headerSecret := r.Header.Get(secretHeader)
	secretStr := string(secret)
	if len(secret) != 0 && headerSecret != secretStr {
		return nil, fmt.Errorf("header %s=%s did not match expected secret", secretHeader, headerSecret)
	}

	// Parse request into a gitlab object based on the object type specified
	// in the gitlabHeader.
	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	switch r.Header.Get(gitlabHeader) {
	case mergeEventHeader:
		var m gitlab.MergeEvent
		if err := json.Unmarshal(bytes, &m); err != nil {
			return nil, err
		}
		return m, nil
	case noteEventHeader:
		var m gitlab.MergeCommentEvent
		if err := json.Unmarshal(bytes, &m); err != nil {
			return nil, err
		}
		return m, nil
	}
	return nil, nil
}
