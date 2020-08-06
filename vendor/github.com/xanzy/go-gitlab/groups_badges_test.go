package gitlab

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestListGroupBadges(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/groups/1/badges",
		func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, "GET")
			fmt.Fprint(w, `[{"id":1, "kind":"group"},{"id":2, "kind":"group"}]`)
		})

	badges, _, err := client.GroupBadges.ListGroupBadges(1, &ListGroupBadgesOptions{})
	if err != nil {
		t.Errorf("GroupBadges.ListGroupBadges returned error: %v", err)
	}

	want := []*GroupBadge{{ID: 1, Kind: GroupBadgeKind}, {ID: 2, Kind: GroupBadgeKind}}
	if !reflect.DeepEqual(want, badges) {
		t.Errorf("GroupBadges.ListGroupBadges returned %+v, want %+v", badges, want)
	}
}

func TestGetGroupBadge(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/groups/1/badges/2",
		func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, "GET")
			fmt.Fprint(w, `{"id":2, "kind":"group"}`)
		})

	badge, _, err := client.GroupBadges.GetGroupBadge(1, 2)
	if err != nil {
		t.Errorf("GroupBadges.GetGroupBadge returned error: %v", err)
	}

	want := &GroupBadge{ID: 2, Kind: GroupBadgeKind}
	if !reflect.DeepEqual(want, badge) {
		t.Errorf("GroupBadges.GetGroupBadge returned %+v, want %+v", badge, want)
	}
}

func TestAddGroupBadge(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/groups/1/badges",
		func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, "POST")
			fmt.Fprint(w, `{"id":3, "link_url":"LINK", "image_url":"IMAGE", "kind":"group"}`)
		})

	opt := &AddGroupBadgeOptions{ImageURL: String("IMAGE"), LinkURL: String("LINK")}
	badge, _, err := client.GroupBadges.AddGroupBadge(1, opt)
	if err != nil {
		t.Errorf("GroupBadges.AddGroupBadge returned error: %v", err)
	}

	want := &GroupBadge{ID: 3, ImageURL: "IMAGE", LinkURL: "LINK", Kind: GroupBadgeKind}
	if !reflect.DeepEqual(want, badge) {
		t.Errorf("GroupBadges.AddGroupBadge returned %+v, want %+v", badge, want)
	}
}

func TestEditGroupBadge(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/groups/1/badges/2",
		func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, "PUT")
			fmt.Fprint(w, `{"id":2, "link_url":"NEW_LINK", "image_url":"NEW_IMAGE", "kind":"group"}`)
		})

	opt := &EditGroupBadgeOptions{ImageURL: String("NEW_IMAGE"), LinkURL: String("NEW_LINK")}
	badge, _, err := client.GroupBadges.EditGroupBadge(1, 2, opt)
	if err != nil {
		t.Errorf("GroupBadges.EditGroupBadge returned error: %v", err)
	}

	want := &GroupBadge{ID: 2, ImageURL: "NEW_IMAGE", LinkURL: "NEW_LINK", Kind: GroupBadgeKind}
	if !reflect.DeepEqual(want, badge) {
		t.Errorf("GroupBadges.EditGroupBadge returned %+v, want %+v", badge, want)
	}
}

func TestRemoveGroupBadge(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/groups/1/badges/2",
		func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, "DELETE")
			w.WriteHeader(http.StatusAccepted)
		},
	)

	resp, err := client.GroupBadges.DeleteGroupBadge(1, 2)
	if err != nil {
		t.Errorf("GroupBadges.DeleteGroupBadge returned error: %v", err)
	}

	want := http.StatusAccepted
	got := resp.StatusCode
	if got != want {
		t.Errorf("GroupsBadges.DeleteGroupBadge returned %d, want %d", got, want)
	}

}
