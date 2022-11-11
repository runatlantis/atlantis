package cli_test

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/palantir/go-githubapp/githubapp"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/file"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/github/cli"
	"github.com/stretchr/testify/assert"
	"go.temporal.io/sdk/testsuite"
)

type testInstallationTransport struct {
	expectedAppID          int64
	expectedInstallationID int64
	expectedPrivateKey     string
	t                      *testing.T
	transport              testTransport
	numCalls               int
}

type testTransport struct {
	token       string
	expectedErr error
}

func (t testTransport) Token(ctx context.Context) (string, error) {
	return t.token, t.expectedErr
}

func (t *testInstallationTransport) New(tr http.RoundTripper, appID, installationID int64, privateKey []byte) (cli.Transport, error) {
	assert.Equal(t.t, t.expectedAppID, appID)
	assert.Equal(t.t, t.expectedInstallationID, installationID)
	assert.Equal(t.t, t.expectedPrivateKey, string(privateKey))

	t.numCalls++

	return t.transport, nil
}
func TestRefresh(t *testing.T) {
	appID := int64(1234)
	installationID := int64(4567)
	privateKey := "somekey"
	token := "70897098"

	cfg := githubapp.Config{
		App: struct {
			IntegrationID int64  "yaml:\"integration_id\" json:\"integrationId\""
			WebhookSecret string "yaml:\"webhook_secret\" json:\"webhookSecret\""
			PrivateKey    string "yaml:\"private_key\" json:\"privateKey\""
		}{
			IntegrationID: appID,
			PrivateKey:    privateKey,
		},
	}
	transportCreator := &testInstallationTransport{
		expectedAppID:          appID,
		expectedInstallationID: installationID,
		expectedPrivateKey:     privateKey,
		t:                      t,
		transport: testTransport{
			token: token,
		},
	}

	t.Run("first write", func(t *testing.T) {
		dir := t.TempDir()

		capturedGitArgs := [][]string{}
		subject := cli.Credentials{
			HomeDir:          dir,
			TransportCreator: transportCreator,
			FileLock:         &file.RWLock{},
			Cfg:              cfg,
			Git: func(s ...string) error {
				capturedGitArgs = append(capturedGitArgs, s)
				return nil
			},
		}

		activity := func(ctx context.Context) error {
			return subject.Refresh(ctx, installationID)
		}

		ts := testsuite.WorkflowTestSuite{}
		env := ts.NewTestActivityEnvironment()

		env.RegisterActivity(activity)
		_, err := env.ExecuteActivity(activity)
		assert.NoError(t, err)

		raw, err := os.ReadFile(filepath.Join(dir, ".git-credentials"))
		assert.NoError(t, err)

		assert.Equal(t, "https://x-access-token:70897098@github.com", string(raw))
		assert.Equal(t, [][]string{
			{
				"config", "--global", "credential.helper", "store",
			},
			{
				"config", "--global", "url.https://x-access-token@github.com.insteadOf", "ssh://git@github.com",
			},
		}, capturedGitArgs)
	})

	t.Run("writes new token", func(t *testing.T) {
		dir := t.TempDir()

		credentialsFile := filepath.Join(dir, ".git-credentials")
		oldContents := "https://x-access-token:123456@github.com"

		err := os.WriteFile(credentialsFile, []byte(oldContents), os.ModePerm)
		assert.NoError(t, err)

		capturedGitArgs := [][]string{}
		subject := cli.Credentials{
			HomeDir:          dir,
			TransportCreator: transportCreator,
			FileLock:         &file.RWLock{},
			Cfg:              cfg,
			Git: func(s ...string) error {
				capturedGitArgs = append(capturedGitArgs, s)
				return nil
			},
		}

		activity := func(ctx context.Context) error {
			return subject.Refresh(ctx, installationID)
		}

		ts := testsuite.WorkflowTestSuite{}
		env := ts.NewTestActivityEnvironment()

		env.RegisterActivity(activity)
		_, err = env.ExecuteActivity(activity)
		assert.NoError(t, err)

		raw, err := os.ReadFile(credentialsFile)
		assert.NoError(t, err)

		assert.Equal(t, "https://x-access-token:70897098@github.com", string(raw))
		assert.Equal(t, [][]string{
			{
				"config", "--global", "credential.helper", "store",
			},
			{
				"config", "--global", "url.https://x-access-token@github.com.insteadOf", "ssh://git@github.com",
			},
		}, capturedGitArgs)

	})

	t.Run("only one transport created", func(t *testing.T) {
		dir := t.TempDir()

		credentialsFile := filepath.Join(dir, ".git-credentials")
		oldContents := "https://x-access-token:123456@github.com"

		err := os.WriteFile(credentialsFile, []byte(oldContents), os.ModePerm)
		assert.NoError(t, err)

		tc := &testInstallationTransport{
			expectedAppID:          appID,
			expectedInstallationID: installationID,
			expectedPrivateKey:     privateKey,
			t:                      t,
			transport: testTransport{
				token: token,
			},
		}

		subject := cli.Credentials{
			HomeDir:          dir,
			TransportCreator: tc,
			FileLock:         &file.RWLock{},
			Cfg:              cfg,
			Git: func(s ...string) error {
				return nil
			},
		}

		activity := func(ctx context.Context) error {
			err := subject.Refresh(ctx, installationID)

			if err != nil {
				return err
			}

			return subject.Refresh(ctx, installationID)
		}

		ts := testsuite.WorkflowTestSuite{}
		env := ts.NewTestActivityEnvironment()

		env.RegisterActivity(activity)
		_, err = env.ExecuteActivity(activity)
		assert.NoError(t, err)

		raw, err := os.ReadFile(credentialsFile)
		assert.NoError(t, err)

		assert.Equal(t, "https://x-access-token:70897098@github.com", string(raw))
		assert.Equal(t, tc.numCalls, 1)

	})

}
