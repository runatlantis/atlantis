package gitlab

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
)

// setup sets up a test HTTP server along with a gitlab.Client that is
// configured to talk to that test server.  Tests should register handlers on
// mux which provide mock responses for the API method being tested.
func setup() (*http.ServeMux, *httptest.Server, *Client) {
	// mux is the HTTP request multiplexer used with the test server.
	mux := http.NewServeMux()

	// server is a test HTTP server used to provide mock API responses.
	server := httptest.NewServer(mux)

	// client is the Gitlab client being tested.
	client := NewClient(nil, "")
	client.SetBaseURL(server.URL)

	return mux, server, client
}

// teardown closes the test HTTP server.
func teardown(server *httptest.Server) {
	server.Close()
}

func testURL(t *testing.T, r *http.Request, want string) {
	if got := r.RequestURI; got != want {
		t.Errorf("Request url: %+v, want %s", got, want)
	}
}

func testMethod(t *testing.T, r *http.Request, want string) {
	if got := r.Method; got != want {
		t.Errorf("Request method: %s, want %s", got, want)
	}
}

func testBody(t *testing.T, r *http.Request, want string) {
	buffer := new(bytes.Buffer)
	_, err := buffer.ReadFrom(r.Body)

	if err != nil {
		t.Fatalf("Failed to Read Body: %v", err)
	}

	if got := buffer.String(); got != want {
		t.Errorf("Request body: %s, want %s", got, want)
	}
}

func mustWriteHTTPResponse(t *testing.T, w io.Writer, fixturePath string) {
	f, err := os.Open(fixturePath)
	if err != nil {
		t.Fatalf("error opening fixture file: %v", err)
	}

	if _, err = io.Copy(w, f); err != nil {
		t.Fatalf("error writing response: %v", err)
	}
}

func errorOption(*retryablehttp.Request) error {
	return errors.New("OptionFunc returns an error")
}

func TestNewClient(t *testing.T) {
	c := NewClient(nil, "")
	expectedBaseURL := defaultBaseURL + apiVersionPath

	if c.BaseURL().String() != expectedBaseURL {
		t.Errorf("NewClient BaseURL is %s, want %s", c.BaseURL().String(), expectedBaseURL)
	}
	if c.UserAgent != userAgent {
		t.Errorf("NewClient UserAgent is %s, want %s", c.UserAgent, userAgent)
	}
}

func TestSetBaseURL(t *testing.T) {
	expectedBaseURL := "http://gitlab.local/foo/" + apiVersionPath
	c := NewClient(nil, "")
	err := c.SetBaseURL("http://gitlab.local/foo")
	if err == nil {
		t.Errorf("Expected a 'no such host' error, got: %v", err)
	}
	if c.BaseURL().String() != expectedBaseURL {
		t.Errorf("BaseURL is %s, want %s", c.BaseURL().String(), expectedBaseURL)
	}
}

func TestCheckResponse(t *testing.T) {
	req, err := NewClient(nil, "").NewRequest("GET", "test", nil, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	resp := &http.Response{
		Request:    req.Request,
		StatusCode: http.StatusBadRequest,
		Body: ioutil.NopCloser(strings.NewReader(`
		{
			"message": {
				"prop1": [
					"message 1",
					"message 2"
				],
				"prop2":[
					"message 3"
				],
				"embed1": {
					"prop3": [
						"msg 1",
						"msg2"
					]
				},
				"embed2": {
					"prop4": [
						"some msg"
					]
				}
			},
			"error": "message 1"
		}`)),
	}

	errResp := CheckResponse(resp)
	if errResp == nil {
		t.Fatal("Expected error response.")
	}

	want := "GET https://gitlab.com/api/v4/test: 400 {error: message 1}, {message: {embed1: {prop3: [msg 1, msg2]}}, {embed2: {prop4: [some msg]}}, {prop1: [message 1, message 2]}, {prop2: [message 3]}}"

	if errResp.Error() != want {
		t.Errorf("Expected error: %s, got %s", want, errResp.Error())
	}
}

func TestRequestWithContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), interface{}("myKey"), interface{}("myValue"))
	req, err := NewClient(nil, "").NewRequest("GET", "test", nil, []OptionFunc{WithContext(ctx)})
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	if req.Context() != ctx {
		t.Fatal("Context was not set correctly")
	}
}

func TestBoolValue(t *testing.T) {
	testCases := map[string]struct {
		data     []byte
		expected bool
	}{
		"should unmarshal true as true": {
			data:     []byte("true"),
			expected: true,
		},
		"should unmarshal false as true": {
			data:     []byte("false"),
			expected: false,
		},
		"should unmarshal \"1\" as true": {
			data:     []byte(`"1"`),
			expected: true,
		},
		"should unmarshal \"0\" as false": {
			data:     []byte(`"0"`),
			expected: false,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			var b BoolValue
			if err := json.Unmarshal(testCase.data, &b); err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if bool(b) != testCase.expected {
				t.Fatalf("Expected %v but got %v", testCase.expected, b)
			}
		})
	}
}

func loadFixture(filePath string) []byte {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
	}

	return content
}
