package drain

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/runatlantis/atlantis/server"
	"github.com/runatlantis/atlantis/server/logging"
)

// Start begins the shutdown process.
func Start(logger *logging.SimpleLogger) error {
	logger.Info("Drain starting")
	httpClient := &http.Client{}
	resp, err := startDrain(httpClient, logger)
	if err != nil {
		return err
	}
	logger.Info("Drain of server initiated succesfully")
	for {
		if resp.DrainCompleted {
			logger.Info("Drain of server completed successfully. You can now send a TERM signal to the server.")
			break
		}
		logger.Info("Drain of server still ongoing, waiting a little bit ...")
		time.Sleep(5 * time.Second)
		resp, err = getDrainStatus(httpClient, logger)
		if err != nil {
			return err
		}
	}
	return nil
}

func startDrain(httpClient *http.Client, logger *logging.SimpleLogger) (*server.DrainResponse, error) {

	req, err := http.NewRequest("POST", "http://localhost:4141/drain", nil)
	if err != nil {
		logger.Err("Failed to create POST request to /drain endpoint: %s", err)
		return nil, err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		logger.Err("Failed to make POST request to /drain endpoint: %s", err)
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Err("Failed to read reponse body of POST request to /drain endpoint: %s", err)
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated {
		logger.Err("Unexpected status code while making POST request to /drain endpoint: %d", resp.StatusCode)
		logger.Info("Response content: %s", string(body))
		return nil, errors.New("Unexpected status code")
	}

	var response server.DrainResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		logger.Err("Failed to parse reponse body of POST request to /drain endpoint: %s", err)
		return nil, err
	}
	return &response, nil
}

func getDrainStatus(httpClient *http.Client, logger *logging.SimpleLogger) (*server.DrainResponse, error) {

	req, err := http.NewRequest("GET", "http://localhost:4141/drain", nil)
	if err != nil {
		logger.Err("Failed to create GET request to /drain endpoint: %s", err)
		return nil, err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		logger.Err("Failed to make GET request to /drain endpoint: %s", err)
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Err("Failed to read reponse body of GET request to /drain endpoint: %s", err)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		logger.Err("Unexpected status code while making GET request to /drain endpoint: %d", resp.StatusCode)
		logger.Info("Response content: %s", string(body))
		return nil, errors.New("Unexpected status code")
	}

	var response server.DrainResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		logger.Err("Failed to parse reponse body of GET request to /drain endpoint: %s", err)
		return nil, err
	}
	return &response, nil
}
