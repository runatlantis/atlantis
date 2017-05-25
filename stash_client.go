package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"github.hootops.com/production-delivery/atlantis/logging"
)

const stashUrl = "http://stash.hootops.com"

type StashClient struct{}

type StashLockResponse struct {
	Message         string `json:"message"`
	PullRequestLink string `json:"pull_request_link"`
	DiscardUrl      string `json:"discard_url"`
	StatusCode      int64  `json:"status_code"`
	Success         bool   `json:"success"`
}

type StashUnlockResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

// todo: refactor so this returns error
func (s *StashClient) LockState(log *logging.SimpleLogger, path string, state []byte) StashLockResponse {
	req, err := http.NewRequest("POST", stashUrl+"/lock/"+path, bytes.NewBuffer(state))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Err("failed making lock request to Stash: %v", err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	log.Debug("stash response body: %s", string(body))

	var jsonData StashLockResponse
	err = json.Unmarshal(body, &jsonData)
	if err != nil {
		log.Err("failed parsing lock stash response: %v", err)
	}

	return jsonData
}

func (s *StashClient) UnlockState(log *logging.SimpleLogger, path string, state []byte) StashUnlockResponse {
	req, err := http.NewRequest("POST", stashUrl+"/unlock/"+path, bytes.NewBuffer(state))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Err("failed making unlock request to Stash: %v", err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	log.Info("stash response body: %s", string(body))

	var jsonData StashUnlockResponse
	err = json.Unmarshal(body, &jsonData)
	if err != nil {
		log.Err("failed parsing unlock stash response: %v", err)
	}

	return jsonData
}
