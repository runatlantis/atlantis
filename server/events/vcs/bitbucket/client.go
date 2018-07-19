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

type Client struct {
	HttpClient      *http.Client
	Username        string
	Password        string
	BaseURL         string
	AtlantisBaseURL string
}

func NewClient(httpClient *http.Client, username string, password string, baseURL string, atlantisBaseUrl string) (*Client, error) {
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
		HttpClient:      httpClient,
		Username:        username,
		Password:        password,
		BaseURL:         urlWithoutPath,
		AtlantisBaseURL: atlantisBaseUrl,
	}, nil
}

func (b *Client) prepRequest(method string, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, path, body)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(b.Username, b.Password)
	return req, nil
}

// GetModifiedFiles returns the names of files that were modified in the merge request.
// The names include the path to the file from the repo root, ex. parent/child/file.txt.
func (b *Client) GetModifiedFiles(repo models.Repo, pull models.PullRequest) ([]string, error) {
	path := fmt.Sprintf("%s/2.0/repositories/%s/pullrequests/%d/diffstat", b.BaseURL, repo.FullName, pull.Num)
	// todo: remove duplication
	req, err := b.prepRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	resp, err := b.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}
	// todo: pagination
	var diffStat DiffStat
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "reading body")
	}
	if err := json.Unmarshal(body, &diffStat); err != nil {
		return nil, err
	}

	hash := make(map[string]bool)
	var unique []string
	for _, v := range diffStat.Values {
		var paths []string
		if v.Old != nil {
			paths = append(paths, *v.Old.Path)
		}
		if v.New != nil {
			paths = append(paths, *v.New.Path)
		}
		for _, path := range paths {
			if !hash[path] {
				unique = append(unique, path)
				hash[path] = true
			}
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
	path := fmt.Sprintf("%s/1.0/repositories/%s/pullrequests/%d/comments", b.BaseURL, repo.FullName, pullNum)
	req, err := b.prepRequest("POST", path, bytes.NewBuffer(bodyBytes))
	req.Header.Add("Content-Type", "application/json")
	if err != nil {
		return errors.Wrap(err, "constructing request")
	}
	resp, err := b.HttpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d for url: %s, body: %s", resp.StatusCode, path, string(body))
	}
	return nil
}

// PullIsApproved returns true if the merge request was approved.
func (b *Client) PullIsApproved(repo models.Repo, pull models.PullRequest) (bool, error) {
	path := fmt.Sprintf("%s/2.0/repositories/%s/pullrequests/%d", b.BaseURL, repo.FullName, pull.Num)
	req, err := b.prepRequest("GET", path, nil)
	if err != nil {
		return false, errors.Wrap(err, "constructing request")
	}
	resp, err := b.HttpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return false, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, errors.Wrap(err, "reading response body")
	}

	parsedPull, err := ParseBitBucketPullRequest(body)
	if err != nil {
		return false, errors.Wrap(err, "parsing response")
	}
	for _, participant := range parsedPull.Participants {
		if *participant.Approved == true {
			return true, nil
		}
	}
	return false, nil
}

// UpdateStatus updates the ~ status of a commit.
func (b *Client) UpdateStatus(repo models.Repo, pull models.PullRequest, state models.CommitStatus, description string) error {
	bbState := "FAILED"
	switch state {
	case models.Pending:
		bbState = "INPROGRESS"
	case models.Success:
		bbState = "SUCCESSFUL"
	case models.Failed:
		bbState = "FAILED"
	}

	bodyBytes, err := json.Marshal(map[string]string{
		"key":         "atlantis",
		"url":         b.AtlantisBaseURL,
		"state":       bbState,
		"description": description,
	})

	path := fmt.Sprintf("%s/2.0/repositories/%s/commit/%s/statuses/build", b.BaseURL, repo.FullName, pull.HeadCommit)
	if err != nil {
		return errors.Wrap(err, "json encoding")
	}
	req, err := b.prepRequest("POST", path, bytes.NewBuffer(bodyBytes))
	req.Header.Add("Content-Type", "application/json")
	if err != nil {
		return errors.Wrap(err, "constructing request")
	}
	resp, err := b.HttpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// Check for both 201 and 200 because on a new status we'll get a 201 but
	// on a status update we'll get a 200.
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}
	return nil
}
