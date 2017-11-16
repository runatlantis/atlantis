package gitlab

import (
	"encoding/json"
	"testing"
)

func TestPushEventUnmarshal(t *testing.T) {
	jsonObject := `{
  "object_kind": "push",
  "before": "95790bf891e76fee5e1747ab589903a6a1f80f22",
  "after": "da1560886d4f094c3e6c9ef40349f7d38b5d27d7",
  "ref": "refs/heads/master",
  "checkout_sha": "da1560886d4f094c3e6c9ef40349f7d38b5d27d7",
  "user_id": 4,
  "user_name": "John Smith",
  "user_email": "john@example.com",
  "user_avatar": "https://s.gravatar.com/avatar/d4c74594d841139328695756648b6bd6?s=8://s.gravatar.com/avatar/d4c74594d841139328695756648b6bd6?s=80",
  "project_id": 15,
  "project":{
    "name":"Diaspora",
    "description":"",
    "web_url":"http://example.com/mike/diaspora",
    "avatar_url":null,
    "git_ssh_url":"git@example.com:mike/diaspora.git",
    "git_http_url":"http://example.com/mike/diaspora.git",
    "namespace":"Mike",
    "visibility":"public",
    "path_with_namespace":"mike/diaspora",
    "default_branch":"master",
    "homepage":"http://example.com/mike/diaspora",
    "url":"git@example.com:mike/diaspora.git",
    "ssh_url":"git@example.com:mike/diaspora.git",
    "http_url":"http://example.com/mike/diaspora.git"
  },
  "repository":{
    "name": "Diaspora",
    "url": "git@example.com:mike/diaspora.git",
    "description": "",
    "homepage": "http://example.com/mike/diaspora",
    "git_http_url":"http://example.com/mike/diaspora.git",
    "git_ssh_url":"git@example.com:mike/diaspora.git",
    "visibility":"public"
  },
  "commits": [
    {
      "id": "b6568db1bc1dcd7f8b4d5a946b0b91f9dacd7327",
      "message": "Update Catalan translation to e38cb41.",
      "timestamp": "2011-12-12T14:27:31+02:00",
      "url": "http://example.com/mike/diaspora/commit/b6568db1bc1dcd7f8b4d5a946b0b91f9dacd7327",
      "author": {
        "name": "Jordi Mallach",
        "email": "jordi@softcatala.org"
      },
      "added": ["CHANGELOG"],
      "modified": ["app/controller/application.rb"],
      "removed": []
    },
    {
      "id": "da1560886d4f094c3e6c9ef40349f7d38b5d27d7",
      "message": "fixed readme",
      "timestamp": "2012-01-03T23:36:29+02:00",
      "url": "http://example.com/mike/diaspora/commit/da1560886d4f094c3e6c9ef40349f7d38b5d27d7",
      "author": {
        "name": "GitLab dev user",
        "email": "gitlabdev@dv6700.(none)"
      },
      "added": ["CHANGELOG"],
      "modified": ["app/controller/application.rb"],
      "removed": []
    }
  ],
  "total_commits_count": 4
}`
	var event *PushEvent
	err := json.Unmarshal([]byte(jsonObject), &event)

	if err != nil {
		t.Errorf("Push Event can not unmarshaled: %v\n ", err.Error())
	}

	if event == nil {
		t.Errorf("Push Event is null")
	}

	if event.ProjectID != 15 {
		t.Errorf("ProjectID is %v, want %v", event.ProjectID, 15)
	}

}

func TestMergeEventUnmarshal(t *testing.T) {

	jsonObject := `{
  "object_kind": "merge_request",
  "user": {
    "name": "Administrator",
    "username": "root",
    "avatar_url": "http://www.gravatar.com/avatar/e64c7d89f26bd1972efa854d13d7dd61?s=40\u0026d=identicon"
  },
  "object_attributes": {
    "id": 99,
    "target_branch": "master",
    "source_branch": "ms-viewport",
    "source_project_id": 14,
    "author_id": 51,
    "assignee_id": 6,
    "title": "MS-Viewport",
    "created_at": "2013-12-03T17:23:34Z",
    "updated_at": "2013-12-03T17:23:34Z",
    "st_commits": null,
    "st_diffs": null,
    "milestone_id": null,
    "state": "opened",
    "merge_status": "unchecked",
    "target_project_id": 14,
    "iid": 1,
    "description": "",
    "source":{
      "name":"Awesome Project",
      "description":"Aut reprehenderit ut est.",
      "web_url":"http://example.com/awesome_space/awesome_project",
      "avatar_url":null,
      "git_ssh_url":"git@example.com:awesome_space/awesome_project.git",
      "git_http_url":"http://example.com/awesome_space/awesome_project.git",
      "namespace":"Awesome Space",
      "visibility":"private",
      "path_with_namespace":"awesome_space/awesome_project",
      "default_branch":"master",
      "homepage":"http://example.com/awesome_space/awesome_project",
      "url":"http://example.com/awesome_space/awesome_project.git",
      "ssh_url":"git@example.com:awesome_space/awesome_project.git",
      "http_url":"http://example.com/awesome_space/awesome_project.git"
    },
    "target": {
      "name":"Awesome Project",
      "description":"Aut reprehenderit ut est.",
      "web_url":"http://example.com/awesome_space/awesome_project",
      "avatar_url":null,
      "git_ssh_url":"git@example.com:awesome_space/awesome_project.git",
      "git_http_url":"http://example.com/awesome_space/awesome_project.git",
      "namespace":"Awesome Space",
      "visibility":"private",
      "path_with_namespace":"awesome_space/awesome_project",
      "default_branch":"master",
      "homepage":"http://example.com/awesome_space/awesome_project",
      "url":"http://example.com/awesome_space/awesome_project.git",
      "ssh_url":"git@example.com:awesome_space/awesome_project.git",
      "http_url":"http://example.com/awesome_space/awesome_project.git"
    },
    "last_commit": {
      "id": "da1560886d4f094c3e6c9ef40349f7d38b5d27d7",
      "message": "fixed readme",
      "timestamp": "2012-01-03T23:36:29+02:00",
      "url": "http://example.com/awesome_space/awesome_project/commits/da1560886d4f094c3e6c9ef40349f7d38b5d27d7",
      "author": {
        "name": "GitLab dev user",
        "email": "gitlabdev@dv6700.(none)"
      }
    },
    "work_in_progress": false,
    "url": "http://example.com/diaspora/merge_requests/1",
    "action": "open",
    "assignee": {
      "name": "User1",
      "username": "user1",
      "avatar_url": "http://www.gravatar.com/avatar/e64c7d89f26bd1972efa854d13d7dd61?s=40\u0026d=identicon"
    }
  }
}`

	var event *MergeEvent
	err := json.Unmarshal([]byte(jsonObject), &event)

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

}

func TestPipelineEventUnmarshal(t *testing.T) {
	jsonObject := `{
   "object_kind": "pipeline",
   "object_attributes":{
      "id": 31,
      "ref": "master",
      "tag": false,
      "sha": "bcbb5ec396a2c0f828686f14fac9b80b780504f2",
      "before_sha": "bcbb5ec396a2c0f828686f14fac9b80b780504f2",
      "status": "success",
      "stages":[
         "build",
         "test",
         "deploy"
      ],
      "created_at": "2016-08-12 15:23:28 UTC",
      "finished_at": "2016-08-12 15:26:29 UTC",
      "duration": 63
   },
   "user":{
      "name": "Administrator",
      "username": "root",
      "avatar_url": "http://www.gravatar.com/avatar/e32bd13e2add097461cb96824b7a829c?s=80\u0026d=identicon"
   },
   "project":{
      "name": "Gitlab Test",
      "description": "Atque in sunt eos similique dolores voluptatem.",
      "web_url": "http://192.168.64.1:3005/gitlab-org/gitlab-test",
      "avatar_url": null,
      "git_ssh_url": "git@192.168.64.1:gitlab-org/gitlab-test.git",
      "git_http_url": "http://192.168.64.1:3005/gitlab-org/gitlab-test.git",
      "namespace": "Gitlab Org",
      "visibility": "private",
      "path_with_namespace": "gitlab-org/gitlab-test",
      "default_branch": "master"
   },
   "commit":{
      "id": "bcbb5ec396a2c0f828686f14fac9b80b780504f2",
      "message": "test\n",
      "timestamp": "2016-08-12T17:23:21+02:00",
      "url": "http://example.com/gitlab-org/gitlab-test/commit/bcbb5ec396a2c0f828686f14fac9b80b780504f2",
      "author":{
         "name": "User",
         "email": "user@gitlab.com"
      }
   },
   "builds":[
      {
         "id": 380,
         "stage": "deploy",
         "name": "production",
         "status": "skipped",
         "created_at": "2016-08-12 15:23:28 UTC",
         "started_at": null,
         "finished_at": null,
         "when": "manual",
         "manual": true,
         "user":{
            "name": "Administrator",
            "username": "root",
            "avatar_url": "http://www.gravatar.com/avatar/e32bd13e2add097461cb96824b7a829c?s=80\u0026d=identicon"
         },
         "runner": null,
         "artifacts_file":{
            "filename": null,
            "size": null
         }
      },
      {
         "id": 377,
         "stage": "test",
         "name": "test-image",
         "status": "success",
         "created_at": "2016-08-12 15:23:28 UTC",
         "started_at": "2016-08-12 15:26:12 UTC",
         "finished_at": null,
         "when": "on_success",
         "manual": false,
         "user":{
            "name": "Administrator",
            "username": "root",
            "avatar_url": "http://www.gravatar.com/avatar/e32bd13e2add097461cb96824b7a829c?s=80\u0026d=identicon"
         },
         "runner": {
            "id": 6,
            "description": "Kubernetes Runner",
            "active": true,
            "is_shared": true
         },
         "artifacts_file":{
            "filename": "artifacts.zip",
            "size": 1319148
         }
      },
      {
         "id": 378,
         "stage": "test",
         "name": "test-build",
         "status": "success",
         "created_at": "2016-08-12 15:23:28 UTC",
         "started_at": "2016-08-12 15:26:12 UTC",
         "finished_at": "2016-08-12 15:26:29 UTC",
         "when": "on_success",
         "manual": false,
         "user":{
            "name": "Administrator",
            "username": "root",
            "avatar_url": "http://www.gravatar.com/avatar/e32bd13e2add097461cb96824b7a829c?s=80\u0026d=identicon"
         },
         "runner": null,
         "artifacts_file":{
            "filename": null,
            "size": null
         }
      },
      {
         "id": 376,
         "stage": "build",
         "name": "build-image",
         "status": "success",
         "created_at": "2016-08-12 15:23:28 UTC",
         "started_at": "2016-08-12 15:24:56 UTC",
         "finished_at": "2016-08-12 15:25:26 UTC",
         "when": "on_success",
         "manual": false,
         "user":{
            "name": "Administrator",
            "username": "root",
            "avatar_url": "http://www.gravatar.com/avatar/e32bd13e2add097461cb96824b7a829c?s=80\u0026d=identicon"
         },
         "runner": null,
         "artifacts_file":{
            "filename": null,
            "size": null
         }
      },
      {
         "id": 379,
         "stage": "deploy",
         "name": "staging",
         "status": "created",
         "created_at": "2016-08-12 15:23:28 UTC",
         "started_at": null,
         "finished_at": null,
         "when": "on_success",
         "manual": false,
         "user":{
            "name": "Administrator",
            "username": "root",
            "avatar_url": "http://www.gravatar.com/avatar/e32bd13e2add097461cb96824b7a829c?s=80\u0026d=identicon"
         },
         "runner": null,
         "artifacts_file":{
            "filename": null,
            "size": null
         }
      }
   ]
}`

	var event *PipelineEvent
	err := json.Unmarshal([]byte(jsonObject), &event)

	if err != nil {
		t.Errorf("Pipeline Event can not unmarshaled: %v\n ", err.Error())
	}

	if event == nil {
		t.Errorf("Pipeline Event is null")
	}

	if event.ObjectAttributes.ID != 31 {
		t.Errorf("ObjectAttributes is %v, want %v", event.ObjectAttributes.ID, 1977)
	}

}

func TestBuildEventUnmarshal(t *testing.T) {
	jsonObject := `{
  "object_kind": "build",
  "ref": "gitlab-script-trigger",
  "tag": false,
  "before_sha": "2293ada6b400935a1378653304eaf6221e0fdb8f",
  "sha": "2293ada6b400935a1378653304eaf6221e0fdb8f",
  "build_id": 1977,
  "build_name": "test",
  "build_stage": "test",
  "build_status": "created",
  "build_started_at": null,
  "build_finished_at": null,
  "build_duration": 23.265997,
  "build_allow_failure": false,
  "project_id": 380,
  "project_name": "gitlab-org/gitlab-test",
  "user": {
    "id": 3,
    "name": "User",
    "email": "user@gitlab.com"
  },
  "commit": {
    "id": 2366,
    "sha": "2293ada6b400935a1378653304eaf6221e0fdb8f",
    "message": "test\n",
    "author_name": "User",
    "author_email": "user@gitlab.com",
    "status": "created",
    "duration": null,
    "started_at": null,
    "finished_at": null
  },
  "repository": {
    "name": "gitlab_test",
    "git_ssh_url": "git@192.168.64.1:gitlab-org/gitlab-test.git",
    "description": "Atque in sunt eos similique dolores voluptatem.",
    "homepage": "http://192.168.64.1:3005/gitlab-org/gitlab-test",
    "git_ssh_url": "git@192.168.64.1:gitlab-org/gitlab-test.git",
    "git_http_url": "http://192.168.64.1:3005/gitlab-org/gitlab-test.git",
    "visibility": "private"
  }
}`
	var event *BuildEvent
	err := json.Unmarshal([]byte(jsonObject), &event)

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

	jsonObject := `{
	"object_kind": "merge_request",
	"user": {
		"name": "Administrator",
		"username": "root",
		"avatar_url": "http://www.gravatar.com/avatar/d22738dc40839e3d95fca77ca3eac067?s=80\u0026d=identicon"
	},
	"project": {
		"name": "example-project",
		"description": "",
		"web_url": "http://example.com/exm-namespace/example-project",
		"avatar_url": null,
		"git_ssh_url": "git@example.com:exm-namespace/example-project.git",
		"git_http_url": "http://example.com/exm-namespace/example-project.git",
		"namespace": "exm-namespace",
		"visibility": "public",
		"path_with_namespace": "exm-namespace/example-project",
		"default_branch": "master",
		"homepage": "http://example.com/exm-namespace/example-project",
		"url": "git@example.com:exm-namespace/example-project.git",
		"ssh_url": "git@example.com:exm-namespace/example-project.git",
		"http_url": "http://example.com/exm-namespace/example-project.git"
	},
	"object_attributes": {
		"id": 15917,
		"target_branch ": "master ",
		"source_branch ": "source-branch-test ",
		"source_project_id ": 87,
		"author_id ": 15,
		"assignee_id ": 29,
		"title ": "source-branch-test ",
		"created_at ": "2016 - 12 - 01 13: 11: 10 UTC ",
		"updated_at ": "2016 - 12 - 01 13: 21: 20 UTC ",
		"milestone_id ": null,
		"state ": "merged ",
		"merge_status ": "can_be_merged ",
		"target_project_id ": 87,
		"iid ": 1402,
		"description ": "word doc support for e - ticket ",
		"position ": 0,
		"locked_at ": null,
		"updated_by_id ": null,
		"merge_error ": null,
		"merge_params": {
			"force_remove_source_branch": "0"
		},
		"merge_when_build_succeeds": false,
		"merge_user_id": null,
		"merge_commit_sha": "ac3ca1559bc39abf963586372eff7f8fdded646e",
		"deleted_at": null,
		"approvals_before_merge": null,
		"rebase_commit_sha": null,
		"in_progress_merge_commit_sha": null,
		"lock_version": 0,
		"time_estimate": 0,
		"source": {
			"name": "example-project",
			"description": "",
			"web_url": "http://example.com/exm-namespace/example-project",
			"avatar_url": null,
			"git_ssh_url": "git@example.com:exm-namespace/example-project.git",
			"git_http_url": "http://example.com/exm-namespace/example-project.git",
			"namespace": "exm-namespace",
			"visibility": "public",
			"path_with_namespace": "exm-namespace/example-project",
			"default_branch": "master",
			"homepage": "http://example.com/exm-namespace/example-project",
			"url": "git@example.com:exm-namespace/example-project.git",
			"ssh_url": "git@example.com:exm-namespace/example-project.git",
			"http_url": "http://example.com/exm-namespace/example-project.git"
		},
		"target": {
			"name": "example-project",
			"description": "",
			"web_url": "http://example.com/exm-namespace/example-project",
			"avatar_url": null,
			"git_ssh_url": "git@example.com:exm-namespace/example-project.git",
			"git_http_url": "http://example.com/exm-namespace/example-project.git",
			"namespace": "exm-namespace",
			"visibility": "public",
			"path_with_namespace": "exm-namespace/example-project",
			"default_branch": "master",
			"homepage": "http://example.com/exm-namespace/example-project",
			"url": "git@example.com:exm-namespace/example-project.git",
			"ssh_url": "git@example.com:exm-namespace/example-project.git",
			"http_url": "http://example.com/exm-namespace/example-project.git"
		},
		"last_commit": {
			"id": "61b6a0d35dbaf915760233b637622e383d3cc9ec",
			"message": "commit message",
			"timestamp": "2016-12-01T15:07:53+02:00",
			"url": "http://example.com/exm-namespace/example-project/commit/61b6a0d35dbaf915760233b637622e383d3cc9ec",
			"author": {
				"name": "Test User",
				"email": "test.user@mail.com"
			}
		},
		"work_in_progress": false,
		"url": "http://example.com/exm-namespace/example-project/merge_requests/1402",
		"action": "merge"
	},
	"repository": {
		"name": "example-project",
		"url": "git@example.com:exm-namespace/example-project.git",
		"description": "",
		"homepage": "http://example.com/exm-namespace/example-project"
	},
	"assignee": {
		"name": "Administrator",
		"username": "root",
		"avatar_url": "http://www.gravatar.com/avatar/d22738dc40839e3d95fca77ca3eac067?s=80\u0026d=identicon"
	}
}`

	var event *MergeEvent
	err := json.Unmarshal([]byte(jsonObject), &event)

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
}
