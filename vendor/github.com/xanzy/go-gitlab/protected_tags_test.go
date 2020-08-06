package gitlab

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListProtectedTags(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/protected_tags", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `[{"name":"1.0.0", "create_access_levels": [{"access_level": 40, "access_level_description": "Maintainers"}]},{"name":"*-release", "create_access_levels": [{"access_level": 30, "access_level_description": "Developers + Maintainers"}]}]`)
	})

	expected := []*ProtectedTag{
		{
			Name: "1.0.0",
			CreateAccessLevels: []*TagAccessDescription{
				{
					AccessLevel:            40,
					AccessLevelDescription: "Maintainers",
				},
			},
		},
		{
			Name: "*-release",
			CreateAccessLevels: []*TagAccessDescription{
				{
					AccessLevel:            30,
					AccessLevelDescription: "Developers + Maintainers",
				},
			},
		},
	}

	opt := &ListProtectedTagsOptions{}
	tags, _, err := client.ProtectedTags.ListProtectedTags(1, opt)
	assert.NoError(t, err, "failed to get response")
	assert.Equal(t, expected, tags)
}

func TestGetProtectedTag(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	tagName := "my-awesome-tag"

	mux.HandleFunc(fmt.Sprintf("/api/v4/projects/1/protected_tags/%s", tagName), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"name":"my-awesome-tag", "create_access_levels": [{"access_level": 30, "access_level_description": "Developers + Maintainers"}]}`)
	})

	expected := &ProtectedTag{
		Name: tagName,
		CreateAccessLevels: []*TagAccessDescription{
			{
				AccessLevel:            30,
				AccessLevelDescription: "Developers + Maintainers",
			},
		},
	}

	tag, _, err := client.ProtectedTags.GetProtectedTag(1, tagName)

	assert.NoError(t, err, "failed to get response")
	assert.Equal(t, expected, tag)
}

func TestProtectRepositoryTags(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/protected_tags", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		fmt.Fprint(w, `{"name":"my-awesome-tag", "create_access_levels": [{"access_level": 30, "access_level_description": "Developers + Maintainers"}]}`)
	})

	expected := &ProtectedTag{
		Name: "my-awesome-tag",
		CreateAccessLevels: []*TagAccessDescription{
			{
				AccessLevel:            30,
				AccessLevelDescription: "Developers + Maintainers",
			},
		},
	}

	opt := &ProtectRepositoryTagsOptions{Name: String("my-awesome-tag"), CreateAccessLevel: AccessLevel(30)}
	tag, _, err := client.ProtectedTags.ProtectRepositoryTags(1, opt)

	assert.NoError(t, err, "failed to get response")
	assert.Equal(t, expected, tag)
}

func TestUnprotectRepositoryTags(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/protected_tags/my-awesome-tag", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
	})

	resp, err := client.ProtectedTags.UnprotectRepositoryTags(1, "my-awesome-tag")
	assert.NoError(t, err, "failed to get response")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
