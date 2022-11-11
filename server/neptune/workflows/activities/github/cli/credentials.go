package cli

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/mitchellh/go-homedir"
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/neptune/logger"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/file"
)

type ghInstallationTransportCreator struct{}

func (t ghInstallationTransportCreator) New(tr http.RoundTripper, appID, installationID int64, privateKey []byte) (Transport, error) {
	return ghinstallation.New(tr, appID, installationID, privateKey)
}

type Transport interface {
	Token(ctx context.Context) (string, error)
}

type transportCreator interface {
	New(tr http.RoundTripper, appID, installationID int64, privateKey []byte) (Transport, error)
}

type Credentials struct {
	Cfg              githubapp.Config
	FileLock         *file.RWLock
	TransportCreator transportCreator
	HomeDir          string
	Git              func(...string) error

	once      sync.Once
	transport Transport
}

func NewCredentials(cfg githubapp.Config, fileLock *file.RWLock) (*Credentials, error) {
	home, err := homedir.Dir()
	if err != nil {
		return nil, errors.Wrap(err, "getting home dir")
	}
	return &Credentials{
		HomeDir:          home,
		TransportCreator: ghInstallationTransportCreator{},
		Cfg:              cfg,
		FileLock:         fileLock,
		Git:              git,
	}, nil
}

func (c *Credentials) Refresh(ctx context.Context, installationID int64) error {
	// initialize our transport once here. We don't support multiple installation ids atm
	// since we are using a global git config
	var initErr error
	c.once.Do(func() {
		transport, err := c.TransportCreator.New(http.DefaultTransport, c.Cfg.App.IntegrationID, installationID, []byte(c.Cfg.App.PrivateKey))
		if err != nil {
			initErr = err
			return
		}
		c.transport = transport
	})

	if initErr != nil {
		return errors.Wrap(initErr, "initializing transport")
	}

	token, err := c.transport.Token(ctx)
	if err != nil {
		return errors.Wrap(err, "refreshing token in transport")
	}

	return errors.Wrap(
		c.writeCredentials(ctx, filepath.Join(c.HomeDir, ".git-credentials"), token),
		"writing credentials",
	)
}

func (c *Credentials) safeWriteFile(file string, contents []byte, perm os.FileMode) error {
	c.FileLock.Lock()
	defer c.FileLock.Unlock()

	return errors.Wrap(
		os.WriteFile(file, contents, perm),
		"writing file",
	)
}

func (c *Credentials) safeReadFile(file string) (string, error) {
	c.FileLock.RLock()
	defer c.FileLock.RUnlock()

	contents, err := os.ReadFile(file)
	if err != nil {
		return "", errors.Wrap(err, "reading file")
	}

	// for some reason this gets read in with an additional new line. Maybe git config
	// might be responsible :shrug
	return strings.TrimSuffix(string(contents), "\n"), nil

}

func (c *Credentials) writeConfig(file string, contents []byte) error {
	if err := c.safeWriteFile(file, contents, os.ModePerm); err != nil {
		return err
	}
	if err := c.Git("config", "--global", "credential.helper", "store"); err != nil {
		return err
	}

	return c.Git("config", "--global", "url.https://x-access-token@github.com.insteadOf", "ssh://git@github.com")
}

func (c *Credentials) writeCredentials(ctx context.Context, file string, token string) error {
	toWrite := fmt.Sprintf(`https://x-access-token:%s@github.com`, token)

	// if it doesn't exist write to file
	if _, err := os.Stat(file); err != nil {
		logger.Info(ctx, "writing global .git-credentials file")
		return c.writeConfig(file, []byte(toWrite))
	}

	contents, err := c.safeReadFile(file)
	if err != nil {
		return errors.Wrap(err, "reading existing credentials")
	}

	// our token was refreshed so let's write it
	if contents != toWrite {
		logger.Info(ctx, "token was refreshed, rewriting credentials")
		return c.writeConfig(file, []byte(toWrite))
	}

	return nil
}

func git(args ...string) error {
	if _, err := exec.Command("git", args...).CombinedOutput(); err != nil {
		return errors.Wrapf(err, "running git command with args %s", strings.Join(args, ","))
	}

	return nil
}
