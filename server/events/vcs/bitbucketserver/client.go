package bitbucketserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strings"

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

// NewClient builds a bitbucket cloud client. Returns an error if the baseURL is
// malformed. httpClient is the client to use to make the requests, username
// and password are used as basic auth in the requests, baseURL is the API's
// baseURL, ex. https://corp.com:7990. Don't include the API version, ex.
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
	if parsedURL.Scheme == "" {
		return nil, fmt.Errorf("must have 'http://' or 'https://' in base url %q", baseURL)
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

	projectKey, err := b.GetProjectKey(repo.Name, repo.SanitizedCloneURL)
	if err != nil {
		return nil, err
	}
	nextPageStart := "0"
	baseURL := fmt.Sprintf("%s/rest/api/1.0/projects/%s/repos/%s/pull-requests/%d/changes",
		b.BaseURL, projectKey, repo.Name, pull.Num)
	// We'll only loop 1000 times as a safety measure.
	maxLoops := 1000
	for i := 0; i < maxLoops; i++ {
		resp, err := b.makeRequest("GET", fmt.Sprintf("%s?start=%s", baseURL, nextPageStart), nil)
		if err != nil {
			return nil, err
		}
		var changes Changes
		if err := json.Unmarshal(resp, &changes); err != nil {
			return nil, errors.Wrapf(err, "Could not parse response %q", string(resp))
		}
		if err := validator.New().Struct(changes); err != nil {
			return nil, errors.Wrapf(err, "API response %q was missing fields", string(resp))
		}
		for _, v := range changes.Values {
			files = append(files, *v.Path.ToString)
		}
		if *changes.IsLastPage {
			break
		}
		nextPageStart = *changes.NextPageStart
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

func (b *Client) GetProjectKey(repoName string, cloneURL string) (string, error) {
	// Get the project key out of the repo clone URL.
	// Given http://bitbucket.corp:7990/scm/at/atlantis-example.git
	// we want to get 'at'.
	expr := fmt.Sprintf(".*/(.*?)/%s\\.git", repoName)
	capture, err := regexp.Compile(expr)
	if err != nil {
		return "", errors.Wrapf(err, "constructing regex from %q", expr)
	}
	matches := capture.FindStringSubmatch(cloneURL)
	if len(matches) != 2 {
		return "", fmt.Errorf("could not extract project key from %q, regex returned %q", cloneURL, strings.Join(matches, ","))
	}
	return matches[1], nil
}

// CreateComment creates a comment on the merge request.
// If comment length is greater than the max comment length we split into
// multiple comments.
func (b *Client) CreateComment(repo models.Repo, pullNum int, comment string) error {
	// maxCommentBodySize is derived from the error message when you go over
	// this limit.
	// real limit is 32768 but we want some padding for code blocks
	const maxCommentBodySize = 32748
	comments := b.splitAtMaxChars(comment, maxCommentBodySize, "\ncontinued...\n")

	projectKey, err := b.GetProjectKey(repo.Name, repo.SanitizedCloneURL)
	if err != nil {
		return err
	}
	path := fmt.Sprintf("%s/rest/api/1.0/projects/%s/repos/%s/pull-requests/%d/comments", b.BaseURL, projectKey, repo.Name, pullNum)

	for _, comment := range comments {
		bodyBytes, err := json.Marshal(map[string]string{"text": comment})
		if err != nil {
			return errors.Wrap(err, "json encoding")
		}
		_, err = b.makeRequest("POST", path, bytes.NewBuffer(bodyBytes))
		if err != nil {
			return err
		}
	}
	return nil
}

// splitAtMaxChars splits comment into a slice with string up to max
// len separated by join which gets appended to the ends of the middle strings.
// If max <= len(join) we return an empty slice since this is an edge case we
// don't want to handle.
func (b *Client) splitAtMaxChars(comment string, max int, join string) []string {
	// If we're under the limit then no need to split.
	if len(comment) <= max {
		return []string{comment}
	}
	// If we can't fit the joining string in then this doesn't make sense.
	if max <= len(join) {
		return nil
	}

	const startCodeBlockPattern = "```diff\n"
	const endCodeBlockPattern = "```\n\n"
	startCodeBlockRegexp := regexp.MustCompile(startCodeBlockPattern)
	endCodeBlockRegexp := regexp.MustCompile(endCodeBlockPattern)

	var comments []string
	maxSize := max - len(join)
	numComments := int(math.Ceil(float64(len(comment)) / float64(maxSize)))
	inCodeBlock := false
	for i := 0; i < numComments; i++ {
		upTo := b.min(len(comment), (i+1)*maxSize)
		portion := comment[i*maxSize : upTo]
		// We need to count # of "```" and "```diff" string
		startMatches := startCodeBlockRegexp.FindAllStringIndex(portion, -1)
		endMatches := endCodeBlockRegexp.FindAllStringIndex(portion, -1)
		countStartCodeBlock := len(startMatches)
		countEndCodeBlock := len(endMatches)
		portionPrefix := ""
		portionSufix := ""
		if countStartCodeBlock != countEndCodeBlock {
			inCodeBlock = true
			if countStartCodeBlock > countEndCodeBlock {
				// We started a block and we need to close it
				portionSufix = "\n" + endCodeBlockPattern
			} else if countStartCodeBlock < countEndCodeBlock {
				// We will close the block so we need to open it first
				inCodeBlock = false
				portionPrefix = startCodeBlockPattern
			}
		} else if countStartCodeBlock == countEndCodeBlock && inCodeBlock {
			// We are already inside a code block so we need to open it and close it
			portionPrefix = startCodeBlockPattern
			portionSufix = "\n" + endCodeBlockPattern
		}

		portion = portionPrefix + portion
		portion += portionSufix
		if i < numComments-1 {
			portion += join
		}
		comments = append(comments, portion)
	}
	return comments
}

func (b *Client) min(a, c int) int {
	if a < c {
		return a
	}
	return c
}

// PullIsApproved returns true if the merge request was approved.
func (b *Client) PullIsApproved(repo models.Repo, pull models.PullRequest) (bool, error) {
	projectKey, err := b.GetProjectKey(repo.Name, repo.SanitizedCloneURL)
	if err != nil {
		return false, err
	}
	path := fmt.Sprintf("%s/rest/api/1.0/projects/%s/repos/%s/pull-requests/%d", b.BaseURL, projectKey, repo.Name, pull.Num)
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
	for _, reviewer := range pullResp.Reviewers {
		if *reviewer.Approved {
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

	path := fmt.Sprintf("%s/rest/build-status/1.0/commits/%s", b.BaseURL, pull.HeadCommit)
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

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != 204 {
		respBody, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("making request %q unexpected status code: %d, body: %s", requestStr, resp.StatusCode, string(respBody))
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "reading response from request %q", requestStr)
	}
	return respBody, nil
}
