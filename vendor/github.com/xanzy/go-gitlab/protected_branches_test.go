package gitlab

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestListProtectedBranches(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/protected_branches", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `[{"id":1,
                              "name":"master",
                              "push_access_levels":[{
                                "access_level":40,
                                "access_level_description":"Maintainers"
                              }],
                              "merge_access_levels":[{
                                "access_level":40,
                                "access_level_description":"Maintainers"
                               }],
                              "code_owner_approval_required":false
                            }]`)
	})
	opt := &ListProtectedBranchesOptions{}
	protectedBranches, _, err := client.ProtectedBranches.ListProtectedBranches("1", opt)
	if err != nil {
		t.Errorf("ProtectedBranches.ListProtectedBranches returned error: %v", err)
	}
	want := []*ProtectedBranch{
		{
			ID:   1,
			Name: "master",
			PushAccessLevels: []*BranchAccessDescription{
				{
					AccessLevel:            40,
					AccessLevelDescription: "Maintainers",
				},
			},
			MergeAccessLevels: []*BranchAccessDescription{
				{
					AccessLevel:            40,
					AccessLevelDescription: "Maintainers",
				},
			},
			CodeOwnerApprovalRequired: false,
		},
	}
	if !reflect.DeepEqual(want, protectedBranches) {
		t.Errorf("ProtectedBranches.ListProtectedBranches returned %+v, want %+v", protectedBranches, want)
	}
}

func TestListProtectedBranchesWithoutCodeOwnerApproval(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/protected_branches", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `[{"id":1,
                             "name":"master",
                             "push_access_levels":[{
                               "access_level":40,
                               "access_level_description":"Maintainers"
                             }],
                             "merge_access_levels":[{
                               "access_level":40,
                               "access_level_description":"Maintainers"
                             }]
                           }]`)
	})
	opt := &ListProtectedBranchesOptions{}
	protectedBranches, _, err := client.ProtectedBranches.ListProtectedBranches("1", opt)
	if err != nil {
		t.Errorf("ProtectedBranches.ListProtectedBranches returned error: %v", err)
	}
	want := []*ProtectedBranch{
		{
			ID:   1,
			Name: "master",
			PushAccessLevels: []*BranchAccessDescription{
				{
					AccessLevel:            40,
					AccessLevelDescription: "Maintainers",
				},
			},
			MergeAccessLevels: []*BranchAccessDescription{
				{
					AccessLevel:            40,
					AccessLevelDescription: "Maintainers",
				},
			},
			CodeOwnerApprovalRequired: false,
		},
	}
	if !reflect.DeepEqual(want, protectedBranches) {
		t.Errorf("Projects.ListProjects returned %+v, want %+v", protectedBranches, want)
	}
}

func TestProtectRepositoryBranches(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/protected_branches", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		fmt.Fprint(w, `{"id":1,
                             "name":"master",
                             "push_access_levels":[{
                               "access_level":40,
                               "access_level_description":"Maintainers"
                             }],
                             "merge_access_levels":[{
                               "access_level":40,
                               "access_level_description":"Maintainers"
                             }],
                             "code_owner_approval_required":true
                            }`)
	})
	opt := &ProtectRepositoryBranchesOptions{
		Name:                      String("master"),
		PushAccessLevel:           AccessLevel(MaintainerPermissions),
		MergeAccessLevel:          AccessLevel(MaintainerPermissions),
		CodeOwnerApprovalRequired: Bool(true),
	}
	projects, _, err := client.ProtectedBranches.ProtectRepositoryBranches("1", opt)
	if err != nil {
		t.Errorf("ProtectedBranches.ProtectRepositoryBranches returned error: %v", err)
	}
	want := &ProtectedBranch{
		ID:   1,
		Name: "master",
		PushAccessLevels: []*BranchAccessDescription{
			{
				AccessLevel:            40,
				AccessLevelDescription: "Maintainers",
			},
		},
		MergeAccessLevels: []*BranchAccessDescription{
			{
				AccessLevel:            40,
				AccessLevelDescription: "Maintainers",
			},
		},
		CodeOwnerApprovalRequired: true,
	}
	if !reflect.DeepEqual(want, projects) {
		t.Errorf("Projects.ListProjects returned %+v, want %+v", projects, want)
	}
}

func TestUpdateRepositoryBranches(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/protected_branches/master", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PATCH")
		codeApprovalQueryParam := r.URL.Query().Get("code_owner_approval_required")
		if codeApprovalQueryParam != "true" {
			t.Errorf("query param code_owner_approval_required should be true but was %s", codeApprovalQueryParam)
		}
	})
	opt := &RequireCodeOwnerApprovalsOptions{
		CodeOwnerApprovalRequired: Bool(true),
	}
	_, err := client.ProtectedBranches.RequireCodeOwnerApprovals("1", "master", opt)
	if err != nil {
		t.Errorf("ProtectedBranches.UpdateRepositoryBranchesOptions returned error: %v", err)
	}
}
