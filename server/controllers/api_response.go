// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/runatlantis/atlantis/server/logging"
)

// ErrorCode represents a machine-readable error code for API responses.
// Clients can use these codes for programmatic error handling.
type ErrorCode string

const (
	// ErrCodeValidation indicates request validation failed.
	ErrCodeValidation ErrorCode = "VALIDATION_ERROR"
	// ErrCodeUnauthorized indicates missing or invalid authentication.
	ErrCodeUnauthorized ErrorCode = "UNAUTHORIZED"
	// ErrCodeForbidden indicates the request is not allowed.
	ErrCodeForbidden ErrorCode = "FORBIDDEN"
	// ErrCodeNotFound indicates the requested resource was not found.
	ErrCodeNotFound ErrorCode = "NOT_FOUND"
	// ErrCodeConflict indicates a conflict with current state.
	ErrCodeConflict ErrorCode = "CONFLICT"
	// ErrCodeInternal indicates an internal server error.
	ErrCodeInternal ErrorCode = "INTERNAL_ERROR"
	// ErrCodeServiceUnavailable indicates a required service is not available.
	ErrCodeServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
	// ErrCodeRateLimited indicates too many requests.
	ErrCodeRateLimited ErrorCode = "RATE_LIMITED"
)

// APIError represents a structured error response.
type APIError struct {
	// Code is a machine-readable error code.
	Code ErrorCode `json:"code"`
	// Message is a human-readable error message.
	Message string `json:"message"`
	// Details contains additional error context (e.g., field-level validation errors).
	Details any `json:"details,omitempty"`
}

// Error implements the error interface.
func (e *APIError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// NewAPIError creates a new APIError.
func NewAPIError(code ErrorCode, message string) *APIError {
	return &APIError{Code: code, Message: message}
}

// WithDetails adds details to an APIError and returns it.
func (e *APIError) WithDetails(details any) *APIError {
	e.Details = details
	return e
}

// ValidationError represents field-level validation errors.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// NewValidationError creates a validation APIError with field details.
func NewValidationError(message string, fields ...ValidationError) *APIError {
	err := NewAPIError(ErrCodeValidation, message)
	if len(fields) > 0 {
		err.Details = fields
	}
	return err
}

// APIResponse is the standard envelope for all API responses.
// Using a concrete type instead of generics for Go 1.18+ compatibility with json.Marshal.
type APIResponse struct {
	// Success indicates whether the request succeeded.
	Success bool `json:"success"`
	// Data contains the response payload on success.
	Data any `json:"data,omitempty"`
	// Error contains error details on failure.
	Error *APIError `json:"error,omitempty"`
	// RequestID is a unique identifier for request tracing.
	RequestID string `json:"request_id"`
	// Timestamp is when the response was generated.
	Timestamp time.Time `json:"timestamp"`
}

// NewSuccessResponse creates a successful API response.
func NewSuccessResponse(requestID string, data any) *APIResponse {
	return &APIResponse{
		Success:   true,
		Data:      data,
		RequestID: requestID,
		Timestamp: time.Now(),
	}
}

// NewErrorResponse creates an error API response.
func NewErrorResponse(requestID string, err *APIError) *APIResponse {
	return &APIResponse{
		Success:   false,
		Error:     err,
		RequestID: requestID,
		Timestamp: time.Now(),
	}
}

// APIResponder provides methods for sending consistent API responses.
type APIResponder struct {
	Logger logging.SimpleLogging
}

// NewAPIResponder creates a new APIResponder.
func NewAPIResponder(logger logging.SimpleLogging) *APIResponder {
	return &APIResponder{Logger: logger}
}

// GenerateRequestID creates a new unique request ID.
func GenerateRequestID() string {
	return uuid.New().String()
}

// GetRequestID extracts the request ID from the request context or header,
// or generates a new one if not present.
func GetRequestID(r *http.Request) string {
	// Check for existing request ID in header (for distributed tracing)
	if id := r.Header.Get("X-Request-ID"); id != "" {
		return id
	}
	return GenerateRequestID()
}

// Success sends a successful JSON response.
func (a *APIResponder) Success(w http.ResponseWriter, r *http.Request, code int, data any) {
	requestID := GetRequestID(r)
	response := NewSuccessResponse(requestID, data)
	a.writeJSON(w, code, response)
	a.Logger.Info("request_id=%s status=%d", requestID, code)
}

// Error sends an error JSON response.
func (a *APIResponder) Error(w http.ResponseWriter, r *http.Request, httpCode int, apiErr *APIError) {
	requestID := GetRequestID(r)
	response := NewErrorResponse(requestID, apiErr)
	a.writeJSON(w, httpCode, response)
	a.Logger.Warn("request_id=%s status=%d error=%s message=%s", requestID, httpCode, apiErr.Code, apiErr.Message)
}

// ValidationFailed sends a validation error response.
func (a *APIResponder) ValidationFailed(w http.ResponseWriter, r *http.Request, message string, fields ...ValidationError) {
	a.Error(w, r, http.StatusBadRequest, NewValidationError(message, fields...))
}

// Unauthorized sends an unauthorized error response.
func (a *APIResponder) Unauthorized(w http.ResponseWriter, r *http.Request, message string) {
	a.Error(w, r, http.StatusUnauthorized, NewAPIError(ErrCodeUnauthorized, message))
}

// Forbidden sends a forbidden error response.
func (a *APIResponder) Forbidden(w http.ResponseWriter, r *http.Request, message string) {
	a.Error(w, r, http.StatusForbidden, NewAPIError(ErrCodeForbidden, message))
}

// NotFound sends a not found error response.
func (a *APIResponder) NotFound(w http.ResponseWriter, r *http.Request, message string) {
	a.Error(w, r, http.StatusNotFound, NewAPIError(ErrCodeNotFound, message))
}

// InternalError sends an internal server error response.
// The full error is logged server-side but only a generic message is returned to the client.
func (a *APIResponder) InternalError(w http.ResponseWriter, r *http.Request, err error) {
	a.Logger.Err("internal error [%s %s]: %v", r.Method, r.URL.Path, err)
	a.Error(w, r, http.StatusInternalServerError, NewAPIError(ErrCodeInternal, "internal server error"))
}

// ServiceUnavailable sends a service unavailable error response.
func (a *APIResponder) ServiceUnavailable(w http.ResponseWriter, r *http.Request, message string) {
	a.Error(w, r, http.StatusServiceUnavailable, NewAPIError(ErrCodeServiceUnavailable, message))
}

func (a *APIResponder) writeJSON(w http.ResponseWriter, code int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		a.Logger.Err("failed to encode JSON response: %v", err)
	}
}

// APIMiddleware provides common middleware functions for API endpoints.
type APIMiddleware struct {
	APISecret []byte
	Logger    logging.SimpleLogging
	Responder *APIResponder
}

// NewAPIMiddleware creates a new APIMiddleware.
func NewAPIMiddleware(apiSecret []byte, logger logging.SimpleLogging) *APIMiddleware {
	return &APIMiddleware{
		APISecret: apiSecret,
		Logger:    logger,
		Responder: NewAPIResponder(logger),
	}
}

// RequireAuth is middleware that validates the API secret token.
// Returns true if authentication passed, false if it failed (response already sent).
func (m *APIMiddleware) RequireAuth(w http.ResponseWriter, r *http.Request) bool {
	if len(m.APISecret) == 0 {
		m.Responder.Error(w, r, http.StatusServiceUnavailable,
			NewAPIError(ErrCodeServiceUnavailable, "API is disabled"))
		return false
	}

	secret := r.Header.Get(atlantisTokenHeader)
	if subtle.ConstantTimeCompare([]byte(secret), m.APISecret) != 1 {
		m.Responder.Unauthorized(w, r, "invalid or missing API token")
		return false
	}

	return true
}

// SetJSONContentType sets the Content-Type header for JSON responses.
func SetJSONContentType(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}
