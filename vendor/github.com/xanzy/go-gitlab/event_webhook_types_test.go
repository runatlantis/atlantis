package gitlab

import (
	"encoding/json"
	"testing"
)

func TestPushEventUnmarshal(t *testing.T) {
	jsonObject := loadFixture("testdata/webhooks/push.json")
	var event *PushEvent
	err := json.Unmarshal(jsonObject, &event)

	if err != nil {
		t.Errorf("Push Event can not unmarshaled: %v\n ", err.Error())
	}

	if event == nil {
		t.Errorf("Push Event is null")
	}

	if event.ProjectID != 15 {
		t.Errorf("ProjectID is %v, want %v", event.ProjectID, 15)
	}

	if event.UserName != "John Smith" {
		t.Errorf("Username is %s, want %s", event.UserName, "John Smith")
	}

	if event.Commits[0] == nil || event.Commits[0].Timestamp == nil {
		t.Errorf("Commit Timestamp isn't nil")
	}

	if event.Commits[0] == nil || event.Commits[0].Author.Name != "Jordi Mallach" {
		t.Errorf("Commit Username is %s, want %s", event.UserName, "Jordi Mallach")
	}
}

func TestMergeEventUnmarshal(t *testing.T) {
	jsonObject := loadFixture("testdata/webhooks/merge_request.json")

	var event *MergeEvent
	err := json.Unmarshal(jsonObject, &event)

	if err != nil {
		t.Errorf("Merge Event can not unmarshaled: %v\n ", err.Error())
	}

	if event == nil {
		t.Errorf("Merge Event is null")
	}

	if event.ObjectAttributes.ID != 99 {
		t.Errorf("ObjectAttributes.ID is %v, want %v", event.ObjectAttributes.ID, 99)
	}

	if event.ObjectAttributes.Source.Homepage != "http://example.com/awesome_space/awesome_project" {
		t.Errorf("ObjectAttributes.Source.Homepage is %v, want %v", event.ObjectAttributes.Source.Homepage, "http://example.com/awesome_space/awesome_project")
	}

	if event.ObjectAttributes.LastCommit.ID != "da1560886d4f094c3e6c9ef40349f7d38b5d27d7" {
		t.Errorf("ObjectAttributes.LastCommit.ID is %v, want %s", event.ObjectAttributes.LastCommit.ID, "da1560886d4f094c3e6c9ef40349f7d38b5d27d7")
	}
	if event.ObjectAttributes.Assignee.Name != "User1" {
		t.Errorf("Assignee.Name is %v, want %v", event.ObjectAttributes.ID, "User1")
	}

	if event.ObjectAttributes.Assignee.Username != "user1" {
		t.Errorf("ObjectAttributes is %v, want %v", event.ObjectAttributes.Assignee.Username, "user1")
	}

	if event.User.Name == "" {
		t.Errorf("Username is %s, want %s", event.User.Name, "Administrator")
	}

	if event.ObjectAttributes.LastCommit.Timestamp == nil {
		t.Errorf("Timestamp isn't nil")
	}

	if name := event.ObjectAttributes.LastCommit.Author.Name; name != "GitLab dev user" {
		t.Errorf("Commit Username is %s, want %s", name, "GitLab dev user")
	}
}

func TestPipelineEventUnmarshal(t *testing.T) {
	jsonObject := loadFixture("testdata/webhooks/pipeline.json")

	var event *PipelineEvent
	err := json.Unmarshal(jsonObject, &event)

	if err != nil {
		t.Errorf("Pipeline Event can not unmarshaled: %v\n ", err.Error())
	}

	if event == nil {
		t.Errorf("Pipeline Event is null")
	}

	if event.ObjectAttributes.ID != 31 {
		t.Errorf("ObjectAttributes is %v, want %v", event.ObjectAttributes.ID, 1977)
	}

	if event.User.Name == "" {
		t.Errorf("Username is %s, want %s", event.User.Name, "Administrator")
	}

	if event.Commit.Timestamp == nil {
		t.Errorf("Timestamp isn't nil")
	}

	if name := event.Commit.Author.Name; name != "User" {
		t.Errorf("Commit Username is %s, want %s", name, "User")
	}
}

func TestBuildEventUnmarshal(t *testing.T) {
	jsonObject := loadFixture("testdata/webhooks/build.json")

	var event *BuildEvent
	err := json.Unmarshal(jsonObject, &event)

	if err != nil {
		t.Errorf("Build Event can not unmarshaled: %v\n ", err.Error())
	}

	if event == nil {
		t.Errorf("Build Event is null")
	}

	if event.BuildID != 1977 {
		t.Errorf("BuildID is %v, want %v", event.BuildID, 1977)
	}
}

func TestMergeEventUnmarshalFromGroup(t *testing.T) {
	jsonObject := loadFixture("testdata/webhooks/group_merge_request.json")

	var event *MergeEvent
	err := json.Unmarshal(jsonObject, &event)

	if err != nil {
		t.Errorf("Group Merge Event can not unmarshaled: %v\n ", err.Error())
	}

	if event == nil {
		t.Errorf("Group Merge Event is null")
	}

	if event.ObjectKind != "merge_request" {
		t.Errorf("ObjectKind is %v, want %v", event.ObjectKind, "merge_request")
	}

	if event.User.Username != "root" {
		t.Errorf("User.Username is %v, want %v", event.User.Username, "root")
	}

	if event.Project.Name != "example-project" {
		t.Errorf("Project.Name is %v, want %v", event.Project.Name, "example-project")
	}

	if event.ObjectAttributes.ID != 15917 {
		t.Errorf("ObjectAttributes.ID is %v, want %v", event.ObjectAttributes.ID, 15917)
	}

	if event.ObjectAttributes.Source.Name != "example-project" {
		t.Errorf("ObjectAttributes.Source.Name is %v, want %v", event.ObjectAttributes.Source.Name, "example-project")
	}

	if event.ObjectAttributes.LastCommit.Author.Email != "test.user@mail.com" {
		t.Errorf("ObjectAttributes.LastCommit.Author.Email is %v, want %v", event.ObjectAttributes.LastCommit.Author.Email, "test.user@mail.com")
	}

	if event.Repository.Name != "example-project" {
		t.Errorf("Repository.Name is %v, want %v", event.Repository.Name, "example-project")
	}

	if event.Assignee.Username != "root" {
		t.Errorf("Assignee.Username is %v, want %v", event.Assignee, "root")
	}

	if event.User.Name == "" {
		t.Errorf("Username is %s, want %s", event.User.Name, "Administrator")
	}

	if event.ObjectAttributes.LastCommit.Timestamp == nil {
		t.Errorf("Timestamp isn't nil")
	}

	if name := event.ObjectAttributes.LastCommit.Author.Name; name != "Test User" {
		t.Errorf("Commit Username is %s, want %s", name, "Test User")
	}
}
