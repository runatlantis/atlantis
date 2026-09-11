// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package planstore

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	securejoin "github.com/cyphar/filepath-securejoin"
	"github.com/redis/go-redis/v9"

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/utils"
)

const redisOpTimeout = 30 * time.Second

// RedisPlanStore implements PlanStore by persisting every plan for a pull
// request as a field in one redis hash, so a single key carries one TTL and
// listing, loading, and deletion never scan the keyspace.
type RedisPlanStore struct {
	client redis.Cmdable
	ttl    time.Duration
	logger logging.SimpleLogging
}

// NewRedisPlanStore builds a store over an existing redis handle. ttl of 0
// disables expiry, leaving cleanup to Remove and DeleteForPull.
func NewRedisPlanStore(client redis.Cmdable, ttl time.Duration, logger logging.SimpleLogging) *RedisPlanStore {
	return &RedisPlanStore{
		client: client,
		ttl:    ttl,
		logger: logger,
	}
}

// pullKey names the per-pull hash. The {owner/repo/pull} hash tag pins it to
// one cluster slot, and every replica derives the same key for a PR.
func pullKey(owner, repo string, pullNum int) string {
	return fmt.Sprintf("plan:{%s/%s/%d}", owner, repo, pullNum)
}

// planField is the hash field for one plan: workspace/repoRelDir/planfile.
func planField(workspace, repoRelDir, planFile string) string {
	return strings.Join([]string{workspace, repoRelDir, planFile}, "/")
}

// encode prepends the head-commit and a newline to the plan bytes so both live
// in one field; decode splits them back.
func encode(headCommit string, data []byte) []byte {
	return append([]byte(headCommit+"\n"), data...)
}

func decode(raw []byte) (headCommit string, data []byte, ok bool) {
	i := bytes.IndexByte(raw, '\n')
	if i < 0 {
		return "", nil, false
	}
	return string(raw[:i]), raw[i+1:], true
}

func (s *RedisPlanStore) Save(ctx command.ProjectContext, planPath string) error {
	data, err := os.ReadFile(planPath)
	if err != nil {
		return fmt.Errorf("reading plan file for redis save: %w", err)
	}
	key := pullKey(ctx.BaseRepo.Owner, ctx.BaseRepo.Name, ctx.Pull.Num)
	field := planField(ctx.Workspace, ctx.RepoRelDir, filepath.Base(planPath))

	rctx, cancel := context.WithTimeout(context.Background(), redisOpTimeout)
	defer cancel()
	pipe := s.client.TxPipeline()
	pipe.HSet(rctx, key, field, encode(ctx.Pull.HeadCommit, data))
	if s.ttl > 0 {
		pipe.Expire(rctx, key, s.ttl)
	}
	if _, err := pipe.Exec(rctx); err != nil {
		return fmt.Errorf("saving plan to redis (key=%s field=%s): %w", key, field, err)
	}
	s.logger.Info("saved plan to redis %s[%s]", key, field)
	return nil
}

func (s *RedisPlanStore) Load(ctx command.ProjectContext, planPath string) error {
	key := pullKey(ctx.BaseRepo.Owner, ctx.BaseRepo.Name, ctx.Pull.Num)
	field := planField(ctx.Workspace, ctx.RepoRelDir, filepath.Base(planPath))

	rctx, cancel := context.WithTimeout(context.Background(), redisOpTimeout)
	defer cancel()
	raw, err := s.client.HGet(rctx, key, field).Bytes()
	if errors.Is(err, redis.Nil) {
		return fmt.Errorf("plan not found in redis (key=%s field=%s), run plan again", key, field)
	}
	if err != nil {
		return fmt.Errorf("loading plan from redis (key=%s field=%s): %w", key, field, err)
	}
	planCommit, data, ok := decode(raw)
	if !ok || planCommit == "" {
		return fmt.Errorf("plan in redis has no head-commit (key=%s field=%s), run plan again", key, field)
	}
	if ctx.Pull.HeadCommit != "" && planCommit != ctx.Pull.HeadCommit {
		return fmt.Errorf("plan was created at commit %.8s but PR is now at %.8s, run plan again", planCommit, ctx.Pull.HeadCommit)
	}
	if err := os.MkdirAll(filepath.Dir(planPath), 0o700); err != nil {
		return fmt.Errorf("creating parent directories for plan file: %w", err)
	}
	return os.WriteFile(planPath, data, 0o600)
}

func (s *RedisPlanStore) Remove(ctx command.ProjectContext, planPath string) error {
	key := pullKey(ctx.BaseRepo.Owner, ctx.BaseRepo.Name, ctx.Pull.Num)
	field := planField(ctx.Workspace, ctx.RepoRelDir, filepath.Base(planPath))

	rctx, cancel := context.WithTimeout(context.Background(), redisOpTimeout)
	defer cancel()
	if err := s.client.HDel(rctx, key, field).Err(); err != nil {
		s.logger.Warn("failed to delete plan from redis (key=%s field=%s): %v", key, field, err)
	}
	return utils.RemoveIgnoreNonExistent(planPath)
}

func (s *RedisPlanStore) fields(owner, repo string, pullNum int) ([]string, error) {
	rctx, cancel := context.WithTimeout(context.Background(), redisOpTimeout)
	defer cancel()
	fields, err := s.client.HKeys(rctx, pullKey(owner, repo, pullNum)).Result()
	if err != nil {
		return nil, fmt.Errorf("listing redis plan fields for %s/%s#%d: %w", owner, repo, pullNum, err)
	}
	return fields, nil
}

func (s *RedisPlanStore) ListWorkspaces(owner, repo string, pullNum int) ([]string, error) {
	fields, err := s.fields(owner, repo, pullNum)
	if err != nil {
		return nil, err
	}
	seen := map[string]struct{}{}
	for _, f := range fields {
		if ws, _, ok := strings.Cut(f, "/"); ok && ws != "" {
			seen[ws] = struct{}{}
		}
	}
	workspaces := make([]string, 0, len(seen))
	for ws := range seen {
		workspaces = append(workspaces, ws)
	}
	sort.Strings(workspaces)
	return workspaces, nil
}

func (s *RedisPlanStore) RestorePlans(pullDir, owner, repo string, pullNum int) error {
	if pullDir == "" {
		return nil // capability probe: redis store supports restore
	}
	key := pullKey(owner, repo, pullNum)
	fields, err := s.fields(owner, repo, pullNum)
	if err != nil {
		return err
	}

	var restored int
	for _, f := range fields {
		// SecureJoin keeps the write inside pullDir even if a field name is
		// tampered with in redis.
		localPath, err := securejoin.SecureJoin(pullDir, f)
		if err != nil {
			return fmt.Errorf("resolving safe path for redis field %s: %w", f, err)
		}
		// Fresh per-plan deadline so one slow HGET doesn't starve the rest.
		getCtx, getCancel := context.WithTimeout(context.Background(), redisOpTimeout)
		raw, err := s.client.HGet(getCtx, key, f).Bytes()
		getCancel()
		if errors.Is(err, redis.Nil) {
			continue // field deleted between HKeys and HGet
		}
		if err != nil {
			return fmt.Errorf("restoring plan from redis (field=%s): %w", f, err)
		}
		_, data, ok := decode(raw)
		if !ok {
			return fmt.Errorf("restoring plan from redis (field=%s): malformed value", f)
		}
		if err := os.MkdirAll(filepath.Dir(localPath), 0o700); err != nil {
			return fmt.Errorf("creating directory for restored plan: %w", err)
		}
		if err := os.WriteFile(localPath, data, 0o600); err != nil {
			return fmt.Errorf("writing restored plan file %s: %w", localPath, err)
		}
		restored++
	}
	s.logger.Info("restored %d plan(s) from redis for %s/%s#%d", restored, owner, repo, pullNum)
	return nil
}

func (s *RedisPlanStore) DeleteForPull(owner, repo string, pullNum int) error {
	rctx, cancel := context.WithTimeout(context.Background(), redisOpTimeout)
	defer cancel()
	if err := s.client.Del(rctx, pullKey(owner, repo, pullNum)).Err(); err != nil {
		s.logger.Warn("failed to delete plans from redis for %s/%s#%d: %v", owner, repo, pullNum, err)
	}
	return nil
}

func (s *RedisPlanStore) DeletePlanForProject(owner, repo string, pullNum int, workspace, repoRelDir, projectName string) error {
	var planFilename string
	if projectName == "" {
		planFilename = workspace + ".tfplan"
	} else {
		planFilename = strings.ReplaceAll(projectName, "/", "::") + "-" + workspace + ".tfplan"
	}
	field := planField(workspace, repoRelDir, planFilename)

	key := pullKey(owner, repo, pullNum)
	rctx, cancel := context.WithTimeout(context.Background(), redisOpTimeout)
	defer cancel()
	if err := s.client.HDel(rctx, key, field).Err(); err != nil {
		s.logger.Warn("failed to delete plan from redis (key=%s field=%s): %v", key, field, err)
	}
	return nil
}

var _ PlanStore = (*RedisPlanStore)(nil)
