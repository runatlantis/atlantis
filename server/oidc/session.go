// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/runatlantis/atlantis/server/logging"
)

const (
	// sessionCookieName is the name of the HTTP cookie storing the OIDC session.
	sessionCookieName = "atlantis_oidc"

	// stateCookieName is the name of the HTTP cookie storing the OAuth2 state.
	stateCookieName = "atlantis_oidc_state"

	// sessionDuration is how long a session cookie remains valid.
	sessionDuration = 8 * time.Hour

	// stateDuration is how long an OAuth2 state parameter remains valid.
	stateDuration = 5 * time.Minute
)

// SessionManager handles OIDC session cookies and OAuth2 state parameters.
// Session cookies are AES-GCM encrypted to avoid storing PII (email) in
// plaintext. State cookies use HMAC-SHA256 signed JWTs.
type SessionManager struct {
	cookieSecret []byte
	encKey       []byte // 32-byte AES-256 key derived from cookieSecret
	secure       bool   // set Secure flag on cookies (requires HTTPS)
	basePath     string
	logger       logging.SimpleLogging
}

// NewSessionManager creates a new SessionManager. If cookieSecret is empty,
// a random 32-byte secret is generated (sessions won't survive restarts).
func NewSessionManager(cookieSecret []byte, secure bool, basePath string, logger logging.SimpleLogging) *SessionManager {
	if len(cookieSecret) == 0 {
		cookieSecret = make([]byte, 32)
		if _, err := rand.Read(cookieSecret); err != nil {
			// This should never happen; rand.Read uses /dev/urandom.
			panic("failed to generate random cookie secret: " + err.Error())
		}
		logger.Warn("no --web-oidc-cookie-secret provided, generated ephemeral secret. " +
			"OIDC sessions will not survive server restarts.")
	}
	// Derive a separate 32-byte AES key from the cookie secret so HMAC
	// signing (state cookies) and AES encryption (session cookies) use
	// independent keys.
	encKeyHash := sha256.Sum256(append([]byte("atlantis-session-enc:"), cookieSecret...))
	return &SessionManager{
		cookieSecret: cookieSecret,
		encKey:       encKeyHash[:],
		secure:       secure,
		basePath:     basePath,
		logger:       logger,
	}
}

func (s *SessionManager) cookiePath() string {
	bp := s.basePath
	if bp == "" || bp == "/" {
		return "/"
	}
	if bp[0] != '/' {
		bp = "/" + bp
	}
	return bp
}

func (s *SessionManager) clearCookie(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     s.cookiePath(),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   s.secure,
		SameSite: http.SameSiteLaxMode,
	})
}

// sessionPayload is the JSON structure encrypted inside the session cookie.
type sessionPayload struct {
	Email     string `json:"email"`
	ExpiresAt int64  `json:"exp"`
}

// encrypt encrypts plaintext using AES-256-GCM and returns a base64url-encoded
// ciphertext (nonce prepended).
func (s *SessionManager) encrypt(plaintext []byte) (string, error) {
	block, err := aes.NewCipher(s.encKey)
	if err != nil {
		return "", fmt.Errorf("creating AES cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("creating GCM: %w", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("generating nonce: %w", err)
	}
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return base64.RawURLEncoding.EncodeToString(ciphertext), nil
}

// decrypt decodes a base64url-encoded ciphertext and decrypts it using
// AES-256-GCM.
func (s *SessionManager) decrypt(encoded string) ([]byte, error) {
	data, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("decoding session cookie: %w", err)
	}
	block, err := aes.NewCipher(s.encKey)
	if err != nil {
		return nil, fmt.Errorf("creating AES cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("creating GCM: %w", err)
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("session cookie too short")
	}
	plaintext, err := gcm.Open(nil, data[:nonceSize], data[nonceSize:], nil)
	if err != nil {
		return nil, fmt.Errorf("decrypting session cookie: %w", err)
	}
	return plaintext, nil
}

// SetSession writes an HTTP-only session cookie containing an AES-GCM
// encrypted payload with the user's email and expiry. The email is never
// stored in plaintext in the cookie.
func (s *SessionManager) SetSession(w http.ResponseWriter, email string) error {
	payload := sessionPayload{
		Email:     email,
		ExpiresAt: time.Now().Add(sessionDuration).Unix(),
	}
	plaintext, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshaling session payload: %w", err)
	}
	encrypted, err := s.encrypt(plaintext)
	if err != nil {
		return fmt.Errorf("encrypting session: %w", err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    encrypted,
		Path:     s.cookiePath(),
		MaxAge:   int(sessionDuration.Seconds()),
		HttpOnly: true,
		Secure:   s.secure,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

// GetSession reads the session cookie, decrypts it, checks expiry, and
// returns the user's email. Returns an error if the cookie is missing,
// expired, or has been tampered with.
func (s *SessionManager) GetSession(r *http.Request) (string, error) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return "", fmt.Errorf("session cookie not found: %w", err)
	}

	plaintext, err := s.decrypt(cookie.Value)
	if err != nil {
		return "", fmt.Errorf("invalid session cookie: %w", err)
	}

	var payload sessionPayload
	if err := json.Unmarshal(plaintext, &payload); err != nil {
		return "", fmt.Errorf("invalid session payload: %w", err)
	}

	if time.Now().Unix() > payload.ExpiresAt {
		return "", errors.New("session expired")
	}

	return payload.Email, nil
}

// ClearSession deletes the session cookie.
func (s *SessionManager) ClearSession(w http.ResponseWriter) {
	s.clearCookie(w, sessionCookieName)
}

// CreateState generates a signed state JWT for OAuth2 CSRF protection.
// The state contains a random nonce and expires after 5 minutes.
// It also sets a state cookie for verification during callback.
func (s *SessionManager) CreateState(w http.ResponseWriter) (string, error) {
	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("generating state nonce: %w", err)
	}

	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   base64.RawURLEncoding.EncodeToString(nonce),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(stateDuration)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.cookieSecret)
	if err != nil {
		return "", fmt.Errorf("signing state token: %w", err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     stateCookieName,
		Value:    signed,
		Path:     s.cookiePath(),
		MaxAge:   int(stateDuration.Seconds()),
		HttpOnly: true,
		Secure:   s.secure,
		SameSite: http.SameSiteLaxMode,
	})

	return signed, nil
}

// VerifyState validates the state parameter from the OAuth2 callback by
// comparing it to the state cookie.
func (s *SessionManager) VerifyState(r *http.Request, state string) error {
	cookie, err := r.Cookie(stateCookieName)
	if err != nil {
		return fmt.Errorf("state cookie not found: %w", err)
	}

	if cookie.Value != state {
		return errors.New("state parameter does not match state cookie")
	}

	// Also verify the JWT is valid and not expired.
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(state, claims, func(_ *jwt.Token) (any, error) {
		return s.cookieSecret, nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil {
		return fmt.Errorf("invalid state token: %w", err)
	}
	if !token.Valid {
		return errors.New("state token is not valid")
	}

	return nil
}

// ClearState deletes the state cookie after callback processing.
func (s *SessionManager) ClearState(w http.ResponseWriter) {
	s.clearCookie(w, stateCookieName)
}
