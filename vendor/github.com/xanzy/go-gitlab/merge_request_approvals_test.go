package gitlab

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestGetApprovalState(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/merge_requests/1/approval_state", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{
			"approval_rules_overwritten": true,
			"rules": [
			{
				"id": 1,
				"name": "security",
				"rule_type": "regular",
				"eligible_approvers": [
					{
						"id": 5,
						"name": "John Doe",
						"username": "jdoe",
						"state": "active",
						"avatar_url": "https://www.gravatar.com/avatar/0?s=80&d=identicon",
						"web_url": "http://localhost/jdoe"
					},
					{
						"id": 50,
						"name": "Group Member 1",
						"username": "group_member_1",
						"state": "active",
						"avatar_url": "https://www.gravatar.com/avatar/0?s=80&d=identicon",
						"web_url": "http://localhost/group_member_1"
					}
				],
				"approvals_required": 3,
				"source_rule": null,
				"users": [
					{
						"id": 5,
						"name": "John Doe",
						"username": "jdoe",
						"state": "active",
						"avatar_url": "https://www.gravatar.com/avatar/0?s=80&d=identicon",
						"web_url": "http://localhost/jdoe"
					}
				],
				"groups": [
					{
						"id": 5,
						"name": "group1",
						"path": "group1",
						"description": "",
						"visibility": "public",
						"lfs_enabled": false,
						"avatar_url": null,
						"web_url": "http://localhost/groups/group1",
						"request_access_enabled": false,
						"full_name": "group1",
						"full_path": "group1",
						"parent_id": null,
						"ldap_cn": null,
						"ldap_access": null
					}
				],
				"contains_hidden_groups": false,
				"approved_by": [
					{
						"id": 5,
						"name": "John Doe",
						"username": "jdoe",
						"state": "active",
						"avatar_url": "https://www.gravatar.com/avatar/0?s=80&d=identicon",
						"web_url": "http://localhost/jdoe"
					}
				],
				"approved": false
			}
		]
		}`)
	})

	approvals, _, err := client.MergeRequestApprovals.GetApprovalState(1, 1)
	if err != nil {
		t.Errorf("MergeRequestApprovals.GetApprovalState returned error: %v", err)
	}

	want := &MergeRequestApprovalState{
		ApprovalRulesOverwritten: true,
		Rules: []*MergeRequestApprovalRule{
			{
				ID:       1,
				Name:     "security",
				RuleType: "regular",
				EligibleApprovers: []*BasicUser{
					{
						ID:        5,
						Name:      "John Doe",
						Username:  "jdoe",
						State:     "active",
						AvatarURL: "https://www.gravatar.com/avatar/0?s=80&d=identicon",
						WebURL:    "http://localhost/jdoe",
					},
					{
						ID:        50,
						Name:      "Group Member 1",
						Username:  "group_member_1",
						State:     "active",
						AvatarURL: "https://www.gravatar.com/avatar/0?s=80&d=identicon",
						WebURL:    "http://localhost/group_member_1",
					},
				},
				ApprovalsRequired: 3,
				Users: []*BasicUser{
					{
						ID:        5,
						Name:      "John Doe",
						Username:  "jdoe",
						State:     "active",
						AvatarURL: "https://www.gravatar.com/avatar/0?s=80&d=identicon",
						WebURL:    "http://localhost/jdoe",
					},
				},
				Groups: []*Group{
					{
						ID:                   5,
						Name:                 "group1",
						Path:                 "group1",
						Description:          "",
						Visibility:           PublicVisibility,
						LFSEnabled:           false,
						AvatarURL:            "",
						WebURL:               "http://localhost/groups/group1",
						RequestAccessEnabled: false,
						FullName:             "group1",
						FullPath:             "group1",
					},
				},
				ApprovedBy: []*BasicUser{
					{
						ID:        5,
						Name:      "John Doe",
						Username:  "jdoe",
						State:     "active",
						AvatarURL: "https://www.gravatar.com/avatar/0?s=80&d=identicon",
						WebURL:    "http://localhost/jdoe",
					},
				},
				Approved: false,
			},
		},
	}

	if !reflect.DeepEqual(want, approvals) {
		t.Errorf("MergeRequestApprovals.GetApprovalState returned %+v, want %+v", approvals, want)
	}
}

func TestGetApprovalRules(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/merge_requests/1/approval_rules", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `[
			{
				"id": 1,
				"name": "security",
				"rule_type": "regular",
				"eligible_approvers": [
					{
						"id": 5,
						"name": "John Doe",
						"username": "jdoe",
						"state": "active",
						"avatar_url": "https://www.gravatar.com/avatar/0?s=80&d=identicon",
						"web_url": "http://localhost/jdoe"
					},
					{
						"id": 50,
						"name": "Group Member 1",
						"username": "group_member_1",
						"state": "active",
						"avatar_url": "https://www.gravatar.com/avatar/0?s=80&d=identicon",
						"web_url": "http://localhost/group_member_1"
					}
				],
				"approvals_required": 3,
				"source_rule": null,
				"users": [
					{
						"id": 5,
						"name": "John Doe",
						"username": "jdoe",
						"state": "active",
						"avatar_url": "https://www.gravatar.com/avatar/0?s=80&d=identicon",
						"web_url": "http://localhost/jdoe"
					}
				],
				"groups": [
					{
						"id": 5,
						"name": "group1",
						"path": "group1",
						"description": "",
						"visibility": "public",
						"lfs_enabled": false,
						"avatar_url": null,
						"web_url": "http://localhost/groups/group1",
						"request_access_enabled": false,
						"full_name": "group1",
						"full_path": "group1",
						"parent_id": null,
						"ldap_cn": null,
						"ldap_access": null
					}
				],
				"contains_hidden_groups": false
			}
		]`)
	})

	approvals, _, err := client.MergeRequestApprovals.GetApprovalRules(1, 1)
	if err != nil {
		t.Errorf("MergeRequestApprovals.GetApprovalRules returned error: %v", err)
	}

	want := []*MergeRequestApprovalRule{
		{
			ID:       1,
			Name:     "security",
			RuleType: "regular",
			EligibleApprovers: []*BasicUser{
				{
					ID:        5,
					Name:      "John Doe",
					Username:  "jdoe",
					State:     "active",
					AvatarURL: "https://www.gravatar.com/avatar/0?s=80&d=identicon",
					WebURL:    "http://localhost/jdoe",
				},
				{
					ID:        50,
					Name:      "Group Member 1",
					Username:  "group_member_1",
					State:     "active",
					AvatarURL: "https://www.gravatar.com/avatar/0?s=80&d=identicon",
					WebURL:    "http://localhost/group_member_1",
				},
			},
			ApprovalsRequired: 3,
			Users: []*BasicUser{
				{
					ID:        5,
					Name:      "John Doe",
					Username:  "jdoe",
					State:     "active",
					AvatarURL: "https://www.gravatar.com/avatar/0?s=80&d=identicon",
					WebURL:    "http://localhost/jdoe",
				},
			},
			Groups: []*Group{
				{
					ID:                   5,
					Name:                 "group1",
					Path:                 "group1",
					Description:          "",
					Visibility:           PublicVisibility,
					LFSEnabled:           false,
					AvatarURL:            "",
					WebURL:               "http://localhost/groups/group1",
					RequestAccessEnabled: false,
					FullName:             "group1",
					FullPath:             "group1",
				},
			},
		},
	}

	if !reflect.DeepEqual(want, approvals) {
		t.Errorf("MergeRequestApprovals.GetApprovalRules returned %+v, want %+v", approvals, want)
	}
}

func TestCreateApprovalRules(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/merge_requests/1/approval_rules", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		fmt.Fprint(w, `{
			"id": 1,
			"name": "security",
			"rule_type": "regular",
			"eligible_approvers": [
				{
					"id": 5,
					"name": "John Doe",
					"username": "jdoe",
					"state": "active",
					"avatar_url": "https://www.gravatar.com/avatar/0?s=80&d=identicon",
					"web_url": "http://localhost/jdoe"
				},
				{
					"id": 50,
					"name": "Group Member 1",
					"username": "group_member_1",
					"state": "active",
					"avatar_url": "https://www.gravatar.com/avatar/0?s=80&d=identicon",
					"web_url": "http://localhost/group_member_1"
				}
			],
			"approvals_required": 3,
			"source_rule": null,
			"users": [
				{
					"id": 5,
					"name": "John Doe",
					"username": "jdoe",
					"state": "active",
					"avatar_url": "https://www.gravatar.com/avatar/0?s=80&d=identicon",
					"web_url": "http://localhost/jdoe"
				}
			],
			"groups": [
				{
					"id": 5,
					"name": "group1",
					"path": "group1",
					"description": "",
					"visibility": "public",
					"lfs_enabled": false,
					"avatar_url": null,
					"web_url": "http://localhost/groups/group1",
					"request_access_enabled": false,
					"full_name": "group1",
					"full_path": "group1",
					"parent_id": null,
					"ldap_cn": null,
					"ldap_access": null
				}
			],
			"contains_hidden_groups": false
		}`)
	})

	opt := &CreateMergeRequestApprovalRuleOptions{
		Name:              String("security"),
		ApprovalsRequired: Int(3),
		UserIDs:           []int{5, 50},
		GroupIDs:          []int{5},
	}

	rule, _, err := client.MergeRequestApprovals.CreateApprovalRule(1, 1, opt)
	if err != nil {
		t.Errorf("MergeRequestApprovals.CreateApprovalRule returned error: %v", err)
	}

	want := &MergeRequestApprovalRule{
		ID:       1,
		Name:     "security",
		RuleType: "regular",
		EligibleApprovers: []*BasicUser{
			{
				ID:        5,
				Name:      "John Doe",
				Username:  "jdoe",
				State:     "active",
				AvatarURL: "https://www.gravatar.com/avatar/0?s=80&d=identicon",
				WebURL:    "http://localhost/jdoe",
			},
			{
				ID:        50,
				Name:      "Group Member 1",
				Username:  "group_member_1",
				State:     "active",
				AvatarURL: "https://www.gravatar.com/avatar/0?s=80&d=identicon",
				WebURL:    "http://localhost/group_member_1",
			},
		},
		ApprovalsRequired: 3,
		Users: []*BasicUser{
			{
				ID:        5,
				Name:      "John Doe",
				Username:  "jdoe",
				State:     "active",
				AvatarURL: "https://www.gravatar.com/avatar/0?s=80&d=identicon",
				WebURL:    "http://localhost/jdoe",
			},
		},
		Groups: []*Group{
			{
				ID:                   5,
				Name:                 "group1",
				Path:                 "group1",
				Description:          "",
				Visibility:           PublicVisibility,
				LFSEnabled:           false,
				AvatarURL:            "",
				WebURL:               "http://localhost/groups/group1",
				RequestAccessEnabled: false,
				FullName:             "group1",
				FullPath:             "group1",
			},
		},
	}

	if !reflect.DeepEqual(want, rule) {
		t.Errorf("MergeRequestApprovals.CreateApprovalRule returned %+v, want %+v", rule, want)
	}
}
