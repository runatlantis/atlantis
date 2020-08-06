package gitlab

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseSystemhookPush(t *testing.T) {
	payload := loadFixture("testdata/systemhooks/push.json")

	parsedEvent, err := ParseSystemhook(payload)
	if err != nil {
		t.Errorf("Error parsing build hook: %s", err)
	}

	event, ok := parsedEvent.(*PushSystemEvent)
	if !ok {
		t.Errorf("Expected PushSystemHookEvent, but parsing produced %T", parsedEvent)
	}
	assert.Equal(t, "push", event.EventName)
}

func TestParseSystemhookTagPush(t *testing.T) {
	payload := loadFixture("testdata/systemhooks/tag_push.json")

	parsedEvent, err := ParseSystemhook(payload)
	if err != nil {
		t.Errorf("Error parsing build hook: %s", err)
	}

	event, ok := parsedEvent.(*TagPushSystemEvent)
	if !ok {
		t.Errorf("Expected TagPushSystemHookEvent, but parsing produced %T", parsedEvent)
	}
	assert.Equal(t, "tag_push", event.EventName)
}

func TestParseSystemhookMergeRequest(t *testing.T) {
	payload := loadFixture("testdata/systemhooks/merge_request.json")

	parsedEvent, err := ParseSystemhook(payload)
	if err != nil {
		t.Errorf("Error parsing build hook: %s", err)
	}

	event, ok := parsedEvent.(*MergeEvent)
	if !ok {
		t.Errorf("Expected MergeRequestSystemHookEvent, but parsing produced %T", parsedEvent)
	}
	assert.Equal(t, "merge_request", event.ObjectKind)
}

func TestParseSystemhookRepositoryUpdate(t *testing.T) {
	payload := loadFixture("testdata/systemhooks/repository_update.json")

	parsedEvent, err := ParseSystemhook(payload)
	if err != nil {
		t.Errorf("Error parsing build hook: %s", err)
	}

	event, ok := parsedEvent.(*RepositoryUpdateSystemEvent)
	if !ok {
		t.Errorf("Expected RepositoryUpdateSystemHookEvent, but parsing produced %T", parsedEvent)
	}
	assert.Equal(t, "repository_update", event.EventName)
}

func TestParseSystemhookProject(t *testing.T) {
	var tests = []struct {
		event   string
		payload []byte
	}{
		{"project_create", loadFixture("testdata/systemhooks/project_create.json")},
		{"project_update", loadFixture("testdata/systemhooks/project_update.json")},
		{"project_destroy", loadFixture("testdata/systemhooks/project_destroy.json")},
		{"project_transfer", loadFixture("testdata/systemhooks/project_transfer.json")},
		{"project_rename", loadFixture("testdata/systemhooks/project_rename.json")},
	}
	for _, tc := range tests {
		t.Run(tc.event, func(t *testing.T) {
			parsedEvent, err := ParseSystemhook(tc.payload)
			if err != nil {
				t.Errorf("Error parsing build hook: %s", err)
			}
			event, ok := parsedEvent.(*ProjectSystemEvent)
			if !ok {
				t.Errorf("Expected ProjectSystemHookEvent, but parsing produced %T", parsedEvent)
			}
			assert.Equal(t, tc.event, event.EventName)
		})
	}
}

func TestParseSystemhookGroup(t *testing.T) {
	var tests = []struct {
		event   string
		payload []byte
	}{
		{"group_create", loadFixture("testdata/systemhooks/group_create.json")},
		{"group_destroy", loadFixture("testdata/systemhooks/group_destroy.json")},
		{"group_rename", loadFixture("testdata/systemhooks/group_rename.json")},
	}
	for _, tc := range tests {
		t.Run(tc.event, func(t *testing.T) {
			parsedEvent, err := ParseSystemhook(tc.payload)
			if err != nil {
				t.Errorf("Error parsing build hook: %s", err)
			}
			event, ok := parsedEvent.(*GroupSystemEvent)
			if !ok {
				t.Errorf("Expected GroupSystemHookEvent, but parsing produced %T", parsedEvent)
			}
			assert.Equal(t, tc.event, event.EventName)
		})
	}
}

func TestParseSystemhookUser(t *testing.T) {
	var tests = []struct {
		event   string
		payload []byte
	}{
		{"user_create", loadFixture("testdata/systemhooks/user_create.json")},
		{"user_destroy", loadFixture("testdata/systemhooks/user_destroy.json")},
		{"user_rename", loadFixture("testdata/systemhooks/user_rename.json")},
	}
	for _, tc := range tests {
		t.Run(tc.event, func(t *testing.T) {
			parsedEvent, err := ParseSystemhook(tc.payload)
			if err != nil {
				t.Errorf("Error parsing build hook: %s", err)
			}
			event, ok := parsedEvent.(*UserSystemEvent)
			if !ok {
				t.Errorf("Expected UserSystemHookEvent, but parsing produced %T", parsedEvent)
			}
			assert.Equal(t, tc.event, event.EventName)
		})
	}
}

func TestParseSystemhookUserGroup(t *testing.T) {
	var tests = []struct {
		event   string
		payload []byte
	}{
		{"user_add_to_group", loadFixture("testdata/systemhooks/user_add_to_group.json")},
		{"user_remove_from_group", loadFixture("testdata/systemhooks/user_remove_from_group.json")},
		{"user_update_for_group", loadFixture("testdata/systemhooks/user_update_for_group.json")},
	}
	for _, tc := range tests {
		t.Run(tc.event, func(t *testing.T) {
			parsedEvent, err := ParseSystemhook(tc.payload)
			if err != nil {
				t.Errorf("Error parsing build hook: %s", err)
			}
			event, ok := parsedEvent.(*UserGroupSystemEvent)
			if !ok {
				t.Errorf("Expected UserGroupSystemHookEvent, but parsing produced %T", parsedEvent)
			}
			assert.Equal(t, tc.event, event.EventName)
		})
	}
}

func TestParseSystemhookUserTeam(t *testing.T) {
	var tests = []struct {
		event   string
		payload []byte
	}{
		{"user_add_to_team", loadFixture("testdata/systemhooks/user_add_to_team.json")},
		{"user_remove_from_team", loadFixture("testdata/systemhooks/user_remove_from_team.json")},
		{"user_update_for_team", loadFixture("testdata/systemhooks/user_update_for_team.json")},
	}
	for _, tc := range tests {
		t.Run(tc.event, func(t *testing.T) {
			parsedEvent, err := ParseSystemhook(tc.payload)
			if err != nil {
				t.Errorf("Error parsing build hook: %s", err)
			}
			event, ok := parsedEvent.(*UserTeamSystemEvent)
			if !ok {
				t.Errorf("Expected UserTeamSystemHookEvent, but parsing produced %T", parsedEvent)
			}
			assert.Equal(t, tc.event, event.EventName)
		})
	}
}

func TestParseHookSystemHook(t *testing.T) {
	parsedEvent1, err := ParseHook("System Hook", loadFixture("testdata/systemhooks/merge_request.json"))
	if err != nil {
		t.Errorf("Error parsing build hook: %s", err)
	}
	parsedEvent2, err := ParseSystemhook(loadFixture("testdata/systemhooks/merge_request.json"))
	if err != nil {
		t.Errorf("Error parsing build hook: %s", err)
	}
	assert.Equal(t, parsedEvent1, parsedEvent2)
}
