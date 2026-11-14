package bitbucketcloud

import (
	"encoding/base64"
	"net/http"
	"strings"
	"testing"

	. "github.com/runatlantis/atlantis/testing"
)

func TestClient_APIUserDefault(t *testing.T) {
	client := New(http.DefaultClient, "user", "pass", "", "runatlantis.io")
	Equals(t, client.username, client.apiUser)
}

func TestClient_AuthHeaderAppPassword(t *testing.T) {
	client := New(http.DefaultClient, "user", "ATBBmockapppassword", "", "runatlantis.io")
	req, err := client.prepRequest("GET", "/mock/url", nil)
	Ok(t, err)
	authHeader := req.Header.Get("Authorization")
	scheme, encoded, ok := strings.Cut(authHeader, " ")
	Assert(t, ok, "expected Authorization header to contain scheme and credentials")
	Equals(t, "Basic", scheme)
	Equals(t, base64.StdEncoding.EncodeToString([]byte(client.apiUser+":"+client.password)), encoded)
}

func TestClient_AuthHeaderApiToken(t *testing.T) {
	// apiUser is the atlassian account email address
	client := New(http.DefaultClient, "user", "ATATmockapitoken", "user@myemail.net", "runatlantis.io")
	req, err := client.prepRequest("GET", "/mock/url", nil)
	Ok(t, err)
	authHeader := req.Header.Get("Authorization")
	scheme, encoded, ok := strings.Cut(authHeader, " ")
	Assert(t, ok, "expected Authorization header to contain scheme and credentials")
	Equals(t, "Basic", scheme)
	Equals(t, base64.StdEncoding.EncodeToString([]byte(client.apiUser+":"+client.password)), encoded)
}

func TestClient_AuthHeaderAccessToken(t *testing.T) {
	// apiUser is not required for Access Token authenticated API requests
	// The username "x-token-auth" is used for git interactions
	client := New(http.DefaultClient, "x-token-auth", "ATCTmockaccesstoken", "", "runatlantis.io")
	req, err := client.prepRequest("GET", "/mock/url", nil)
	Ok(t, err)
	authHeader := req.Header.Get("Authorization")
	scheme, token, ok := strings.Cut(authHeader, " ")
	Assert(t, ok, "expected Authorization header to contain scheme and token")
	Equals(t, "Bearer", scheme)
	Equals(t, client.password, token)
}
