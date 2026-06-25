package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/go-playground/validator/v10"
	cfg "github.com/runatlantis/atlantis/server/core/config"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/logging"
	tally "github.com/uber-go/tally/v4"
)

const atlantisConfigTokenHeader = "X-Atlantis-Token"

// ConfigReloadRequest represents the request structure for config reload
type ConfigReloadRequest struct {
	// Force reload even if config hasn't changed
	Force bool `json:"force,omitempty"`
}

// ConfigReloadResponse represents the response structure for config reload
type ConfigReloadResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// ConfigController handles configuration management API endpoints
type ConfigController struct {
	APISecret       []byte
	Logger          logging.SimpleLogging        `validate:"required"`
	CommandRunner   *events.DefaultCommandRunner `validate:"required"`
	RepoConfig      string                       // Path to repo config file
	RepoConfigJSON  string                       // JSON config string
	ParserValidator *cfg.ParserValidator         `validate:"required"`
	Scope           tally.Scope                  `validate:"required"`

	// Mutex to ensure thread-safe config updates
	configMutex sync.RWMutex

	// Store the original args
	GlobalCfgArgs valid.GlobalCfgArgs
}

// ReloadConfig handles POST /api/config/reload
func (c *ConfigController) ReloadConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Validate API secret (same as other API endpoints)
	if len(c.APISecret) == 0 {
		c.respondError(w, http.StatusBadRequest, "API is disabled - no api-secret configured")
		return
	}

	secret := r.Header.Get(atlantisConfigTokenHeader)
	if secret != string(c.APISecret) {
		c.respondError(w, http.StatusUnauthorized, fmt.Sprintf("header %s did not match expected secret", atlantisConfigTokenHeader))
		return
	}

	// Parse request body
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		c.respondError(w, http.StatusBadRequest, "failed to read request body")
		return
	}

	var request ConfigReloadRequest
	if len(bytes) > 0 {
		if err = json.Unmarshal(bytes, &request); err != nil {
			c.respondError(w, http.StatusBadRequest, fmt.Sprintf("failed to parse request: %v", err.Error()))
			return
		}
		if err = validator.New().Struct(request); err != nil {
			c.respondError(w, http.StatusBadRequest, fmt.Sprintf("request validation failed: %v", err.Error()))
			return
		}
	}

	// Perform the configuration reload
	err = c.reloadConfiguration()
	if err != nil {
		c.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := ConfigReloadResponse{
		Success: true,
		Message: "Configuration reloaded successfully",
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		c.respondError(w, http.StatusInternalServerError, "failed to marshal response")
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(responseJSON)
	c.Logger.Info("Configuration reloaded successfully via API")
}

// reloadConfiguration performs the actual configuration reload
func (c *ConfigController) reloadConfiguration() error {
	c.configMutex.Lock()
	defer c.configMutex.Unlock()

	c.Logger.Info("Starting configuration reload")

	// Use the original args that were used at startup
	globalCfg := valid.NewGlobalCfgFromArgs(c.GlobalCfgArgs)

	// Parse the configuration using the same logic as server startup
	parserValidator := &cfg.ParserValidator{}
	var err error

	if c.RepoConfig != "" {
		c.Logger.Info("Reloading configuration from file: %s", c.RepoConfig)
		globalCfg, err = parserValidator.ParseGlobalCfg(c.RepoConfig, globalCfg)
		if err != nil {
			return fmt.Errorf("parsing %s file: %w", c.RepoConfig, err)
		}
	} else if c.RepoConfigJSON != "" {
		c.Logger.Info("Reloading configuration from JSON")
		globalCfg, err = parserValidator.ParseGlobalCfgJSON(c.RepoConfigJSON, globalCfg)
		if err != nil {
			return fmt.Errorf("parsing repo-config-json: %w", err)
		}
	} else {
		c.Logger.Info("No server-side configuration specified, using defaults")
	}

	// Atomically update the global configuration in the command runner
	c.CommandRunner.UpdateGlobalCfg(globalCfg)

	c.Logger.Info("Configuration reload completed successfully")
	return nil
}

// respondError sends an error response
func (c *ConfigController) respondError(w http.ResponseWriter, statusCode int, message string) {
	c.Logger.Warn("Config reload API error: %s", message)

	response := ConfigReloadResponse{
		Success: false,
		Error:   message,
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(statusCode)
	w.Write(responseJSON)
}
