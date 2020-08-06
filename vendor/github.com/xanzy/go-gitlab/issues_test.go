package gitlab

import (
	"fmt"
	"log"
	"net/http"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetIssue(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/issues/5", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"id":1, "description": "This is test project", "author" : {"id" : 1, "name": "snehal"}, "assignees":[{"id":1}],"merge_requests_count": 1}`)
	})

	issue, _, err := client.Issues.GetIssue("1", 5)
	if err != nil {
		log.Fatal(err)
	}

	want := &Issue{
		ID:                1,
		Description:       "This is test project",
		Author:            &IssueAuthor{ID: 1, Name: "snehal"},
		Assignees:         []*IssueAssignee{{ID: 1}},
		MergeRequestCount: 1,
	}

	if !reflect.DeepEqual(want, issue) {
		t.Errorf("Issues.GetIssue returned %+v, want %+v", issue, want)
	}
}

func TestDeleteIssue(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/issues/5", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
		fmt.Fprint(w, `{"id":1, "description": "This is test project", "author" : {"id" : 1, "name": "snehal"}, "assignees":[{"id":1}]}`)
	})

	_, err := client.Issues.DeleteIssue("1", 5)
	if err != nil {
		log.Fatal(err)
	}
}

func TestMoveIssue(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/issues/11/move", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		mustWriteHTTPResponse(t, w, "testdata/issue_move.json")
	})

	issue, _, err := client.Issues.MoveIssue("1", 11, &MoveIssueOptions{ToProjectID: Int(5)})
	if err != nil {
		log.Fatal(err)
	}

	want := &Issue{
		ID:        92,
		IID:       11,
		ProjectID: 5,
	}

	assert.Equal(t, want.IID, issue.IID)
	assert.Equal(t, want.ProjectID, issue.ProjectID)
}

func TestListIssues(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/issues", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testURL(t, r, "/api/v4/issues?assignee_id=2&author_id=1")
		fmt.Fprint(w, `[{"id":1, "description": "This is test project", "author" : {"id" : 1, "name": "snehal"}, "assignees":[{"id":1}]}]`)
	})

	listProjectIssue := &ListIssuesOptions{
		AuthorID:   Int(01),
		AssigneeID: Int(02),
	}

	issues, _, err := client.Issues.ListIssues(listProjectIssue)

	if err != nil {
		log.Fatal(err)
	}

	want := []*Issue{{
		ID:          1,
		Description: "This is test project",
		Author:      &IssueAuthor{ID: 1, Name: "snehal"},
		Assignees:   []*IssueAssignee{{ID: 1}},
	}}

	if !reflect.DeepEqual(want, issues) {
		t.Errorf("Issues.ListIssues returned %+v, want %+v", issues, want)
	}
}

func TestListProjectIssues(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/issues", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testURL(t, r, "/api/v4/projects/1/issues?assignee_id=2&author_id=1")
		fmt.Fprint(w, `[{"id":1, "description": "This is test project", "author" : {"id" : 1, "name": "snehal"}, "assignees":[{"id":1}]}]`)
	})

	listProjectIssue := &ListProjectIssuesOptions{
		AuthorID:   Int(01),
		AssigneeID: Int(02),
	}
	issues, _, err := client.Issues.ListProjectIssues("1", listProjectIssue)
	if err != nil {
		log.Fatal(err)
	}

	want := []*Issue{{
		ID:          1,
		Description: "This is test project",
		Author:      &IssueAuthor{ID: 1, Name: "snehal"},
		Assignees:   []*IssueAssignee{{ID: 1}},
	}}

	if !reflect.DeepEqual(want, issues) {
		t.Errorf("Issues.ListProjectIssues returned %+v, want %+v", issues, want)
	}
}

func TestListGroupIssues(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/groups/1/issues", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testURL(t, r, "/api/v4/groups/1/issues?assignee_id=2&author_id=1&state=Open")
		fmt.Fprint(w, `[{"id":1, "description": "This is test project", "author" : {"id" : 1, "name": "snehal"}, "assignees":[{"id":1}]}]`)
	})

	listGroupIssue := &ListGroupIssuesOptions{
		State:      String("Open"),
		AuthorID:   Int(01),
		AssigneeID: Int(02),
	}

	issues, _, err := client.Issues.ListGroupIssues("1", listGroupIssue)
	if err != nil {
		log.Fatal(err)
	}

	want := []*Issue{{
		ID:          1,
		Description: "This is test project",
		Author:      &IssueAuthor{ID: 1, Name: "snehal"},
		Assignees:   []*IssueAssignee{{ID: 1}},
	}}

	if !reflect.DeepEqual(want, issues) {
		t.Errorf("Issues.ListGroupIssues returned %+v, want %+v", issues, want)
	}
}

func TestCreateIssue(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/issues", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		fmt.Fprint(w, `{"id":1, "title" : "Title of issue", "description": "This is description of an issue", "author" : {"id" : 1, "name": "snehal"}, "assignees":[{"id":1}]}`)
	})

	createIssueOptions := &CreateIssueOptions{
		Title:       String("Title of issue"),
		Description: String("This is description of an issue"),
	}

	issue, _, err := client.Issues.CreateIssue("1", createIssueOptions)

	if err != nil {
		log.Fatal(err)
	}

	want := &Issue{
		ID:          1,
		Title:       "Title of issue",
		Description: "This is description of an issue",
		Author:      &IssueAuthor{ID: 1, Name: "snehal"},
		Assignees:   []*IssueAssignee{{ID: 1}},
	}

	if !reflect.DeepEqual(want, issue) {
		t.Errorf("Issues.CreateIssue returned %+v, want %+v", issue, want)
	}
}

func TestUpdateIssue(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/issues/5", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PUT")
		fmt.Fprint(w, `{"id":1, "title" : "Title of issue", "description": "This is description of an issue", "author" : {"id" : 1, "name": "snehal"}, "assignees":[{"id":1}]}`)
	})

	updateIssueOpt := &UpdateIssueOptions{
		Title:       String("Title of issue"),
		Description: String("This is description of an issue"),
	}
	issue, _, err := client.Issues.UpdateIssue(1, 5, updateIssueOpt)

	if err != nil {
		log.Fatal(err)
	}

	want := &Issue{
		ID:          1,
		Title:       "Title of issue",
		Description: "This is description of an issue",
		Author:      &IssueAuthor{ID: 1, Name: "snehal"},
		Assignees:   []*IssueAssignee{{ID: 1}},
	}

	if !reflect.DeepEqual(want, issue) {
		t.Errorf("Issues.UpdateIssue returned %+v, want %+v", issue, want)
	}
}

func TestSubscribeToIssue(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/issues/5/subscribe", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		fmt.Fprint(w, `{"id":1, "title" : "Title of issue", "description": "This is description of an issue", "author" : {"id" : 1, "name": "snehal"}, "assignees":[{"id":1}]}`)
	})

	issue, _, err := client.Issues.SubscribeToIssue("1", 5)

	if err != nil {
		log.Fatal(err)
	}

	want := &Issue{
		ID:          1,
		Title:       "Title of issue",
		Description: "This is description of an issue",
		Author:      &IssueAuthor{ID: 1, Name: "snehal"},
		Assignees:   []*IssueAssignee{{ID: 1}},
	}

	if !reflect.DeepEqual(want, issue) {
		t.Errorf("Issues.SubscribeToIssue returned %+v, want %+v", issue, want)
	}
}

func TestUnsubscribeFromIssue(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/issues/5/unsubscribe", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		fmt.Fprint(w, `{"id":1, "title" : "Title of issue", "description": "This is description of an issue", "author" : {"id" : 1, "name": "snehal"}, "assignees":[{"id":1}]}`)
	})

	issue, _, err := client.Issues.UnsubscribeFromIssue("1", 5)
	if err != nil {
		log.Fatal(err)
	}

	want := &Issue{
		ID:          1,
		Title:       "Title of issue",
		Description: "This is description of an issue",
		Author:      &IssueAuthor{ID: 1, Name: "snehal"},
		Assignees:   []*IssueAssignee{{ID: 1}},
	}

	if !reflect.DeepEqual(want, issue) {
		t.Errorf("Issues.UnsubscribeFromIssue returned %+v, want %+v", issue, want)
	}
}

func TestListMergeRequestsClosingIssue(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/issues/5/closed_by", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testURL(t, r, "/api/v4/projects/1/issues/5/closed_by?page=1&per_page=10")

		fmt.Fprint(w, `[{"id":1, "title" : "test merge one"},{"id":2, "title" : "test merge two"}]`)
	})

	listMergeRequestsClosingIssueOpt := &ListMergeRequestsClosingIssueOptions{
		Page:    1,
		PerPage: 10,
	}
	mergeRequest, _, err := client.Issues.ListMergeRequestsClosingIssue("1", 5, listMergeRequestsClosingIssueOpt)
	if err != nil {
		log.Fatal(err)
	}

	want := []*MergeRequest{{ID: 1, Title: "test merge one"}, {ID: 2, Title: "test merge two"}}

	if !reflect.DeepEqual(want, mergeRequest) {
		t.Errorf("Issues.ListMergeRequestsClosingIssue returned %+v, want %+v", mergeRequest, want)
	}
}

func TestListMergeRequestsRelatedToIssue(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/issues/5/related_merge_requests", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testURL(t, r, "/api/v4/projects/1/issues/5/related_merge_requests?page=1&per_page=10")

		fmt.Fprint(w, `[{"id":1, "title" : "test merge one"},{"id":2, "title" : "test merge two"}]`)
	})

	listMergeRequestsRelatedToIssueOpt := &ListMergeRequestsRelatedToIssueOptions{
		Page:    1,
		PerPage: 10,
	}
	mergeRequest, _, err := client.Issues.ListMergeRequestsRelatedToIssue("1", 5, listMergeRequestsRelatedToIssueOpt)
	if err != nil {
		log.Fatal(err)
	}

	want := []*MergeRequest{{ID: 1, Title: "test merge one"}, {ID: 2, Title: "test merge two"}}

	if !reflect.DeepEqual(want, mergeRequest) {
		t.Errorf("Issues.ListMergeRequestsClosingIssue returned %+v, want %+v", mergeRequest, want)
	}
}

func TestSetTimeEstimate(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/issues/5/time_estimate", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		fmt.Fprint(w, `{"human_time_estimate": "3h 30m", "human_total_time_spent": null, "time_estimate": 12600, "total_time_spent": 0}`)
	})

	setTimeEstiOpt := &SetTimeEstimateOptions{
		Duration: String("3h 30m"),
	}

	timeState, _, err := client.Issues.SetTimeEstimate("1", 5, setTimeEstiOpt)
	if err != nil {
		log.Fatal(err)
	}
	want := &TimeStats{HumanTimeEstimate: "3h 30m", HumanTotalTimeSpent: "", TimeEstimate: 12600, TotalTimeSpent: 0}

	if !reflect.DeepEqual(want, timeState) {
		t.Errorf("Issues.SetTimeEstimate returned %+v, want %+v", timeState, want)
	}
}

func TestResetTimeEstimate(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/issues/5/reset_time_estimate", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		fmt.Fprint(w, `{"human_time_estimate": null, "human_total_time_spent": null, "time_estimate": 0, "total_time_spent": 0}`)
	})

	timeState, _, err := client.Issues.ResetTimeEstimate("1", 5)
	if err != nil {
		log.Fatal(err)
	}
	want := &TimeStats{HumanTimeEstimate: "", HumanTotalTimeSpent: "", TimeEstimate: 0, TotalTimeSpent: 0}

	if !reflect.DeepEqual(want, timeState) {
		t.Errorf("Issues.ResetTimeEstimate returned %+v, want %+v", timeState, want)
	}
}

func TestAddSpentTime(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/issues/5/add_spent_time", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		testURL(t, r, "/api/v4/projects/1/issues/5/add_spent_time")
		fmt.Fprint(w, `{"human_time_estimate": null, "human_total_time_spent": "1h", "time_estimate": 0, "total_time_spent": 3600}`)
	})
	addSpentTimeOpt := &AddSpentTimeOptions{
		Duration: String("1h"),
	}

	timeState, _, err := client.Issues.AddSpentTime("1", 5, addSpentTimeOpt)
	if err != nil {
		log.Fatal(err)
	}
	want := &TimeStats{HumanTimeEstimate: "", HumanTotalTimeSpent: "1h", TimeEstimate: 0, TotalTimeSpent: 3600}

	if !reflect.DeepEqual(want, timeState) {
		t.Errorf("Issues.AddSpentTime returned %+v, want %+v", timeState, want)
	}
}

func TestResetSpentTime(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/issues/5/reset_spent_time", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		testURL(t, r, "/api/v4/projects/1/issues/5/reset_spent_time")
		fmt.Fprint(w, `{"human_time_estimate": null, "human_total_time_spent": "", "time_estimate": 0, "total_time_spent": 0}`)
	})

	timeState, _, err := client.Issues.ResetSpentTime("1", 5)
	if err != nil {
		log.Fatal(err)
	}

	want := &TimeStats{HumanTimeEstimate: "", HumanTotalTimeSpent: "", TimeEstimate: 0, TotalTimeSpent: 0}
	if !reflect.DeepEqual(want, timeState) {
		t.Errorf("Issues.ResetSpentTime returned %+v, want %+v", timeState, want)
	}
}

func TestGetTimeSpent(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/issues/5/time_stats", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testURL(t, r, "/api/v4/projects/1/issues/5/time_stats")
		fmt.Fprint(w, `{"human_time_estimate": "2h", "human_total_time_spent": "1h", "time_estimate": 7200, "total_time_spent": 3600}`)
	})

	timeState, _, err := client.Issues.GetTimeSpent("1", 5)
	if err != nil {
		log.Fatal(err)
	}

	want := &TimeStats{HumanTimeEstimate: "2h", HumanTotalTimeSpent: "1h", TimeEstimate: 7200, TotalTimeSpent: 3600}
	if !reflect.DeepEqual(want, timeState) {
		t.Errorf("Issues.GetTimeSpent returned %+v, want %+v", timeState, want)
	}
}
