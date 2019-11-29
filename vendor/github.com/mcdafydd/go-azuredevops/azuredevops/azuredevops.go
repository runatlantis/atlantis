package azuredevops

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"reflect"
	"strings"
	"time"

	"github.com/google/go-querystring/query"
)

const (
	// DefaultBaseURL is the default URI base for most Azure Devops REST API endpoints
	DefaultBaseURL string = "https://dev.azure.com/"
	// DefaultVsspsBaseURL is the default URI base for some Azure Devops REST API endpoints
	DefaultVsspsBaseURL string = "https://vssps.dev.azure.com/"
	// userAgent our HTTP client's user-agent
	userAgent string = "go-azuredevops"
)

// Client for interacting with the Azure DevOps API
type Client struct {
	client *http.Client

	// BaseURL Comprised of baseURL and account
	BaseURL url.URL

	VsspsBaseURL url.URL

	UserAgent string

	// Account Default tenant identifier
	Account string

	// Services used to proxy to other API endpoints
	Boards           *BoardsService
	BuildDefinitions *BuildDefinitionsService
	Builds           *BuildsService
	DeliveryPlans    *DeliveryPlansService
	Favourites       *FavouritesService
	Git              *GitService
	Iterations       *IterationsService
	PullRequests     *PullRequestsService
	Teams            *TeamsService
	Tests            *TestsService
	Users            *UsersService
	WorkItems        *WorkItemsService
}

// NewClient returns a new Azure DevOps API client. If a nil httpClient is
// provided, http.DefaultClient will be used. To use API methods which require
// authentication, provide an http.Client that will perform the authentication
// for you (such as that provided by the golang.org/x/oauth2 library).
func NewClient(httpClient *http.Client) (*Client, error) {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	c := &Client{}
	baseURL, _ := url.Parse(DefaultBaseURL)
	vsspsBaseURL, _ := url.Parse(DefaultVsspsBaseURL)

	c.client = httpClient
	c.BaseURL = *baseURL
	c.VsspsBaseURL = *vsspsBaseURL
	c.UserAgent = userAgent

	c.Boards = &BoardsService{client: c}
	c.BuildDefinitions = &BuildDefinitionsService{client: c}
	c.Builds = &BuildsService{client: c}
	c.DeliveryPlans = &DeliveryPlansService{client: c}
	c.Favourites = &FavouritesService{client: c}
	c.Git = &GitService{client: c}
	c.Iterations = &IterationsService{client: c}
	c.PullRequests = &PullRequestsService{client: c}
	c.Teams = &TeamsService{client: c}
	c.Tests = &TestsService{client: c}
	c.Users = &UsersService{client: c}
	c.WorkItems = &WorkItemsService{client: c}

	return c, nil
}

// NewRequest creates an API request. A relative URL can be provided in urlStr,
// in which case it is resolved relative to the BaseURL of the Client.
// Relative URLs should always be specified without a preceding slash. If
// specified, the value pointed to by body is JSON encoded and included as the
// request body.
func (c *Client) NewRequest(method, urlStr string, body interface{}) (*http.Request, error) {
	if !strings.HasSuffix(c.BaseURL.Path, "/") {
		return nil, fmt.Errorf("BaseURL must have a trailing slash, but %q does not", c.BaseURL.String())
	}

	u, err := c.BaseURL.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)
		err := enc.Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}
	return req, nil
}

// Execute sends an API request and returns the API response. The API response is
// JSON decoded and stored in the value pointed to by r, or returned as an
// error if an API error has occurred. If r implements the io.Writer
// interface, the raw response body will be written to r, without attempting to
// first decode it.
//
// The provided ctx must be non-nil. If it is canceled or times out,
// ctx.Err() will be returned.
func (c *Client) Execute(ctx context.Context, req *http.Request, r interface{}) (*http.Response, error) {
	req = req.WithContext(ctx)
	debugReq(req)
	resp, err := c.client.Do(req)
	if err != nil {
		// If we got an error, and the context has been canceled,
		// the context's error is probably more useful.
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// If the error type is *url.Error, sanitize its URL before returning.
		if e, ok := err.(*url.Error); ok {
			if url, err := url.Parse(e.URL); err == nil {
				e.URL = sanitizeURL(url).String()
				return nil, e
			}
		}

		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return nil, fmt.Errorf("Request to %s responded with status %d", req.URL, resp.StatusCode)
	}

	if r != nil {
		if w, ok := r.(io.Writer); ok {
			io.Copy(w, resp.Body)
		} else {
			decErr := json.NewDecoder(resp.Body).Decode(r)
			if decErr == io.EOF {
				decErr = nil // ignore EOF errors caused by empty response body
			}
			if decErr != nil {
				err = decErr
			}
		}
	}

	return resp, err
}

// BasicAuthTransport is an http.RoundTripper that authenticates all requests
// using HTTP Basic Authentication with the provided username and password. It
// additionally supports users who have two-factor authentication enabled on
// their GitHub account.
type BasicAuthTransport struct {
	Username string // Azure Devops username
	Password string // Azure Devops password
	OTP      string // one-time password for users with two-factor auth enabled

	// Transport is the underlying HTTP transport to use when making requests.
	// It will default to http.DefaultTransport if nil.
	Transport http.RoundTripper
}

// RoundTrip implements the RoundTripper interface.
func (t *BasicAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// To set extra headers, we must make a copy of the Request so
	// that we don't modify the Request we were given. This is required by the
	// specification of http.RoundTripper.
	//
	// Since we are going to modify only req.Header here, we only need a deep copy
	// of req.Header.
	req2 := new(http.Request)
	*req2 = *req
	req2.Header = make(http.Header, len(req.Header))
	for k, s := range req.Header {
		req2.Header[k] = append([]string(nil), s...)
	}

	//req2.SetBasicAuth(t.Username, t.Password)
	req2.SetBasicAuth("", t.Password)

	return t.transport().RoundTrip(req2)
}

// Client returns an *http.Client that makes requests that are authenticated
// using HTTP Basic Authentication.
func (t *BasicAuthTransport) Client() *http.Client {
	return &http.Client{Transport: t}
}

func (t *BasicAuthTransport) transport() http.RoundTripper {
	if t.Transport != nil {
		return t.Transport
	}
	return http.DefaultTransport
}

// addOptions adds the parameters in opt as URL query parameters to s. opt
// must be a struct whose fields may contain "url" tags.
// From: https://github.com/google/go-github/blob/master/github/github.go
func addOptions(s string, opt interface{}) (string, error) {
	v := reflect.ValueOf(opt)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return s, nil
	}

	u, err := url.Parse(s)
	if err != nil {
		return s, err
	}

	qs, err := query.Values(opt)
	if err != nil {
		return s, err
	}

	for k, v := range u.Query() {
		qs[k] = v
	}

	u.RawQuery = qs.Encode()
	return u.String(), nil
}

// formatRef helper function for API calls that need a branch reference
// as an input parameter.  Doesn't do much error checking.
// Examples:
// *ref = "mybranch" => *ref = "refs/heads/mybranch"
// *ref = "refs/heads/abranch" => *ref = "refs/heads/abranch"
func formatRef(ref *string) error {
	matched, err := path.Match("refs/heads/*", *ref)
	if matched && err == nil {
		return nil
	} else if err != nil {
		return err
	} else {
		*ref = fmt.Sprintf("refs/heads/%s", *ref)
		return nil
	}
}

// sanitizeURL redacts the client_secret parameter from the URL which may be
// exposed to the user.
func sanitizeURL(uri *url.URL) *url.URL {
	if uri == nil {
		return nil
	}
	params := uri.Query()
	if len(params.Get("client_secret")) > 0 {
		params.Set("client_secret", "REDACTED")
		uri.RawQuery = params.Encode()
	}
	return uri
}

type Time struct {
	Time time.Time
}

func (t *Time) UnmarshalJSON(b []byte) error {
	t2 := time.Time{}
	err := json.Unmarshal(b, &t2)

	// ignore errors for 0001-01-01T00:00:00 dates. The Azure DevOps service
	// returns default dates in a format that is invalid for a time.Time. The
	// correct value would have a 'z' at the end to represent utc. We are going
	// to ignore this error, and set the value to the default time.Time value.
	// https://github.com/microsoft/azure-devops-go-api/issues/17
	if err != nil {
		if parseError, ok := err.(*time.ParseError); ok && parseError.Value == "\"0001-01-01T00:00:00\"" {
			err = nil
		}
	}

	t.Time = t2
	return err
}

func (t *Time) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Time)
}

func (t Time) String() string {
	return t.Time.String()
}

func (t Time) Equal(u Time) bool {
	return t.Time.Equal(u.Time)
}

// Bool is a helper routine that allocates a new bool value
// to store v and returns a pointer to it.
func Bool(v bool) *bool { return &v }

// Int is a helper routine that allocates a new int value
// to store v and returns a pointer to it.
func Int(v int) *int { return &v }

// Int64 is a helper routine that allocates a new int64 value
// to store v and returns a pointer to it.
func Int64(v int64) *int64 { return &v }

// String is a helper routine that allocates a new string value
// to store v and returns a pointer to it.
func String(v string) *string { return &v }
