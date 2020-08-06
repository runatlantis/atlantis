package retryablehttp

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestRoundTripper_implements(t *testing.T) {
	// Compile-time proof of interface satisfaction.
	var _ http.RoundTripper = &RoundTripper{}
}

func TestRoundTripper_init(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer ts.Close()

	// Start with a new empty RoundTripper.
	rt := &RoundTripper{}

	// RoundTrip once.
	req, _ := http.NewRequest("GET", ts.URL, nil)
	if _, err := rt.RoundTrip(req); err != nil {
		t.Fatal(err)
	}

	// Check that the Client was initialized.
	if rt.Client == nil {
		t.Fatal("expected rt.Client to be initialized")
	}

	// Save the Client for later comparison.
	initialClient := rt.Client

	// RoundTrip again.
	req, _ = http.NewRequest("GET", ts.URL, nil)
	if _, err := rt.RoundTrip(req); err != nil {
		t.Fatal(err)
	}

	// Check that the underlying Client is unchanged.
	if rt.Client != initialClient {
		t.Fatalf("expected %v, got %v", initialClient, rt.Client)
	}
}

func TestRoundTripper_RoundTrip(t *testing.T) {
	var reqCount int32 = 0

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqNo := atomic.AddInt32(&reqCount, 1)
		if reqNo < 3 {
			w.WriteHeader(404)
		} else {
			w.WriteHeader(200)
			w.Write([]byte("success!"))
		}
	}))
	defer ts.Close()

	// Make a client with some custom settings to verify they are used.
	retryClient := NewClient()
	retryClient.CheckRetry = func(_ context.Context, resp *http.Response, _ error) (bool, error) {
		return resp.StatusCode == 404, nil
	}

	// Get the standard client and execute the request.
	client := retryClient.StandardClient()
	resp, err := client.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	// Check the response to ensure the client behaved as expected.
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if v, err := ioutil.ReadAll(resp.Body); err != nil {
		t.Fatal(err)
	} else if string(v) != "success!" {
		t.Fatalf("expected %q, got %q", "success!", v)
	}
}
