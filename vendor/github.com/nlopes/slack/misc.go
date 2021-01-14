package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type SlackResponse struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error"`
}

func (t SlackResponse) Err() error {
	if t.Ok {
		return nil
	}

	// handle pure text based responses like chat.post
	// which while they have a slack response in their data structure
	// it doesn't actually get set during parsing.
	if strings.TrimSpace(t.Error) == "" {
		return nil
	}

	return errors.New(t.Error)
}

// StatusCodeError represents an http response error.
// type httpStatusCode interface { HTTPStatusCode() int } to handle it.
type statusCodeError struct {
	Code   int
	Status string
}

func (t statusCodeError) Error() string {
	// TODO: this is a bad error string, should clean it up with a breaking changes
	// merger.
	return fmt.Sprintf("Slack server error: %s.", t.Status)
}

func (t statusCodeError) HTTPStatusCode() int {
	return t.Code
}

type RateLimitedError struct {
	RetryAfter time.Duration
}

func (e *RateLimitedError) Error() string {
	return fmt.Sprintf("Slack rate limit exceeded, retry after %s", e.RetryAfter)
}

func fileUploadReq(ctx context.Context, path, fieldname, filename string, values url.Values, r io.Reader) (*http.Request, error) {
	body := &bytes.Buffer{}
	wr := multipart.NewWriter(body)

	ioWriter, err := wr.CreateFormFile(fieldname, filename)
	if err != nil {
		wr.Close()
		return nil, err
	}
	_, err = io.Copy(ioWriter, r)
	if err != nil {
		wr.Close()
		return nil, err
	}
	// Close the multipart writer or the footer won't be written
	wr.Close()
	req, err := http.NewRequest("POST", path, body)
	req = req.WithContext(ctx)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", wr.FormDataContentType())
	req.URL.RawQuery = (values).Encode()
	return req, nil
}

func parseResponseBody(body io.ReadCloser, intf interface{}, debug bool) error {
	response, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}

	// FIXME: will be api.Debugf
	if debug {
		logger.Printf("parseResponseBody: %s\n", string(response))
	}

	return json.Unmarshal(response, intf)
}

func postLocalWithMultipartResponse(ctx context.Context, client HTTPRequester, path, fpath, fieldname string, values url.Values, intf interface{}, debug bool) error {
	fullpath, err := filepath.Abs(fpath)
	if err != nil {
		return err
	}
	file, err := os.Open(fullpath)
	if err != nil {
		return err
	}
	defer file.Close()
	return postWithMultipartResponse(ctx, client, path, filepath.Base(fpath), fieldname, values, file, intf, debug)
}

func postWithMultipartResponse(ctx context.Context, client HTTPRequester, path, name, fieldname string, values url.Values, r io.Reader, intf interface{}, debug bool) error {
	req, err := fileUploadReq(ctx, SLACK_API+path, fieldname, name, values, r)
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		retry, err := strconv.ParseInt(resp.Header.Get("Retry-After"), 10, 64)
		if err != nil {
			return err
		}
		return &RateLimitedError{time.Duration(retry) * time.Second}
	}

	// Slack seems to send an HTML body along with 5xx error codes. Don't parse it.
	if resp.StatusCode != http.StatusOK {
		logResponse(resp, debug)
		return statusCodeError{Code: resp.StatusCode, Status: resp.Status}
	}

	return parseResponseBody(resp.Body, intf, debug)
}

func doPost(ctx context.Context, client HTTPRequester, req *http.Request, intf interface{}, debug bool) error {
	req = req.WithContext(ctx)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		retry, err := strconv.ParseInt(resp.Header.Get("Retry-After"), 10, 64)
		if err != nil {
			return err
		}
		return &RateLimitedError{time.Duration(retry) * time.Second}
	}

	// Slack seems to send an HTML body along with 5xx error codes. Don't parse it.
	if resp.StatusCode != http.StatusOK {
		logResponse(resp, debug)
		return statusCodeError{Code: resp.StatusCode, Status: resp.Status}
	}

	return parseResponseBody(resp.Body, intf, debug)
}

// post JSON.
func postJSON(ctx context.Context, client HTTPRequester, endpoint, token string, json []byte, intf interface{}, debug bool) error {
	reqBody := bytes.NewBuffer(json)
	req, err := http.NewRequest("POST", endpoint, reqBody)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	return doPost(ctx, client, req, intf, debug)
}

// post a url encoded form.
func postForm(ctx context.Context, client HTTPRequester, endpoint string, values url.Values, intf interface{}, debug bool) error {
	reqBody := strings.NewReader(values.Encode())
	req, err := http.NewRequest("POST", endpoint, reqBody)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return doPost(ctx, client, req, intf, debug)
}

// post to a slack web method.
func postSlackMethod(ctx context.Context, client HTTPRequester, path string, values url.Values, intf interface{}, debug bool) error {
	return postForm(ctx, client, SLACK_API+path, values, intf, debug)
}

func parseAdminResponse(ctx context.Context, client HTTPRequester, method string, teamName string, values url.Values, intf interface{}, debug bool) error {
	endpoint := fmt.Sprintf(SLACK_WEB_API_FORMAT, teamName, method, time.Now().Unix())
	return postForm(ctx, client, endpoint, values, intf, debug)
}

func logResponse(resp *http.Response, debug bool) error {
	if debug {
		text, err := httputil.DumpResponse(resp, true)
		if err != nil {
			return err
		}

		logger.Print(string(text))
	}

	return nil
}

func okJSONHandler(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	response, _ := json.Marshal(SlackResponse{
		Ok: true,
	})
	rw.Write(response)
}

type errorString string

func (t errorString) Error() string {
	return string(t)
}

// timerReset safely reset a timer, see time.Timer.Reset for details.
func timerReset(t *time.Timer, d time.Duration) {
	if !t.Stop() {
		<-t.C
	}
	t.Reset(d)
}
