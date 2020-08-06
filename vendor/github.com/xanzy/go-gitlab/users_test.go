package gitlab

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestBlockUser(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	path := fmt.Sprintf("/%susers/1/block", apiVersionPath)
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		w.WriteHeader(http.StatusCreated)
	})

	err := client.Users.BlockUser(1)
	if err != nil {
		t.Errorf("Users.BlockUser returned error: %v", err)
	}
}

func TestBlockUser_UserNotFound(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	path := fmt.Sprintf("/%susers/1/block", apiVersionPath)
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		w.WriteHeader(http.StatusNotFound)
	})

	err := client.Users.BlockUser(1)
	if err != ErrUserNotFound {
		t.Errorf("Users.BlockUser error.\nExpected: %+v\nGot: %+v", ErrUserNotFound, err)
	}
}

func TestBlockUser_BlockPrevented(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	path := fmt.Sprintf("/%susers/1/block", apiVersionPath)
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		w.WriteHeader(http.StatusForbidden)
	})

	err := client.Users.BlockUser(1)
	if err != ErrUserBlockPrevented {
		t.Errorf("Users.BlockUser error.\nExpected: %+v\nGot: %+v", ErrUserBlockPrevented, err)
	}
}

func TestBlockUser_UnknownError(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	path := fmt.Sprintf("/%susers/1/block", apiVersionPath)
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		w.WriteHeader(http.StatusTeapot)
	})

	want := fmt.Sprintf("Received unexpected result code: %d", http.StatusTeapot)

	err := client.Users.BlockUser(1)
	if err.Error() != want {
		t.Errorf("Users.BlockUser error.\nExpected: %s\nGot: %v", want, err)
	}
}

func TestBlockUser_BadResponseFromNet(t *testing.T) {
	client := NewClient(nil, "")
	client.SetBaseURL("")

	want := "unsupported protocol scheme"

	err := client.Users.BlockUser(1)
	if !strings.Contains(err.Error(), want) {
		t.Errorf("Users.BlockUser error.\nExpected error to contain: %s\nGot: %v", want, err)
	}
}

//  ------------------------  Unblock user -------------------------
func TestUnblockUser(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	path := fmt.Sprintf("/%susers/1/unblock", apiVersionPath)
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		w.WriteHeader(http.StatusCreated)
	})

	err := client.Users.UnblockUser(1)
	if err != nil {
		t.Errorf("Users.UnblockUser returned error: %v", err)
	}
}

func TestUnblockUser_UserNotFound(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	path := fmt.Sprintf("/%susers/1/unblock", apiVersionPath)
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		w.WriteHeader(http.StatusNotFound)
	})

	err := client.Users.UnblockUser(1)
	if err != ErrUserNotFound {
		t.Errorf("Users.UnblockUser error.\nExpected: %v\nGot: %v", ErrUserNotFound, err)
	}
}

func TestUnblockUser_UnblockPrevented(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	path := fmt.Sprintf("/%susers/1/unblock", apiVersionPath)
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		w.WriteHeader(http.StatusForbidden)
	})

	err := client.Users.UnblockUser(1)
	if err != ErrUserUnblockPrevented {
		t.Errorf("Users.UnblockUser error.\nExpected: %v\nGot: %v", ErrUserUnblockPrevented, err)
	}
}

func TestUnblockUser_UnknownError(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	path := fmt.Sprintf("/%susers/1/unblock", apiVersionPath)
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		w.WriteHeader(http.StatusTeapot)
	})

	want := fmt.Sprintf("Received unexpected result code: %d", http.StatusTeapot)

	err := client.Users.UnblockUser(1)
	if err.Error() != want {
		t.Errorf("Users.UnblockUser error.\nExpected: %s\n\tGot: %v", want, err)
	}
}

func TestUnblockUser_BadResponseFromNet(t *testing.T) {
	client := NewClient(nil, "")
	client.SetBaseURL("")

	want := "unsupported protocol scheme"

	err := client.Users.UnblockUser(1)
	if !strings.Contains(err.Error(), want) {
		t.Errorf("Users.BlockUser error.\nExpected error to contain: %s\nGot: %v", want, err)
	}
}

func TestDeactivateUser(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	path := fmt.Sprintf("/%susers/1/deactivate", apiVersionPath)
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		w.WriteHeader(http.StatusCreated)
	})

	err := client.Users.DeactivateUser(1)
	if err != nil {
		t.Errorf("Users.DeactivateUser returned error: %v", err)
	}
}

func TestDeactivateUser_UserNotFound(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	path := fmt.Sprintf("/%susers/1/deactivate", apiVersionPath)
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		w.WriteHeader(http.StatusNotFound)
	})

	err := client.Users.DeactivateUser(1)
	if err != ErrUserNotFound {
		t.Errorf("Users.DeactivateUser error.\nExpected: %+v\n\tGot: %+v", ErrUserNotFound, err)
	}
}

func TestDeactivateUser_DeactivatePrevented(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	path := fmt.Sprintf("/%susers/1/deactivate", apiVersionPath)
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		w.WriteHeader(http.StatusForbidden)
	})

	err := client.Users.DeactivateUser(1)
	if err != ErrUserDeactivatePrevented {
		t.Errorf("Users.DeactivateUser error.\nExpected: %+v\n\tGot: %+v", ErrUserDeactivatePrevented, err)
	}
}

func TestActivateUser(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	path := fmt.Sprintf("/%susers/1/activate", apiVersionPath)
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		w.WriteHeader(http.StatusCreated)
	})

	err := client.Users.ActivateUser(1)
	if err != nil {
		t.Errorf("Users.ActivateUser returned error: %v", err)
	}
}

func TestActivateUser_ActivatePrevented(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	path := fmt.Sprintf("/%susers/1/activate", apiVersionPath)
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		w.WriteHeader(http.StatusForbidden)
	})

	err := client.Users.ActivateUser(1)
	if err != ErrUserActivatePrevented {
		t.Errorf("Users.ActivateUser error.\nExpected: %+v\n\tGot: %+v", ErrUserActivatePrevented, err)
	}
}

func TestActivateUser_UserNotFound(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	path := fmt.Sprintf("/%susers/1/activate", apiVersionPath)
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		w.WriteHeader(http.StatusNotFound)
	})

	err := client.Users.ActivateUser(1)
	if err != ErrUserNotFound {
		t.Errorf("Users.ActivateUser error.\nExpected: %+v\n\tGot: %+v", ErrUserNotFound, err)
	}
}
