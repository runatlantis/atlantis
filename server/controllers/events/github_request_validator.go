// Copyright 2017 HootSuite Media Inc.
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// Modified hereafter by contributors to runatlantis/atlantis.

package events

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/google/go-github/v66/github"
)

//go:generate pegomock generate --package mocks -o mocks/mock_github_request_validator.go GithubRequestValidator

// GithubRequestValidator handles checking if GitHub requests are signed
// properly by the secret.
type GithubRequestValidator interface {
	// Validate returns the JSON payload of the request.
	// If secret is not empty, it checks that the request was signed
	// by secret and returns an error if it was not.
	// If secret is empty, it does not check if the request was signed.
	Validate(r *http.Request, secret []byte) ([]byte, error)
}

// DefaultGithubRequestValidator handles checking if GitHub requests are signed
// properly by the secret.
type DefaultGithubRequestValidator struct{}

// Validate returns the JSON payload of the request.
// If secret is not empty, it checks that the request was signed
// by secret and returns an error if it was not.
// If secret is empty, it does not check if the request was signed.
func (d *DefaultGithubRequestValidator) Validate(r *http.Request, secret []byte) ([]byte, error) {
	if len(secret) != 0 {
		return d.validateAgainstSecret(r, secret)
	}
	return d.validateWithoutSecret(r)
}

func (d *DefaultGithubRequestValidator) validateAgainstSecret(r *http.Request, secret []byte) ([]byte, error) {
	payload, err := github.ValidatePayload(r, secret)
	if err != nil {
		return nil, err
	}
	return payload, nil
}

func (d *DefaultGithubRequestValidator) validateWithoutSecret(r *http.Request) ([]byte, error) {
	switch ct := r.Header.Get("Content-Type"); ct {
	case "application/json":
		payload, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, fmt.Errorf("could not read body: %s", err)
		}
		return payload, nil
	case "application/x-www-form-urlencoded":
		// GitHub stores the json payload as a form value.
		payloadForm := r.FormValue("payload")
		if payloadForm == "" {
			return nil, errors.New("webhook request did not contain expected 'payload' form value")
		}
		return []byte(payloadForm), nil
	default:
		return nil, fmt.Errorf("webhook request has unsupported Content-Type %q", ct)
	}
}
