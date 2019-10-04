// Copyright 2016 The go-github AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file provides functions for validating payloads from GitHub Webhooks.
// GitHub API docs: https://developer.github.com/webhooks/securing/#validating-payloads-from-github

// Adapted for Azure Devops

package azuredevops

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	activityIDHeader     = "X-VSS-ActivityId"
	subscriptionIDHeader = "X-VSS-SubscriptionId"
	// requestIDHeader is the Azure Devops header key used to pass the unique ID for the webhook event.
	requestIDHeader = "Request-Id"
)

var (
	// eventTypeMapping maps webhooks types to their corresponding go-azuredevops
	// resource struct types.
	eventTypeMapping = map[string]string{
		"git.pullrequest.created":                   "PullRequestEvent",
		"git.pullrequest.merged":                    "PullRequestEvent",
		"git.pullrequest.updated":                   "PullRequestEvent",
		"git.push":                                  "PushEvent",
		"ms.vss-code.git-pullrequest-comment-event": "PullRequestCommentedEvent",
		"workitem.commented":                        "WorkItemCommentedEvent",
		"workitem.updated":                          "WorkItemUpdatedEvent",
	}
)

// GetActivityID returns the value of the activityIDHeader webhook header.
//
// Haven't found vendor documentation yet.  This could be a GUID that identifies
// the webhook request ID.  A different GUID is also present in the body of
// webhook requests.
func GetActivityID(r *http.Request) string {
	return r.Header.Get(activityIDHeader)
}

// GetRequestID returns the value of the requestIDHeader webhook header.
//
// Haven't found vendor documentation yet.  This could be a GUID that identifies
// the webhook request ID.  A different GUID is also present in the body of
// webhook requests.
func GetRequestID(r *http.Request) string {
	return r.Header.Get(requestIDHeader)
}

// GetSubscriptionID returns the value of the subscriptionIDHeader webhook header.
//
// Haven't found vendor documentation yet.  This could be a GUID that identifies
// the webhook event type and settings in the Azure Devops tenant
func GetSubscriptionID(r *http.Request) string {
	return r.Header.Get(subscriptionIDHeader)
}

// ValidatePayload validates an incoming Azure Devops Webhook event request
// and returns the (JSON) payload.
// The Content-Type header of the payload must be "application/json" or
// an error is returned.  A charset may be included with the content type.
// user is the supplied username for Basic authentication
// pass is the supplied password for Basic authentication
// If your webhook does not contain a username or password, you can pass nil or an empty slice.
// This is intended for local development purposes only as all webhooks should ideally
// set up a secret token.
// It is up to the caller to process failed validation and return a proper 401 response
// to the user, such as:
//
// w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
// w.WriteHeader(401)
// w.Write([]byte("Unauthorized.\n"))
//
//
// Example usage:
//
//     func (s *Event) ServeHTTP(w http.ResponseWriter, r *http.Request) {
//       payload, err := azuredevops.ValidatePayload(r, s.user, s.pass)
//       if err != nil { ... }
//       // Process payload...
//     }
//
func ValidatePayload(r *http.Request, user, pass []byte) (payload []byte, err error) {
	var body []byte // Raw body that GitHub uses to calculate the signature.

	switch ct := r.Header.Get("Content-Type"); ct {
	case "application/json":
		var err error
		if body, err = ioutil.ReadAll(r.Body); err != nil {
			return nil, err
		}

		// If the content type is application/json,
		// the JSON payload is just the original body.
		payload = body
	case "application/json; charset=utf-8":
		var err error
		if body, err = ioutil.ReadAll(r.Body); err != nil {
			return nil, err
		}

		// If the content type is application/json,
		// the JSON payload is just the original body.
		payload = body
	default:
		return nil, fmt.Errorf("Webhook request has unsupported Content-Type %q", ct)
	}

	// Only validate the authentication if a username and password exist. This is
	// intended for local development only and all webhooks should ideally set up
	// a Basic authentication username and password.
	username, password, ok := r.BasicAuth()

	if !ok || subtle.ConstantTimeCompare([]byte(user), []byte(username)) != 1 || subtle.ConstantTimeCompare([]byte(pass), []byte(password)) != 1 {
		return nil, errors.New("ValidatePayload authentication failed")
	}
	//Authorization: Basic <credentials>
	return payload, nil
}

// ParseWebHook parses the event payload into a corresponding struct.
// An error will be returned for unrecognized event types.
//
// Example usage:
//
//     func (s *EventMonitor) ServeHTTP(w http.ResponseWriter, r *http.Request) {
//       payload, err := azuredevops.ValidatePayload(r, s.user, s.pass)
//       if err != nil { ... }
//       event, err := azuredevops.ParseWebHook(payload)
//       if err != nil { ... }
//       switch event.PayloadType {
//	 	 case azuredevops.WorkItemEvent:
//		        processWorkItemEvent(&event)
//	     case azuredevops.PullRequestEvent:
//              processPullRequestEvent(&event)
//       ...
//       }
//     }
//
// https://docs.microsoft.com/en-us/azure/devops/service-hooks/events?view=azure-devops
func ParseWebHook(payload []byte) (*Event, error) {
	event := new(Event) // returns pointer
	err := json.Unmarshal(payload, event)
	if err != nil {
		return nil, err
	}

	_, err = event.ParsePayload()

	return event, err
}
