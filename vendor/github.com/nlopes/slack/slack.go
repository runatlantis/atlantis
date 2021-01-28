package slack

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
)

// Added as a var so that we can change this for testing purposes
var SLACK_API string = "https://slack.com/api/"
var SLACK_WEB_API_FORMAT string = "https://%s.slack.com/api/users.admin.%s?t=%s"

// HTTPClient sets a custom http.Client
// deprecated: in favor of SetHTTPClient()
var HTTPClient = &http.Client{}

var customHTTPClient HTTPRequester = HTTPClient

// HTTPRequester defines the minimal interface needed for an http.Client to be implemented.
//
// Use it in conjunction with the SetHTTPClient function to allow for other capabilities
// like a tracing http.Client
type HTTPRequester interface {
	Do(*http.Request) (*http.Response, error)
}

// SetHTTPClient allows you to specify a custom http.Client
// Use this instead of the package level HTTPClient variable if you want to use a custom client like the
// Stackdriver Trace HTTPClient https://godoc.org/cloud.google.com/go/trace#HTTPClient
func SetHTTPClient(client HTTPRequester) {
	customHTTPClient = client
}

// ResponseMetadata holds pagination metadata
type ResponseMetadata struct {
	Cursor string `json:"next_cursor"`
}

func (t *ResponseMetadata) initialize() *ResponseMetadata {
	if t != nil {
		return t
	}

	return &ResponseMetadata{}
}

type AuthTestResponse struct {
	URL    string `json:"url"`
	Team   string `json:"team"`
	User   string `json:"user"`
	TeamID string `json:"team_id"`
	UserID string `json:"user_id"`
}

type authTestResponseFull struct {
	SlackResponse
	AuthTestResponse
}

type Client struct {
	token      string
	info       Info
	debug      bool
	httpclient HTTPRequester
}

// Option defines an option for a Client
type Option func(*Client)

// OptionHTTPClient - provide a custom http client to the slack client.
func OptionHTTPClient(c HTTPRequester) func(*Client) {
	return func(s *Client) {
		s.httpclient = c
	}
}

// New builds a slack client from the provided token and options.
func New(token string, options ...Option) *Client {
	s := &Client{
		token:      token,
		httpclient: customHTTPClient,
	}

	for _, opt := range options {
		opt(s)
	}

	return s
}

// AuthTest tests if the user is able to do authenticated requests or not
func (api *Client) AuthTest() (response *AuthTestResponse, error error) {
	return api.AuthTestContext(context.Background())
}

// AuthTestContext tests if the user is able to do authenticated requests or not with a custom context
func (api *Client) AuthTestContext(ctx context.Context) (response *AuthTestResponse, error error) {
	api.Debugf("Challenging auth...")
	responseFull := &authTestResponseFull{}
	err := postSlackMethod(ctx, api.httpclient, "auth.test", url.Values{"token": {api.token}}, responseFull, api.debug)
	if err != nil {
		api.Debugf("failed to test for auth: %s", err)
		return nil, err
	}
	if !responseFull.Ok {
		api.Debugf("auth response was not Ok: %s", responseFull.Error)
		return nil, errors.New(responseFull.Error)
	}

	api.Debugf("Auth challenge was successful with response %+v", responseFull.AuthTestResponse)
	return &responseFull.AuthTestResponse, nil
}

// SetDebug switches the api into debug mode
// When in debug mode, it logs various info about what its doing
// If you ever use this in production, don't call SetDebug(true)
func (api *Client) SetDebug(debug bool) {
	api.debug = debug
	if debug && logger == nil {
		SetLogger(log.New(os.Stdout, "nlopes/slack", log.LstdFlags|log.Lshortfile))
	}
}

// Debugf print a formatted debug line.
func (api *Client) Debugf(format string, v ...interface{}) {
	if api.debug {
		logger.Output(2, fmt.Sprintf(format, v...))
	}
}

// Debugln print a debug line.
func (api *Client) Debugln(v ...interface{}) {
	if api.debug {
		logger.Output(2, fmt.Sprintln(v...))
	}
}
