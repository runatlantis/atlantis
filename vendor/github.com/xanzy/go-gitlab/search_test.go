package gitlab

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSearchService_Users(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/search", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		mustWriteHTTPResponse(t, w, "testdata/search_users.json")
	})

	opts := &SearchOptions{PerPage: 2}
	users, _, err := client.Search.Users("doe", opts)

	require.NoError(t, err)

	want := []*User{{
		ID:        1,
		Username:  "user1",
		Name:      "John Doe1",
		State:     "active",
		AvatarURL: "http://www.gravatar.com/avatar/c922747a93b40d1ea88262bf1aebee62?s=80&d=identicon",
	}}
	require.Equal(t, want, users)
}

func TestSearchService_UsersByGroup(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/groups/3/-/search", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		mustWriteHTTPResponse(t, w, "testdata/search_users.json")
	})

	opts := &SearchOptions{PerPage: 2}
	users, _, err := client.Search.UsersByGroup("3", "doe", opts)

	require.NoError(t, err)

	want := []*User{{
		ID:        1,
		Username:  "user1",
		Name:      "John Doe1",
		State:     "active",
		AvatarURL: "http://www.gravatar.com/avatar/c922747a93b40d1ea88262bf1aebee62?s=80&d=identicon",
	}}
	require.Equal(t, want, users)
}

func TestSearchService_UsersByProject(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/6/-/search", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		mustWriteHTTPResponse(t, w, "testdata/search_users.json")
	})

	opts := &SearchOptions{PerPage: 2}
	users, _, err := client.Search.UsersByProject("6", "doe", opts)

	require.NoError(t, err)

	want := []*User{{
		ID:        1,
		Username:  "user1",
		Name:      "John Doe1",
		State:     "active",
		AvatarURL: "http://www.gravatar.com/avatar/c922747a93b40d1ea88262bf1aebee62?s=80&d=identicon",
	}}
	require.Equal(t, want, users)
}
