// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

type testCertificateOptions struct {
	commonName string
	dnsNames   []string
	notBefore  time.Time
	notAfter   time.Time
}

func TestValidateTLSCertificate(t *testing.T) {
	now := time.Date(2026, 3, 12, 0, 0, 0, 0, time.UTC)

	cases := []struct {
		name            string
		hostname        string
		createFiles     func(t *testing.T, dir string) (string, string)
		expectedErr     string
		expectedWarning []string
	}{
		{
			name:     "valid cert and key with matching SAN",
			hostname: "atlantis.example.com",
			createFiles: func(t *testing.T, dir string) (string, string) {
				certPEM, keyPEM, _ := newSelfSignedCertificatePEM(t, testCertificateOptions{
					commonName: "ignored-cn.example.com",
					dnsNames:   []string{"atlantis.example.com"},
					notBefore:  now.Add(-time.Hour),
					notAfter:   now.Add(24 * time.Hour),
				})
				return writeCertificateFiles(t, dir, "valid", certPEM, keyPEM)
			},
		},
		{
			name:     "expired cert returns warning",
			hostname: "atlantis.example.com",
			createFiles: func(t *testing.T, dir string) (string, string) {
				certPEM, keyPEM, _ := newSelfSignedCertificatePEM(t, testCertificateOptions{
					commonName: "atlantis.example.com",
					dnsNames:   []string{"atlantis.example.com"},
					notBefore:  now.Add(-48 * time.Hour),
					notAfter:   now.Add(-24 * time.Hour),
				})
				return writeCertificateFiles(t, dir, "expired", certPEM, keyPEM)
			},
			expectedWarning: []string{"expired"},
		},
		{
			name:     "not yet valid cert returns warning",
			hostname: "atlantis.example.com",
			createFiles: func(t *testing.T, dir string) (string, string) {
				certPEM, keyPEM, _ := newSelfSignedCertificatePEM(t, testCertificateOptions{
					commonName: "atlantis.example.com",
					dnsNames:   []string{"atlantis.example.com"},
					notBefore:  now.Add(24 * time.Hour),
					notAfter:   now.Add(48 * time.Hour),
				})
				return writeCertificateFiles(t, dir, "not-yet-valid", certPEM, keyPEM)
			},
			expectedWarning: []string{"not yet valid", "not_before="},
		},
		{
			name:     "cert and key mismatch returns error",
			hostname: "atlantis.example.com",
			createFiles: func(t *testing.T, dir string) (string, string) {
				certPEM, _, _ := newSelfSignedCertificatePEM(t, testCertificateOptions{
					commonName: "atlantis.example.com",
					dnsNames:   []string{"atlantis.example.com"},
					notBefore:  now.Add(-time.Hour),
					notAfter:   now.Add(24 * time.Hour),
				})
				_, wrongKeyPEM, _ := newSelfSignedCertificatePEM(t, testCertificateOptions{
					commonName: "different.example.com",
					dnsNames:   []string{"different.example.com"},
					notBefore:  now.Add(-time.Hour),
					notAfter:   now.Add(24 * time.Hour),
				})
				return writeCertificateFiles(t, dir, "mismatch", certPEM, wrongKeyPEM)
			},
			expectedErr: "loading tls cert and key pair",
		},
		{
			name:     "encrypted key returns error",
			hostname: "atlantis.example.com",
			createFiles: func(t *testing.T, dir string) (string, string) {
				certPEM, _, keyDER := newSelfSignedCertificatePEM(t, testCertificateOptions{
					commonName: "atlantis.example.com",
					dnsNames:   []string{"atlantis.example.com"},
					notBefore:  now.Add(-time.Hour),
					notAfter:   now.Add(24 * time.Hour),
				})
				encryptedBlock, err := x509.EncryptPEMBlock(rand.Reader, "PRIVATE KEY", keyDER, []byte("passphrase"), x509.PEMCipherAES256)
				Ok(t, err)
				encryptedKey := pem.EncodeToMemory(encryptedBlock)
				return writeCertificateFiles(t, dir, "encrypted", certPEM, encryptedKey)
			},
			expectedErr: "does not support encrypted keys",
		},
		{
			name:     "hostname mismatch warns and SAN is preferred over CN",
			hostname: "atlantis.example.com",
			createFiles: func(t *testing.T, dir string) (string, string) {
				certPEM, keyPEM, _ := newSelfSignedCertificatePEM(t, testCertificateOptions{
					commonName: "atlantis.example.com",
					dnsNames:   []string{"other.example.com"},
					notBefore:  now.Add(-time.Hour),
					notAfter:   now.Add(24 * time.Hour),
				})
				return writeCertificateFiles(t, dir, "hostname-mismatch", certPEM, keyPEM)
			},
			expectedWarning: []string{"san entries do not match atlantis hostname"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			certFile, keyFile := tc.createFiles(t, t.TempDir())
			validationResult, err := validateTLSCertificate(certFile, keyFile, tc.hostname, now)

			if tc.expectedErr != "" {
				ErrContains(t, tc.expectedErr, err)
				return
			}

			Ok(t, err)

			Assert(t, len(tc.expectedWarning) != 0 || len(validationResult.Warnings) == 0, "expected no warnings, got %v", validationResult.Warnings)
			for _, expectedWarningSubstring := range tc.expectedWarning {
				Assert(t, containsSubstring(validationResult.Warnings, expectedWarningSubstring), "expected warning containing %q, got %v", expectedWarningSubstring, validationResult.Warnings)
			}
		})
	}
}

func TestGetSSLCertificate_ReloadErrorIsReturned(t *testing.T) {
	now := time.Date(2026, 3, 12, 0, 0, 0, 0, time.UTC)
	initialCert, initialKey, _ := newSelfSignedCertificatePEM(t, testCertificateOptions{
		commonName: "atlantis.example.com",
		dnsNames:   []string{"atlantis.example.com"},
		notBefore:  now.Add(-time.Hour),
		notAfter:   now.Add(24 * time.Hour),
	})
	rotatedCert, _, _ := newSelfSignedCertificatePEM(t, testCertificateOptions{
		commonName: "atlantis.example.com",
		dnsNames:   []string{"atlantis.example.com"},
		notBefore:  now.Add(-time.Hour),
		notAfter:   now.Add(48 * time.Hour),
	})
	_, mismatchedKey, _ := newSelfSignedCertificatePEM(t, testCertificateOptions{
		commonName: "different.example.com",
		dnsNames:   []string{"different.example.com"},
		notBefore:  now.Add(-time.Hour),
		notAfter:   now.Add(48 * time.Hour),
	})

	dir := t.TempDir()
	certFile, keyFile := writeCertificateFiles(t, dir, "reload", initialCert, initialKey)

	logger := logging.NewNoopLogger(t).WithHistory()
	atlantisURL := mustParseURL(t, "https://atlantis.example.com")
	s := &Server{
		AtlantisURL: atlantisURL,
		SSLCertFile: certFile,
		SSLKeyFile:  keyFile,
		Logger:      logger,
	}

	_, err := s.GetSSLCertificate(nil)
	Ok(t, err)

	Ok(t, os.WriteFile(certFile, rotatedCert, 0600))
	Ok(t, os.WriteFile(keyFile, mismatchedKey, 0600))
	s.CertLastRefreshTime = s.CertLastRefreshTime.Add(-time.Second)
	s.KeyLastRefreshTime = s.KeyLastRefreshTime.Add(-time.Second)

	_, err = s.GetSSLCertificate(nil)
	Assert(t, err != nil, "expected reload error, got nil")
	ErrContains(t, "while loading tls cert", err)

	history := logger.GetHistory()
	Assert(t, !strings.Contains(history, "tls certificate reload"), "unexpected explicit reload error log in history, got: %s", history)
}

func TestGetSSLCertificate_ReloadWarningIsLogged(t *testing.T) {
	now := time.Date(2026, 3, 12, 0, 0, 0, 0, time.UTC)
	initialCert, initialKey, _ := newSelfSignedCertificatePEM(t, testCertificateOptions{
		commonName: "atlantis.example.com",
		dnsNames:   []string{"atlantis.example.com"},
		notBefore:  now.Add(-time.Hour),
		notAfter:   now.Add(24 * time.Hour),
	})
	expiredCert, expiredKey, _ := newSelfSignedCertificatePEM(t, testCertificateOptions{
		commonName: "atlantis.example.com",
		dnsNames:   []string{"atlantis.example.com"},
		notBefore:  now.Add(-48 * time.Hour),
		notAfter:   now.Add(-24 * time.Hour),
	})

	dir := t.TempDir()
	certFile, keyFile := writeCertificateFiles(t, dir, "reload-warning", initialCert, initialKey)

	logger := logging.NewNoopLogger(t).WithHistory()
	s := &Server{
		AtlantisURL: mustParseURL(t, "https://atlantis.example.com"),
		SSLCertFile: certFile,
		SSLKeyFile:  keyFile,
		Logger:      logger,
	}

	_, err := s.GetSSLCertificate(nil)
	Ok(t, err)

	Ok(t, os.WriteFile(certFile, expiredCert, 0600))
	Ok(t, os.WriteFile(keyFile, expiredKey, 0600))
	s.CertLastRefreshTime = s.CertLastRefreshTime.Add(-time.Second)
	s.KeyLastRefreshTime = s.KeyLastRefreshTime.Add(-time.Second)

	_, err = s.GetSSLCertificate(nil)
	Ok(t, err)

	history := logger.GetHistory()
	Assert(t, strings.Contains(history, "tls certificate warning,") && strings.Contains(history, "expired"), "expected expiry warning in log history, got: %s", history)
}

func TestStart_InvalidTLSConfigFailsFast(t *testing.T) {
	now := time.Date(2026, 3, 12, 0, 0, 0, 0, time.UTC)
	certPEM, _, _ := newSelfSignedCertificatePEM(t, testCertificateOptions{
		commonName: "atlantis.example.com",
		dnsNames:   []string{"atlantis.example.com"},
		notBefore:  now.Add(-time.Hour),
		notAfter:   now.Add(24 * time.Hour),
	})
	_, wrongKeyPEM, _ := newSelfSignedCertificatePEM(t, testCertificateOptions{
		commonName: "different.example.com",
		dnsNames:   []string{"different.example.com"},
		notBefore:  now.Add(-time.Hour),
		notAfter:   now.Add(24 * time.Hour),
	})
	certFile, keyFile := writeCertificateFiles(t, t.TempDir(), "startup-fail-fast", certPEM, wrongKeyPEM)

	logger := logging.NewNoopLogger(t).WithHistory()
	s := &Server{
		AtlantisURL:            mustParseURL(t, "https://atlantis.example.com"),
		Router:                 mux.NewRouter(),
		Port:                   4141,
		SSLCertFile:            certFile,
		SSLKeyFile:             keyFile,
		Logger:                 logger,
		DisableGlobalApplyLock: true,
		EnableProfilingAPI:     false,
		WebAuthentication:      false,
	}

	err := s.Start()
	Assert(t, err != nil, "expected tls startup validation error, got nil")
	ErrContains(t, "invalid tls configuration", err)
}

func TestStdlibServerErrorLogWriter_Write(t *testing.T) {
	logger := logging.NewNoopLogger(t).WithHistory()
	writer := &stdlibServerErrorLogWriter{logger: logger}

	message := "http: TLS handshake error from 127.0.0.1:12345: remote error: tls: bad certificate\n"
	n, err := writer.Write([]byte(message))
	Ok(t, err)
	Equals(t, len(message), n)

	history := logger.GetHistory()
	Assert(t, strings.Contains(history, "http server error"), "expected bridged stdlib server error in history, got: %s", history)
}

func newSelfSignedCertificatePEM(t *testing.T, options testCertificateOptions) ([]byte, []byte, []byte) {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	Ok(t, err)

	keyDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	Ok(t, err)
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER})

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	Ok(t, err)

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Atlantis TLS test suite"},
			CommonName:   options.commonName,
		},
		NotBefore:   options.notBefore,
		NotAfter:    options.notAfter,
		DNSNames:    options.dnsNames,
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	Ok(t, err)
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	return certPEM, keyPEM, keyDER
}

func writeCertificateFiles(t *testing.T, dir string, prefix string, certPEM []byte, keyPEM []byte) (string, string) {
	t.Helper()

	certFile := filepath.Join(dir, fmt.Sprintf("%s-cert.pem", prefix))
	keyFile := filepath.Join(dir, fmt.Sprintf("%s-key.pem", prefix))

	Ok(t, os.WriteFile(certFile, certPEM, 0600))
	Ok(t, os.WriteFile(keyFile, keyPEM, 0600))
	return certFile, keyFile
}

func mustParseURL(t *testing.T, rawURL string) *url.URL {
	t.Helper()
	parsed, err := ParseAtlantisURL(rawURL)
	Ok(t, err)
	return parsed
}

func containsSubstring(messages []string, substring string) bool {
	for _, message := range messages {
		if strings.Contains(message, substring) {
			return true
		}
	}
	return false
}
