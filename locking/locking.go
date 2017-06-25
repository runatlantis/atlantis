package locking

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/hootsuite/atlantis/models"
)

type Backend interface {
	TryLock(lock models.ProjectLock) (bool, models.ProjectLock, error)
	Unlock(project models.Project, env string) (*models.ProjectLock, error)
	List() ([]models.ProjectLock, error)
	GetLock(project models.Project, env string) (models.ProjectLock, error)
	UnlockByPull(repoFullName string, pullNum int) ([]models.ProjectLock, error)
}

type TryLockResponse struct {
	LockAcquired bool
	CurrLock     models.ProjectLock
	LockKey      string
}

type Client struct {
	backend Backend
}

func NewClient(backend Backend) *Client {
	return &Client{
		backend: backend,
	}
}

// keyRegex matches and captures {repoFullName}/{path}/{env} where path can have multiple /'s in it
var keyRegex = regexp.MustCompile(`^(.*?\/.*?)\/(.*)\/(.*)$`)

func (c *Client) TryLock(p models.Project, env string, pull models.PullRequest, user models.User) (TryLockResponse, error) {
	lock := models.ProjectLock{
		Env:     env,
		Time:    time.Now(),
		Project: p,
		User:    user,
		Pull:    pull,
	}
	lockAcquired, currLock, err := c.backend.TryLock(lock)
	if err != nil {
		return TryLockResponse{}, err
	}
	return TryLockResponse{lockAcquired, currLock, c.key(p, env)}, nil
}

func (c *Client) Unlock(key string) (*models.ProjectLock, error) {
	project, env, err := c.lockKeyToProjectEnv(key)
	if err != nil {
		return nil, err
	}

	return c.backend.Unlock(project, env)
}

func (c *Client) List() (map[string]models.ProjectLock, error) {
	m := make(map[string]models.ProjectLock)
	locks, err := c.backend.List()
	if err != nil {
		return m, err
	}
	for _, lock := range locks {
		m[c.key(lock.Project, lock.Env)] = lock
	}
	return m, nil
}

func (c *Client) UnlockByPull(repoFullName string, pullNum int) ([]models.ProjectLock, error) {
	return c.backend.UnlockByPull(repoFullName, pullNum)
}

func (c *Client) GetLock(key string) (models.ProjectLock, error) {
	project, env, err := c.lockKeyToProjectEnv(key)
	if err != nil {
		return models.ProjectLock{}, err
	}

	projectLock, err := c.backend.GetLock(project, env)
	if err != nil {
		return models.ProjectLock{}, err
	}

	return projectLock, nil
}

func (c *Client) key(p models.Project, env string) string {
	return fmt.Sprintf("%s/%s/%s", p.RepoFullName, p.Path, env)
}

func (c *Client) lockKeyToProjectEnv(key string) (models.Project, string, error) {
	matches := keyRegex.FindStringSubmatch(key)
	if len(matches) != 4 {
		return models.Project{}, "", errors.New("invalid key format")
	}

	return models.Project{matches[1], matches[2]}, matches[3], nil
}
