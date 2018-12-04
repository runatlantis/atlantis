package bitbucketcloud

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/models"
	"gopkg.in/go-playground/validator.v9"
)

type Client struct {
	HttpClient  *http.Client
	Username    string
	Password    string
	BaseURL     string
	AtlantisURL string
}

// NewClient builds a bitbucket cloud client. atlantisURL is the
// URL for Atlantis that will be linked to from the build status icons. This
// linking is annoying because we don't have anywhere good to link but a URL is
// required.
func NewClient(httpClient *http.Client, username string, password string, atlantisURL string) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{
		HttpClient:  httpClient,
		Username:    username,
		Password:    password,
		BaseURL:     BaseURL,
		AtlantisURL: atlantisURL,
	}
}

// GetModifiedFiles returns the names of files that were modified in the merge request.
// The names include the path to the file from the repo root, ex. parent/child/file.txt.
func (b *Client) GetModifiedFiles(repo models.Repo, pull models.PullRequest) ([]string, error) {
	var files []string

	nextPageURL := fmt.Sprintf("%s/2.0/repositories/%s/pullrequests/%d/diffstat", b.BaseURL, repo.FullName, pull.Num)
	// We'll only loop 1000 times as a safety measure.
	maxLoops := 1000
	for i := 0; i < maxLoops; i++ {
		resp, err := b.makeRequest("GET", nextPageURL, nil)
		if err != nil {
			return nil, err
		}
		var diffStat DiffStat
		if err := json.Unmarshal(resp, &diffStat); err != nil {
			return nil, errors.Wrapf(err, "Could not parse response %q", string(resp))
		}
		if err := validator.New().Struct(diffStat); err != nil {
			return nil, errors.Wrapf(err, "API response %q was missing fields", string(resp))
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
	// NOTE: I tried to find the maximum size of a comment for bitbucket.org but
	// I got up to 200k chars without issue so for now I'm not going to bother
	// to detect this.
	bodyBytes, err := json.Marshal(map[string]map[string]string{"content": {
		"raw": comment,
	}})
	if err != nil {
		return errors.Wrap(err, "json encoding")
	}
	path := fmt.Sprintf("%s/2.0/repositories/%s/pullrequests/%d/comments", b.BaseURL, repo.FullName, pullNum)
	_, err = b.makeRequest("POST", path, bytes.NewBuffer(bodyBytes))
	return err
}

// PullIsApproved returns true if the merge request was approved.
func (b *Client) PullIsApproved(repo models.Repo, pull models.PullRequest) (bool, error) {
	path := fmt.Sprintf("%s/2.0/repositories/%s/pullrequests/%d", b.BaseURL, repo.FullName, pull.Num)
	resp, err := b.makeRequest("GET", path, nil)
	if err != nil {
		return false, err
	}
	var pullResp PullRequest
	if err := json.Unmarshal(resp, &pullResp); err != nil {
		return false, errors.Wrapf(err, "Could not parse response %q", string(resp))
	}
	if err := validator.New().Struct(pullResp); err != nil {
		return false, errors.Wrapf(err, "API response %q was missing fields", string(resp))
	}
	for _, participant := range pullResp.Participants {
		// Bitbucket allows the author to approve their own pull request. This
		// defeats the purpose of approvals so we don't count that approval.
		if *participant.Approved && *participant.User.Username != pull.Author {
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

	path := fmt.Sprintf("%s/2.0/repositories/%s/commit/%s/statuses/build", b.BaseURL, repo.FullName, pull.HeadCommit)
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
	defer resp.Body.Close() // nolint: errcheck
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
