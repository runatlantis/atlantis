// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package github_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/events/vcs/github"
	"github.com/runatlantis/atlantis/server/events/vcs/github/testdata"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func Test_githubTokenRotator_GenerateJob(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	defer disableSSLVerification()()
	testServer, err := testdata.GithubAppTestServer(t)
	Ok(t, err)

	anonCreds := &github.AnonymousCredentials{}
	anonClient, err := github.New(testServer, anonCreds, github.Config{}, 0, logging.NewNoopLogger(t))
	Ok(t, err)
	tempSecrets, err := anonClient.ExchangeCode(logger, "good-code")
	Ok(t, err)
	type fields struct {
		githubCredentials github.Credentials
	}
	tests := []struct {
		name             string
		fields           fields
		credsFileWritten bool
		wantErr          bool
	}{
		{
			name: "Should write .git-credentials file on start",
			fields: fields{&github.AppCredentials{
				AppID:    tempSecrets.ID,
				Key:      []byte(testdata.PrivateKey),
				Hostname: testServer,
			}},
			credsFileWritten: true,
			wantErr:          false,
		},
		{
			name: "Should return an error if pem data is missing or wrong",
			fields: fields{&github.AppCredentials{
				AppID:    tempSecrets.ID,
				Key:      []byte("some bad formatted pem key"),
				Hostname: testServer,
			}},
			credsFileWritten: false,
			wantErr:          true,
		},
		{
			name: "Should return an error if app id is missing or wrong",
			fields: fields{&github.AppCredentials{
				AppID:    3819,
				Key:      []byte(testdata.PrivateKey),
				Hostname: testServer,
			}},
			credsFileWritten: false,
			wantErr:          true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			t.Setenv("HOME", tmpDir)
			r := github.NewTokenRotator(logging.NewNoopLogger(t), tt.fields.githubCredentials, testServer, "x-access-token", tmpDir)
			got, err := r.GenerateJob()
			if (err != nil) != tt.wantErr {
				t.Errorf("githubTokenRotator.GenerateJob() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.credsFileWritten {
				credsFileContent := fmt.Sprintf(`https://x-access-token:some-token@%s`, testServer)
				actContents, err := os.ReadFile(filepath.Join(tmpDir, ".git-credentials"))
				Ok(t, err)
				Equals(t, credsFileContent, string(actContents))
			}
			Equals(t, 30*time.Second, got.Period)
		})
	}
}
