package controllers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/runatlantis/atlantis/server/controllers"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tally "github.com/uber-go/tally/v4"
)

func TestConfigController_ReloadConfig_Success(t *testing.T) {
	// Setup
	logger := logging.NewNoopLogger(t)
	scope := tally.NoopScope
	apiSecret := "test-secret"

	// Create a mock command runner with initial config
	commandRunner := &events.DefaultCommandRunner{
		GlobalCfg: valid.NewGlobalCfgFromArgs(valid.GlobalCfgArgs{}),
	}

	configController := &controllers.ConfigController{
		APISecret:          []byte(apiSecret),
		Logger:             logger,
		CommandRunner:      commandRunner,
		RepoConfig:         "",
		RepoConfigJSON:     "",
		PolicyCheckEnabled: false,
		Scope:              scope,
	}

	// Create test request
	reqBody := controllers.ConfigReloadRequest{}
	jsonBody, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", "/api/config/reload", bytes.NewBuffer(jsonBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Atlantis-Token", apiSecret)

	rr := httptest.NewRecorder()

	// Execute
	configController.ReloadConfig(rr, req)

	// Verify
	assert.Equal(t, http.StatusOK, rr.Code)

	var response controllers.ConfigReloadResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Success)
	assert.Contains(t, response.Message, "Configuration reloaded successfully")
}

func TestConfigController_ReloadConfig_InvalidAuth(t *testing.T) {
	// Setup
	logger := logging.NewNoopLogger(t)
	scope := tally.NoopScope
	apiSecret := "test-secret"

	commandRunner := &events.DefaultCommandRunner{
		GlobalCfg: valid.NewGlobalCfgFromArgs(valid.GlobalCfgArgs{}),
	}

	configController := &controllers.ConfigController{
		APISecret:          []byte(apiSecret),
		Logger:             logger,
		CommandRunner:      commandRunner,
		RepoConfig:         "",
		RepoConfigJSON:     "",
		PolicyCheckEnabled: false,
		Scope:              scope,
	}

	// Create test request with wrong token
	reqBody := controllers.ConfigReloadRequest{}
	jsonBody, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", "/api/config/reload", bytes.NewBuffer(jsonBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Atlantis-Token", "wrong-secret")

	rr := httptest.NewRecorder()

	// Execute
	configController.ReloadConfig(rr, req)

	// Verify
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestConfigController_ReloadConfig_NoAuth(t *testing.T) {
	// Setup
	logger := logging.NewNoopLogger(t)
	scope := tally.NoopScope
	apiSecret := "test-secret"

	commandRunner := &events.DefaultCommandRunner{
		GlobalCfg: valid.NewGlobalCfgFromArgs(valid.GlobalCfgArgs{}),
	}

	configController := &controllers.ConfigController{
		APISecret:          []byte(apiSecret),
		Logger:             logger,
		CommandRunner:      commandRunner,
		RepoConfig:         "",
		RepoConfigJSON:     "",
		PolicyCheckEnabled: false,
		Scope:              scope,
	}

	// Create test request without token
	reqBody := controllers.ConfigReloadRequest{}
	jsonBody, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", "/api/config/reload", bytes.NewBuffer(jsonBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	// Execute
	configController.ReloadConfig(rr, req)

	// Verify
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}
