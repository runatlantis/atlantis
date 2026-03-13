// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/runatlantis/atlantis/server/logging"
)

type tlsValidationResult struct {
	Certificate tls.Certificate
	Warnings    []string
}

func (s *Server) loadAndValidateSSLCertificate() (*tls.Certificate, time.Time, time.Time, []string, error) {
	certStat, err := os.Stat(s.SSLCertFile)
	if err != nil {
		return nil, time.Time{}, time.Time{}, nil, fmt.Errorf("while getting cert file modification time: %w", err)
	}

	keyStat, err := os.Stat(s.SSLKeyFile)
	if err != nil {
		return nil, time.Time{}, time.Time{}, nil, fmt.Errorf("while getting key file modification time: %w", err)
	}

	validationResult, err := validateTLSCertificate(s.SSLCertFile, s.SSLKeyFile, s.tlsValidationHostname(), time.Now())
	if err != nil {
		return nil, time.Time{}, time.Time{}, nil, err
	}

	return &validationResult.Certificate, certStat.ModTime(), keyStat.ModTime(), validationResult.Warnings, nil
}

func (s *Server) logTLSWarnings(warnings []string) {
	if s.Logger == nil {
		return
	}
	for _, warning := range warnings {
		s.Logger.Warn("tls certificate warning, %s", warning)
	}
}

func (s *Server) tlsValidationHostname() string {
	if s.AtlantisURL == nil {
		return ""
	}
	return s.AtlantisURL.Hostname()
}

func validateTLSCertificate(certFile string, keyFile string, hostname string, now time.Time) (*tlsValidationResult, error) {
	keyPEMBytes, err := os.ReadFile(keyFile)
	if err != nil {
		return nil, fmt.Errorf("reading tls key file: %w", err)
	}

	if isEncryptedPrivateKeyPEM(keyPEMBytes) {
		return nil, fmt.Errorf("tls private key is encrypted, atlantis does not support encrypted keys because no passphrase can be configured")
	}

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("loading tls cert and key pair: %w", err)
	}

	if len(cert.Certificate) == 0 {
		return nil, fmt.Errorf("tls certificate did not include a leaf certificate")
	}

	leafCertificate, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return nil, fmt.Errorf("parsing tls leaf certificate: %w", err)
	}

	var warnings []string
	if now.Before(leafCertificate.NotBefore) {
		warnings = append(warnings, fmt.Sprintf("certificate %q is not yet valid, not_before=%s", certFile, leafCertificate.NotBefore.UTC().Format(time.RFC3339)))
	}
	if now.After(leafCertificate.NotAfter) {
		warnings = append(warnings, fmt.Sprintf("certificate %q is expired, not_after=%s", certFile, leafCertificate.NotAfter.UTC().Format(time.RFC3339)))
	}

	if hostnameWarning := validateCertificateHostname(leafCertificate, hostname, certFile); hostnameWarning != "" {
		warnings = append(warnings, hostnameWarning)
	}

	return &tlsValidationResult{
		Certificate: cert,
		Warnings:    warnings,
	}, nil
}

func isEncryptedPrivateKeyPEM(keyPEMBytes []byte) bool {
	remaining := keyPEMBytes
	for {
		block, rest := pem.Decode(remaining)
		if block == nil {
			return false
		}
		remaining = rest

		blockType := strings.ToUpper(strings.TrimSpace(block.Type))
		if blockType == "ENCRYPTED PRIVATE KEY" {
			return true
		}
		if strings.Contains(blockType, "PRIVATE KEY") && x509.IsEncryptedPEMBlock(block) {
			return true
		}
	}
}

func validateCertificateHostname(certificate *x509.Certificate, hostname string, certFile string) string {
	hostname = strings.TrimSuffix(strings.TrimSpace(hostname), ".")
	if hostname == "" {
		return ""
	}

	if len(certificate.DNSNames) > 0 || len(certificate.IPAddresses) > 0 {
		if err := certificate.VerifyHostname(hostname); err != nil {
			return fmt.Sprintf("certificate %q san entries do not match atlantis hostname %q", certFile, hostname)
		}
		return ""
	}

	commonName := strings.TrimSuffix(strings.TrimSpace(certificate.Subject.CommonName), ".")
	if commonName == "" {
		return fmt.Sprintf("certificate %q has no san entries and no common name, so it cannot be matched against atlantis hostname %q", certFile, hostname)
	}
	if commonNameMatchesHostname(commonName, hostname) {
		return ""
	}

	return fmt.Sprintf("certificate %q common name %q does not match atlantis hostname %q", certFile, commonName, hostname)
}

func commonNameMatchesHostname(commonName string, hostname string) bool {
	commonName = strings.ToLower(commonName)
	hostname = strings.ToLower(hostname)

	if commonName == hostname {
		return true
	}

	if net.ParseIP(hostname) != nil {
		return false
	}

	if strings.HasPrefix(commonName, "*.") {
		suffix := strings.TrimPrefix(commonName, "*")
		if !strings.HasSuffix(hostname, suffix) {
			return false
		}
		// Wildcard only matches a single left-most label.
		return strings.Count(hostname, ".") == strings.Count(strings.TrimPrefix(suffix, "."), ".")+1
	}
	return false
}

type stdlibServerErrorLogWriter struct {
	logger logging.SimpleLogging
}

func (w *stdlibServerErrorLogWriter) Write(p []byte) (int, error) {
	message := strings.TrimSpace(string(p))
	if message != "" && w.logger != nil {
		w.logger.Err("http server error, %q", message)
	}
	return len(p), nil
}
