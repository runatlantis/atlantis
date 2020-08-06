package gitlab

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestGetDroneCIService(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/services/drone-ci", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"id":1}`)
	})
	want := &DroneCIService{Service: Service{ID: 1}}

	service, _, err := client.Services.GetDroneCIService(1)
	if err != nil {
		t.Fatalf("Services.GetDroneCIService returns an error: %v", err)
	}
	if !reflect.DeepEqual(want, service) {
		t.Errorf("Services.GetDroneCIService returned %+v, want %+v", service, want)
	}
}

func TestSetDroneCIService(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/services/drone-ci", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PUT")
	})

	opt := &SetDroneCIServiceOptions{String("t"), String("u"), Bool(true)}

	_, err := client.Services.SetDroneCIService(1, opt)
	if err != nil {
		t.Fatalf("Services.SetDroneCIService returns an error: %v", err)
	}
}

func TestDeleteDroneCIService(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/services/drone-ci", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
	})

	_, err := client.Services.DeleteDroneCIService(1)
	if err != nil {
		t.Fatalf("Services.DeleteDroneCIService returns an error: %v", err)
	}
}

func TestGetJiraService(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/0/services/jira", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"id":1, "properties": {"jira_issue_transition_id": "2"}}`)
	})

	mux.HandleFunc("/api/v4/projects/1/services/jira", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"id":1, "properties": {"jira_issue_transition_id": 2}}`)
	})

	mux.HandleFunc("/api/v4/projects/2/services/jira", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"id":1, "properties": {"jira_issue_transition_id": "2,3"}}`)
	})

	mux.HandleFunc("/api/v4/projects/3/services/jira", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"id":1, "properties": {}}`)
	})

	want := []*JiraService{
		{
			Service: Service{ID: 1},
			Properties: &JiraServiceProperties{
				JiraIssueTransitionID: "2",
			},
		},
		{
			Service: Service{ID: 1},
			Properties: &JiraServiceProperties{
				JiraIssueTransitionID: "2",
			},
		},
		{
			Service: Service{ID: 1},
			Properties: &JiraServiceProperties{
				JiraIssueTransitionID: "2,3",
			},
		},
		{
			Service:    Service{ID: 1},
			Properties: &JiraServiceProperties{},
		},
	}

	for testcase := 0; testcase < len(want); testcase++ {
		service, _, err := client.Services.GetJiraService(testcase)
		if err != nil {
			t.Fatalf("Services.GetJiraService returns an error: %v", err)
		}

		if !reflect.DeepEqual(want[testcase], service) {
			t.Errorf("Services.GetJiraService returned %+v, want %+v", service, want[testcase])
		}
	}
}

func TestSetJiraService(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/services/jira", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PUT")
	})

	opt := &SetJiraServiceOptions{
		URL:                   String("asd"),
		APIURL:                String("asd"),
		ProjectKey:            String("as"),
		Username:              String("aas"),
		Password:              String("asd"),
		Active:                Bool(true),
		JiraIssueTransitionID: String("2,3"),
		CommitEvents:          Bool(true),
		CommentOnEventEnabled: Bool(true),
		MergeRequestsEvents:   Bool(true),
	}

	_, err := client.Services.SetJiraService(1, opt)
	if err != nil {
		t.Fatalf("Services.SetJiraService returns an error: %v", err)
	}
}

func TestDeleteJiraService(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/services/jira", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
	})

	_, err := client.Services.DeleteJiraService(1)
	if err != nil {
		t.Fatalf("Services.DeleteJiraService returns an error: %v", err)
	}
}

func TestGetSlackService(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/services/slack", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"id":1}`)
	})
	want := &SlackService{Service: Service{ID: 1}}

	service, _, err := client.Services.GetSlackService(1)
	if err != nil {
		t.Fatalf("Services.GetSlackService returns an error: %v", err)
	}
	if !reflect.DeepEqual(want, service) {
		t.Errorf("Services.GetSlackService returned %+v, want %+v", service, want)
	}
}

func TestSetSlackService(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/services/slack", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PUT")
	})

	opt := &SetSlackServiceOptions{
		WebHook:  String("webhook_uri"),
		Username: String("username"),
		Channel:  String("#development"),
	}

	_, err := client.Services.SetSlackService(1, opt)
	if err != nil {
		t.Fatalf("Services.SetSlackService returns an error: %v", err)
	}
}

func TestDeleteSlackService(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/services/slack", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
	})

	_, err := client.Services.DeleteSlackService(1)
	if err != nil {
		t.Fatalf("Services.DeleteSlackService returns an error: %v", err)
	}
}

func TestGetPipelinesEmailService(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/services/pipelines-email", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"id":1}`)
	})
	want := &PipelinesEmailService{Service: Service{ID: 1}}

	service, _, err := client.Services.GetPipelinesEmailService(1)
	if err != nil {
		t.Fatalf("Services.GetPipelinesEmailService returns an error: %v", err)
	}
	if !reflect.DeepEqual(want, service) {
		t.Errorf("Services.GetPipelinesEmailService returned %+v, want %+v", service, want)
	}
}

func TestSetPipelinesEmailService(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/services/pipelines-email", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PUT")
	})

	opt := &SetPipelinesEmailServiceOptions{
		Recipients:                String("test@email.com"),
		NotifyOnlyBrokenPipelines: Bool(true),
		NotifyOnlyDefaultBranch:   Bool(false),
		AddPusher:                 nil,
		BranchesToBeNotified:      nil,
		PipelineEvents:            nil,
	}

	_, err := client.Services.SetPipelinesEmailService(1, opt)
	if err != nil {
		t.Fatalf("Services.SetPipelinesEmailService returns an error: %v", err)
	}
}

func TestDeletePipelinesEmailService(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/services/pipelines-email", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
	})

	_, err := client.Services.DeletePipelinesEmailService(1)
	if err != nil {
		t.Fatalf("Services.DeletePipelinesEmailService returns an error: %v", err)
	}
}

func TestCustomIssueTrackerService(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/services/custom-issue-tracker", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"id": 1, "title": "5", "push_events": true, "properties": {"new_issue_url":"1", "issues_url": "2", "project_url": "3"}}`)
	})
	want := &CustomIssueTrackerService{
		Service: Service{
			ID:         1,
			Title:      "5",
			PushEvents: true,
		},
		Properties: &CustomIssueTrackerServiceProperties{
			NewIssueURL: "1",
			IssuesURL:   "2",
			ProjectURL:  "3",
		},
	}

	service, _, err := client.Services.GetCustomIssueTrackerService(1)
	if err != nil {
		t.Fatalf("Services.GetCustomIssueTrackerService returns an error: %v", err)
	}
	if !reflect.DeepEqual(want, service) {
		t.Errorf("Services.GetCustomIssueTrackerService returned %+v, want %+v", service, want)
	}
}

func TestSetCustomIssueTrackerService(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/services/custom-issue-tracker", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PUT")
	})

	opt := &SetCustomIssueTrackerServiceOptions{
		NewIssueURL: String("1"),
		IssuesURL:   String("2"),
		ProjectURL:  String("3"),
		Description: String("4"),
		Title:       String("5"),
		PushEvents:  Bool(true),
	}

	_, err := client.Services.SetCustomIssueTrackerService(1, opt)
	if err != nil {
		t.Fatalf("Services.SetCustomIssueTrackerService returns an error: %v", err)
	}
}

func TestDeleteCustomIssueTrackerService(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/services/custom-issue-tracker", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
	})

	_, err := client.Services.DeleteCustomIssueTrackerService(1)
	if err != nil {
		t.Fatalf("Services.DeleteCustomIssueTrackerService returns an error: %v", err)
	}
}
