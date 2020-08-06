//
// Copyright 2020, Eric Stevens
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package gitlab

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"
)

func TestListGroupHooks(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/groups/1/hooks", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `
[
	{
		"id": 1,
		"url": "http://example.com/hook",
		"group_id": 3,
		"push_events": true,
		"issues_events": true,
		"confidential_issues_events": true,
		"merge_requests_events": true,
		"tag_push_events": true,
		"note_events": true,
		"job_events": true,
		"pipeline_events": true,
		"wiki_page_events": true,
		"enable_ssl_verification": true,
		"created_at": "2012-10-12T17:04:47Z"
	}
]`)
	})

	groupHooks, _, err := client.Groups.ListGroupHooks(1)
	if err != nil {
		t.Error(err)
	}

	datePointer := time.Date(2012, 10, 12, 17, 4, 47, 0, time.UTC)
	want := []*GroupHook{{
		ID:                       1,
		URL:                      "http://example.com/hook",
		GroupID:                  3,
		PushEvents:               true,
		IssuesEvents:             true,
		ConfidentialIssuesEvents: true,
		MergeRequestsEvents:      true,
		TagPushEvents:            true,
		NoteEvents:               true,
		JobEvents:                true,
		PipelineEvents:           true,
		WikiPageEvents:           true,
		EnableSSLVerification:    true,
		CreatedAt:                &datePointer,
	}}

	if !reflect.DeepEqual(groupHooks, want) {
		t.Errorf("listGroupHooks returned \ngot:\n%v\nwant:\n%v", Stringify(groupHooks), Stringify(want))
	}
}

func TestGetGroupHook(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/groups/1/hooks/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `
{
	"id": 1,
	"url": "http://example.com/hook",
	"group_id": 3,
	"push_events": true,
	"issues_events": true,
	"confidential_issues_events": true,
	"merge_requests_events": true,
	"tag_push_events": true,
	"note_events": true,
	"job_events": true,
	"pipeline_events": true,
	"wiki_page_events": true,
	"enable_ssl_verification": true,
	"created_at": "2012-10-12T17:04:47Z"
}`)
	})

	groupHook, _, err := client.Groups.GetGroupHook(1, 1)
	if err != nil {
		t.Error(err)
	}

	datePointer := time.Date(2012, 10, 12, 17, 4, 47, 0, time.UTC)
	want := &GroupHook{
		ID:                       1,
		URL:                      "http://example.com/hook",
		GroupID:                  3,
		PushEvents:               true,
		IssuesEvents:             true,
		ConfidentialIssuesEvents: true,
		MergeRequestsEvents:      true,
		TagPushEvents:            true,
		NoteEvents:               true,
		JobEvents:                true,
		PipelineEvents:           true,
		WikiPageEvents:           true,
		EnableSSLVerification:    true,
		CreatedAt:                &datePointer,
	}

	if !reflect.DeepEqual(groupHook, want) {
		t.Errorf("getGroupHooks returned \ngot:\n%v\nwant:\n%v", Stringify(groupHook), Stringify(want))
	}
}

func TestAddGroupHook(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/groups/1/hooks", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		fmt.Fprint(w, `
{
	"id": 1,
	"url": "http://example.com/hook",
	"group_id": 3,
	"push_events": true,
	"issues_events": true,
	"confidential_issues_events": true,
	"merge_requests_events": true,
	"tag_push_events": true,
	"note_events": true,
	"job_events": true,
	"pipeline_events": true,
	"wiki_page_events": true,
	"enable_ssl_verification": true,
	"created_at": "2012-10-12T17:04:47Z"
}`)
	})

	url := "http://www.example.com/hook"
	opt := &AddGroupHookOptions{
		URL: &url,
	}

	groupHooks, _, err := client.Groups.AddGroupHook(1, opt)
	if err != nil {
		t.Error(err)
	}

	datePointer := time.Date(2012, 10, 12, 17, 4, 47, 0, time.UTC)
	want := &GroupHook{
		ID:                       1,
		URL:                      "http://example.com/hook",
		GroupID:                  3,
		PushEvents:               true,
		IssuesEvents:             true,
		ConfidentialIssuesEvents: true,
		ConfidentialNoteEvents:   false,
		MergeRequestsEvents:      true,
		TagPushEvents:            true,
		NoteEvents:               true,
		JobEvents:                true,
		PipelineEvents:           true,
		WikiPageEvents:           true,
		EnableSSLVerification:    true,
		CreatedAt:                &datePointer,
	}

	if !reflect.DeepEqual(groupHooks, want) {
		t.Errorf("AddGroupHook returned \ngot:\n%v\nwant:\n%v", Stringify(groupHooks), Stringify(want))
	}
}

func TestEditGroupHook(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/groups/1/hooks/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PUT")
		fmt.Fprint(w, `
{
	"id": 1,
	"url": "http://example.com/hook",
	"group_id": 3,
	"push_events": true,
	"issues_events": true,
	"confidential_issues_events": true,
	"merge_requests_events": true,
	"tag_push_events": true,
	"note_events": true,
	"job_events": true,
	"pipeline_events": true,
	"wiki_page_events": true,
	"enable_ssl_verification": true,
	"created_at": "2012-10-12T17:04:47Z"
}`)
	})

	url := "http://www.example.com/hook"
	opt := &EditGroupHookOptions{
		URL: &url,
	}

	groupHooks, _, err := client.Groups.EditGroupHook(1, 1, opt)
	if err != nil {
		t.Error(err)
	}

	datePointer := time.Date(2012, 10, 12, 17, 4, 47, 0, time.UTC)
	want := &GroupHook{
		ID:                       1,
		URL:                      "http://example.com/hook",
		GroupID:                  3,
		PushEvents:               true,
		IssuesEvents:             true,
		ConfidentialIssuesEvents: true,
		ConfidentialNoteEvents:   false,
		MergeRequestsEvents:      true,
		TagPushEvents:            true,
		NoteEvents:               true,
		JobEvents:                true,
		PipelineEvents:           true,
		WikiPageEvents:           true,
		EnableSSLVerification:    true,
		CreatedAt:                &datePointer,
	}

	if !reflect.DeepEqual(groupHooks, want) {
		t.Errorf("EditGroupHook returned \ngot:\n%v\nwant:\n%v", Stringify(groupHooks), Stringify(want))
	}
}

func TestDeleteGroupHook(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/groups/1/hooks/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
	})

	_, err := client.Groups.DeleteGroupHook(1, 1)
	if err != nil {
		t.Error(err)
	}
}
