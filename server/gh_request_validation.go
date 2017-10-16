package server

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/google/go-github/github"
)

// GHRequestValidator validates GitHub requests.
type GHRequestValidator interface {
	// Validate returns the JSON payload of the request.
	// If secret is not empty, it checks that the request was signed
	// by secret and returns an error if it was not.
	// If secret is empty, it does not check if the request was signed.
	Validate(r *http.Request, secret []byte) ([]byte, error)
}

// GHRequestValidation implements the GHRequestValidator interface.
type GHRequestValidation struct{}

// Validate returns the JSON payload of the request.
// If secret is not empty, it checks that the request was signed
// by secret and returns an error if it was not.
// If secret is empty, it does not check if the request was signed.
func (g *GHRequestValidation) Validate(r *http.Request, secret []byte) ([]byte, error) {
	if len(secret) != 0 {
		return g.validateAgainstSecret(r, secret)
	}
	return g.validateWithoutSecret(r)
}

func (g *GHRequestValidation) validateAgainstSecret(r *http.Request, secret []byte) ([]byte, error) {
	payload, err := github.ValidatePayload(r, secret)
	if err != nil {
		return nil, err
	}
	return payload, nil
}

func (g *GHRequestValidation) validateWithoutSecret(r *http.Request) ([]byte, error) {
	switch ct := r.Header.Get("Content-Type"); ct {
	case "application/json":
		payload, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return nil, fmt.Errorf("could not read body: %s", err)
		}
		return payload, nil
	case "application/x-www-form-urlencoded":
		// GitHub stores the json payload as a form value
		payloadForm := r.FormValue("payload")
		if payloadForm == "" {
			return nil, errors.New("webhook request did not contain expected 'payload' form value")
		}
		return []byte(payloadForm), nil
	default:
		return nil, fmt.Errorf("webhook request has unsupported Content-Type %q", ct)
	}
}
