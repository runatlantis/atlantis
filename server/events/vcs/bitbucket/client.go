package bitbucket

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/models"
)

const (
	// The comments API is only available in the 1.0 API right now.
	commentsAPIPathFmt     = "%s/1.0/repositories/%s/pullrequests/%d/comments"
	diffStatAPIPathFmt     = "%s/2.0/repositories/%s/pullrequests/%d/diffstat"
	pullApprovedAPIPathFmt = "%s/2.0/repositories/%s/pullrequests/%d"
	buildStatusAPIPathFmt  = "%s/2.0/repositories/%s/commit/%s/statuses/build"
)

type Client struct {
	HttpClient  *http.Client
	Username    string
	Password    string
	BaseURL     string
	AtlantisURL string
}

// NewClient builds a bitbucket client. Returns an error if the baseURL is
// malformed. httpClient is the client to use to make the requests, username
// and password are used as basic auth in the requests, baseURL is the API's
// baseURL, ex. https://api.bitbucket.org. Don't include the API version, ex.
// '/1.0' since that changes based on the API call. atlantisURL is the
// URL for Atlantis that will be linked to from the build status icons. This
// linking is annoying because we don't have anywhere good to link but a URL is
// required.
func NewClient(httpClient *http.Client, username string, password string, baseURL string, atlantisURL string) (*Client, error) {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	// Remove the trailing '/' from the URL.
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, errors.Wrapf(err, "parsing %s", baseURL)
	}
	urlWithoutPath := fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)
	return &Client{
		HttpClient:  httpClient,
		Username:    username,
		Password:    password,
		BaseURL:     urlWithoutPath,
		AtlantisURL: atlantisURL,
	}, nil
}

// GetModifiedFiles returns the names of files that were modified in the merge request.
// The names include the path to the file from the repo root, ex. parent/child/file.txt.
func (b *Client) GetModifiedFiles(repo models.Repo, pull models.PullRequest) ([]string, error) {
	var files []string

	nextPageURL := fmt.Sprintf(diffStatAPIPathFmt, b.BaseURL, repo.FullName, pull.Num)
	// We'll only loop 100 times as a safety measure.
	maxLoops := 100
	for i := 0; i < maxLoops; i++ {
		resp, err := b.makeRequest("GET", nextPageURL, nil)
		if err != nil {
			return nil, err
		}
		var diffStat DiffStat
		if err := json.Unmarshal(resp, &diffStat); err != nil {
			return nil, err
		}
		for _, v := range diffStat.Values {
			if v.Old != nil {
				files = append(files, *v.Old.Path)
			}
			if v.New != nil {
				files = append(files, *v.New.Path)
			}
		}
		if diffStat.Next == nil || *diffStat.Next == "" {
			break
		}
		nextPageURL = *diffStat.Next
	}

	// Now ensure all files are unique.
	hash := make(map[string]bool)
	var unique []string
	for _, f := range files {
		if !hash[f] {
			unique = append(unique, f)
			hash[f] = true
		}
	}
	return unique, nil
}

// CreateComment creates a comment on the merge request.
func (b *Client) CreateComment(repo models.Repo, pullNum int, comment string) error {
	bodyBytes, err := json.Marshal(map[string]string{"content": comment})
	if err != nil {
		return errors.Wrap(err, "json encoding")
	}
	path := fmt.Sprintf(commentsAPIPathFmt, b.BaseURL, repo.FullName, pullNum)
	_, err = b.makeRequest("POST", path, bytes.NewBuffer(bodyBytes))
	return err
}

// PullIsApproved returns true if the merge request was approved.
func (b *Client) PullIsApproved(repo models.Repo, pull models.PullRequest) (bool, error) {
	path := fmt.Sprintf(pullApprovedAPIPathFmt, b.BaseURL, repo.FullName, pull.Num)
	resp, err := b.makeRequest("GET", path, nil)
	if err != nil {
		return false, err
	}
	var pullResp PullRequest
	if err := json.Unmarshal(resp, &pullResp); err != nil {
		return false, err
	}
	for _, participant := range pullResp.Participants {
		if *participant.Approved == true {
			return true, nil
		}
	}
	return false, nil
}

// UpdateStatus updates the status of a commit.
func (b *Client) UpdateStatus(repo models.Repo, pull models.PullRequest, status models.CommitStatus, description string) error {
	bbState := "FAILED"
	switch status {
	case models.PendingCommitStatus:
		bbState = "INPROGRESS"
	case models.SuccessCommitStatus:
		bbState = "SUCCESSFUL"
	case models.FailedCommitStatus:
		bbState = "FAILED"
	}

	bodyBytes, err := json.Marshal(map[string]string{
		"key":         "atlantis",
		"url":         b.AtlantisURL,
		"state":       bbState,
		"description": description,
	})

	path := fmt.Sprintf(buildStatusAPIPathFmt, b.BaseURL, repo.FullName, pull.HeadCommit)
	if err != nil {
		return errors.Wrap(err, "json encoding")
	}
	_, err = b.makeRequest("POST", path, bytes.NewBuffer(bodyBytes))
	return err
}

// prepRequest adds the HTTP basic auth.
func (b *Client) prepRequest(method string, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, path, body)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(b.Username, b.Password)
	return req, nil
}

func (b *Client) makeRequest(method string, path string, reqBody io.Reader) ([]byte, error) {
	req, err := b.prepRequest(method, path, reqBody)
	if err != nil {
		return nil, errors.Wrap(err, "constructing request")
	}
	if reqBody != nil {
		req.Header.Add("Content-Type", "application/json")
	}
	resp, err := b.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	requestStr := fmt.Sprintf("%s %s", method, path)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("making request %q unexpected status code: %d, body: %s", requestStr, resp.StatusCode, string(respBody))
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "reading response from request %q", requestStr)
	}
	return respBody, nil
}
