// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package planstore

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	securejoin "github.com/cyphar/filepath-securejoin"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/logging"
)

// S3Client is the subset of the S3 API used by S3PlanStore, extracted for testability.
type S3Client interface {
	HeadBucket(ctx context.Context, params *s3.HeadBucketInput, optFns ...func(*s3.Options)) (*s3.HeadBucketOutput, error)
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
	ListObjectsV2(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
}

// S3PlanStoreConfig holds configuration for connecting to S3.
type S3PlanStoreConfig struct {
	Bucket         string
	Region         string
	Prefix         string
	Endpoint       string
	ForcePathStyle bool
	Profile        string
}

// S3PlanStore implements PlanStore by persisting plan files to S3.
type S3PlanStore struct {
	client S3Client
	bucket string
	prefix string
	logger logging.SimpleLogging
}

// NewS3PlanStore creates an S3PlanStore using the AWS SDK default credential chain.
func NewS3PlanStore(cfg S3PlanStoreConfig, logger logging.SimpleLogging) (*S3PlanStore, error) {
	var opts []func(*awsconfig.LoadOptions) error
	opts = append(opts, awsconfig.WithRegion(cfg.Region))

	if cfg.Profile != "" {
		opts = append(opts, awsconfig.WithSharedConfigProfile(cfg.Profile))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("loading AWS config: %w", err)
	}

	var s3Opts []func(*s3.Options)
	if cfg.Endpoint != "" {
		s3Opts = append(s3Opts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
			if cfg.ForcePathStyle {
				o.UsePathStyle = true
			}
		})
	} else if cfg.ForcePathStyle {
		s3Opts = append(s3Opts, func(o *s3.Options) {
			o.UsePathStyle = true
		})
	}

	client := s3.NewFromConfig(awsCfg, s3Opts...)

	if _, err := client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(cfg.Bucket),
	}); err != nil {
		return nil, fmt.Errorf("validating S3 plan store bucket %q: %w", cfg.Bucket, err)
	}

	return NewS3PlanStoreWithClient(client, cfg.Bucket, cfg.Prefix, logger), nil
}

// s3OpTimeout is the per-operation timeout for S3 API calls.
const s3OpTimeout = 30 * time.Second

// managedPlanVersionMarker separates immutable managed artifacts from the
// deterministic .tfplan object retained for legacy restore and discovery.
// Version keys intentionally do not end in .tfplan so RestorePlans ignores
// them until it has an accepted durable generation to select.
const managedPlanVersionMarker = ".atlantis-managed/"

// s3Ctx returns a context with the standard S3 operation timeout.
func s3Ctx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), s3OpTimeout)
}

// NewS3PlanStoreWithClient creates an S3PlanStore with an injected S3Client (for testing).
func NewS3PlanStoreWithClient(client S3Client, bucket, prefix string, logger logging.SimpleLogging) *S3PlanStore {
	return &S3PlanStore{
		client: client,
		bucket: bucket,
		prefix: strings.TrimSuffix(prefix, "/"),
		logger: logger,
	}
}

// Save uploads the plan file at planPath to S3.
func (s *S3PlanStore) Save(ctx command.ProjectContext, planPath string) error {
	metadata := map[string]string{}
	if ctx.Pull.HeadCommit != "" {
		metadata["head-commit"] = ctx.Pull.HeadCommit
	}
	if ctx.User.Username != "" {
		metadata["planned-by"] = ctx.User.Username
	}
	if ctx.GeneratedPlanHash != nil && *ctx.GeneratedPlanHash != "" {
		metadata["managed-plan-sha256"] = *ctx.GeneratedPlanHash
	}
	if ctx.PlanGeneration != "" {
		metadata["plan-generation"] = ctx.PlanGeneration
	}

	canonicalKey := s.s3Key(ctx, planPath)
	if !ctx.RequiresAtlantisManagedPlanFile {
		f, err := os.Open(planPath)
		if err != nil {
			return fmt.Errorf("opening plan file for S3 upload: %w", err)
		}
		defer f.Close()
		if err := s.uploadPlan(canonicalKey, f, metadata); err != nil {
			return err
		}
		s.logger.Info("uploaded plan to s3://%s/%s", s.bucket, canonicalKey)
		return nil
	}

	if ctx.PlanGeneration == "" && !isSyntheticNonPRAPI(ctx) {
		return fmt.Errorf("saving managed plan to S3: plan generation is empty")
	}
	if ctx.GeneratedPlanHash == nil || *ctx.GeneratedPlanHash == "" {
		return fmt.Errorf("saving managed plan to S3: managed plan hash is empty")
	}

	snapshot, cleanup, err := snapshotManagedPlan(planPath, *ctx.GeneratedPlanHash)
	if err != nil {
		return err
	}
	defer cleanup()

	immutableKey := ""
	if ctx.PlanGeneration != "" {
		immutableKey = s.managedPlanKey(ctx, planPath, ctx.PlanGeneration, *ctx.GeneratedPlanHash)
		if err := s.uploadPlan(immutableKey, snapshot, metadata); err != nil {
			return err
		}
		if _, err := snapshot.Seek(0, io.SeekStart); err != nil {
			return fmt.Errorf("rewinding managed plan snapshot for canonical S3 upload: %w", err)
		}
	}
	if err := s.uploadPlan(canonicalKey, snapshot, metadata); err != nil {
		return err
	}

	if immutableKey != "" {
		s.logger.Info("uploaded immutable managed plan to s3://%s/%s and canonical discovery plan to s3://%s/%s", s.bucket, immutableKey, s.bucket, canonicalKey)
	} else {
		s.logger.Info("uploaded synthetic non-PR managed plan to s3://%s/%s", s.bucket, canonicalKey)
	}
	return nil
}

func (s *S3PlanStore) uploadPlan(key string, body io.Reader, metadata map[string]string) error {
	opCtx, opCancel := s3Ctx()
	defer opCancel()
	if _, err := s.client.PutObject(opCtx, &s3.PutObjectInput{
		Bucket:   aws.String(s.bucket),
		Key:      aws.String(key),
		Body:     body,
		Metadata: metadata,
	}); err != nil {
		return fmt.Errorf("uploading plan to S3 (key=%s): %w", key, err)
	}
	return nil
}

func snapshotManagedPlan(planPath, expectedHash string) (*os.File, func(), error) {
	source, err := os.Open(planPath)
	if err != nil {
		return nil, nil, fmt.Errorf("opening plan file for S3 upload: %w", err)
	}
	defer source.Close()

	snapshot, err := os.CreateTemp("", "atlantis-managed-plan-upload-*")
	if err != nil {
		return nil, nil, fmt.Errorf("creating managed plan snapshot for S3 upload: %w", err)
	}
	cleanup := func() {
		_ = snapshot.Close()
		_ = os.Remove(snapshot.Name())
	}
	hasher := sha256.New()
	if _, err := io.Copy(io.MultiWriter(snapshot, hasher), source); err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("snapshotting managed plan for S3 upload: %w", err)
	}
	actualHash := hex.EncodeToString(hasher.Sum(nil))
	if actualHash != expectedHash {
		cleanup()
		return nil, nil, fmt.Errorf("managed plan hash changed before S3 upload: expected %s, got %s", expectedHash, actualHash)
	}
	if _, err := snapshot.Seek(0, io.SeekStart); err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("rewinding managed plan snapshot for S3 upload: %w", err)
	}
	return snapshot, cleanup, nil
}

// Load downloads the plan file from S3 and writes it to planPath.
func (s *S3PlanStore) Load(ctx command.ProjectContext, planPath string) error {
	key := s.s3Key(ctx, planPath)
	partialManagedIdentity := (ctx.RecordedManagedPlanHash == "") != (ctx.AcceptedPlanGeneration == "")
	if partialManagedIdentity && !isSyntheticNonPRAPI(ctx) {
		return fmt.Errorf("loading managed plan from S3: durable plan hash and accepted generation must both be set")
	}
	immutableManagedPlan := !partialManagedIdentity && ctx.RecordedManagedPlanHash != "" && ctx.AcceptedPlanGeneration != ""
	if immutableManagedPlan {
		key = s.managedPlanKey(ctx, planPath, ctx.AcceptedPlanGeneration, ctx.RecordedManagedPlanHash)
	}

	opCtx, opCancel := s3Ctx()
	defer opCancel()
	resp, err := s.client.GetObject(opCtx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("downloading plan from S3 (key=%s): %w", key, err)
	}
	defer resp.Body.Close()

	// Reject stale plans: the plan must have been created at the same commit
	// the PR currently points to. This prevents applying outdated plans after
	// new commits are pushed (e.g. across container restarts).
	// Note: different S3/S3-compatible implementations may return user-defined
	// metadata keys with different casing, so we look up "head-commit"
	// case-insensitively.
	var planCommit string
	for k, v := range resp.Metadata {
		if strings.EqualFold(k, "head-commit") {
			planCommit = v
			break
		}
	}
	if planCommit == "" {
		return fmt.Errorf("plan in S3 has no head-commit metadata (key=%s) — run plan again", key)
	}
	if ctx.Pull.HeadCommit != "" && planCommit != ctx.Pull.HeadCommit {
		return fmt.Errorf("plan was created at commit %.8s but PR is now at %.8s — run plan again", planCommit, ctx.Pull.HeadCommit)
	}
	if immutableManagedPlan {
		storedHash := metadataValue(resp.Metadata, "managed-plan-sha256")
		if storedHash != ctx.RecordedManagedPlanHash {
			return fmt.Errorf("managed plan in S3 has hash metadata %q, expected %q (key=%s) — run plan again", storedHash, ctx.RecordedManagedPlanHash, key)
		}
		storedGeneration := metadataValue(resp.Metadata, "plan-generation")
		if storedGeneration != ctx.AcceptedPlanGeneration {
			return fmt.Errorf("managed plan in S3 has generation metadata %q, expected %q (key=%s) — run plan again", storedGeneration, ctx.AcceptedPlanGeneration, key)
		}
	}

	if err := os.MkdirAll(filepath.Dir(planPath), 0o700); err != nil {
		return fmt.Errorf("creating parent directories for plan file: %w", err)
	}

	f, err := os.Create(planPath)
	if err != nil {
		return fmt.Errorf("creating local plan file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		return fmt.Errorf("writing plan file from S3: %w", err)
	}

	s.logger.Debug("downloaded plan from s3://%s/%s", s.bucket, key)
	return nil
}

func isSyntheticNonPRAPI(ctx command.ProjectContext) bool {
	return ctx.API && ctx.Pull.Num <= 0
}

func metadataValue(metadata map[string]string, name string) string {
	for key, value := range metadata {
		if strings.EqualFold(key, name) {
			return value
		}
	}
	return ""
}

// Remove deletes legacy/hashless artifacts and uniquely scoped synthetic
// non-PR API artifacts. PR-backed digest-bound managed plans are retained until
// explicit pull/project cleanup because an apply cannot prove artifact ownership.
func (s *S3PlanStore) Remove(ctx command.ProjectContext, planPath string) error {
	key := s.s3Key(ctx, planPath)
	if ctx.RecordedManagedPlanHash == "" || isSyntheticNonPRAPI(ctx) {
		opCtx, opCancel := s3Ctx()
		defer opCancel()
		if _, err := s.client.DeleteObject(opCtx, &s3.DeleteObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(key)}); err != nil {
			s.logger.Warn("failed to delete plan from S3 (key=%s): %v", key, err)
		} else {
			s.logger.Debug("deleted plan from s3://%s/%s", s.bucket, key)
		}
	} else {
		s.logger.Debug("retaining digest-bound managed plan in s3://%s/%s until explicit cleanup", s.bucket, key)
	}

	return (&LocalPlanStore{}).Remove(ctx, planPath)
}

// RestorePlans lists all plan files for a pull request in S3 (via prefix scan)
// and downloads them into pullDir so PendingPlanFinder can discover them.
// Only called from the "apply all" path where we don't know which projects
// were planned. The single-project path skips this and uses Load directly.
//
// Restored plans are validated by the apply path before execution. The runtime
// apply step consumes the already validated local bytes and does not load the
// deterministic S3 key again.
// ListWorkspaces scans the pull request's prefix in S3 and returns the unique
// workspace names (first path segment after owner/repo/pullNum/) that have at
// least one .tfplan stored. Callers use this to clone each workspace before
// invoking RestorePlans, so plan files don't get wiped by a subsequent Clone.
func (s *S3PlanStore) ListWorkspaces(owner, repo string, pullNum int) ([]string, error) {
	prefixParts := []string{}
	if s.prefix != "" {
		prefixParts = append(prefixParts, s.prefix)
	}
	prefixParts = append(prefixParts, owner, repo, strconv.Itoa(pullNum))
	listPrefix := strings.Join(prefixParts, "/") + "/"

	seen := map[string]struct{}{}
	var continuationToken *string
	for {
		opCtx, opCancel := s3Ctx()
		resp, err := s.client.ListObjectsV2(opCtx, &s3.ListObjectsV2Input{
			Bucket:            aws.String(s.bucket),
			Prefix:            aws.String(listPrefix),
			ContinuationToken: continuationToken,
		})
		opCancel()
		if err != nil {
			return nil, fmt.Errorf("listing workspaces from S3 (prefix=%s): %w", listPrefix, err)
		}
		for _, obj := range resp.Contents {
			key := aws.ToString(obj.Key)
			if !strings.HasSuffix(key, ".tfplan") {
				continue
			}
			rel := strings.TrimPrefix(key, listPrefix)
			workspace, _, ok := strings.Cut(rel, "/")
			if !ok || workspace == "" {
				continue
			}
			seen[workspace] = struct{}{}
		}
		if !aws.ToBool(resp.IsTruncated) {
			break
		}
		continuationToken = resp.NextContinuationToken
	}

	workspaces := make([]string, 0, len(seen))
	for w := range seen {
		workspaces = append(workspaces, w)
	}
	sort.Strings(workspaces)
	return workspaces, nil
}

func (s *S3PlanStore) RestorePlans(pullDir, owner, repo string, pullNum int) error {
	if pullDir == "" {
		return nil // capability probe: external store supports restore
	}
	// Build the S3 prefix for all plans under this pull request.
	prefixParts := []string{}
	if s.prefix != "" {
		prefixParts = append(prefixParts, s.prefix)
	}
	prefixParts = append(prefixParts, owner, repo, strconv.Itoa(pullNum))
	listPrefix := strings.Join(prefixParts, "/") + "/"

	var restored int
	var continuationToken *string
	for {
		listCtx, listCancel := s3Ctx()
		resp, err := s.client.ListObjectsV2(listCtx, &s3.ListObjectsV2Input{
			Bucket:            aws.String(s.bucket),
			Prefix:            aws.String(listPrefix),
			ContinuationToken: continuationToken,
		})
		listCancel()
		if err != nil {
			return fmt.Errorf("listing plans from S3 (prefix=%s): %w", listPrefix, err)
		}

		for _, obj := range resp.Contents {
			key := aws.ToString(obj.Key)
			if !strings.HasSuffix(key, ".tfplan") {
				continue
			}

			// Strip the prefix up to and including <pullNum>/ to get the relative path.
			relPath := strings.TrimPrefix(key, listPrefix)

			// SecureJoin guarantees the result stays within pullDir,
			// preventing path traversal from untrusted S3 keys.
			localPath, err := securejoin.SecureJoin(pullDir, relPath)
			if err != nil {
				return fmt.Errorf("resolving safe path for S3 key %s: %w", key, err)
			}

			if err := os.MkdirAll(filepath.Dir(localPath), 0o700); err != nil {
				return fmt.Errorf("creating directory for restored plan: %w", err)
			}

			if err := s.downloadObjectTo(key, localPath); err != nil {
				return err
			}

			restored++
			s.logger.Info("restored plan from s3://%s/%s to %s", s.bucket, key, localPath)
		}

		if !aws.ToBool(resp.IsTruncated) {
			break
		}
		continuationToken = resp.NextContinuationToken
	}

	s.logger.Info("restored %d plan(s) from S3 for %s/%s#%d", restored, owner, repo, pullNum)
	return nil
}

// downloadObjectTo fetches the S3 object at key and writes it to localPath.
// Each call uses its own bounded context so a slow object doesn't starve the
// rest of a paginated restore.
func (s *S3PlanStore) downloadObjectTo(key, localPath string) error {
	getCtx, getCancel := s3Ctx()
	defer getCancel()
	getResp, err := s.client.GetObject(getCtx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("downloading plan from S3 (key=%s): %w", key, err)
	}
	defer getResp.Body.Close()

	f, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("creating local plan file %s: %w", localPath, err)
	}
	defer f.Close()

	if _, err := io.Copy(f, getResp.Body); err != nil {
		return fmt.Errorf("writing restored plan file %s: %w", localPath, err)
	}
	return nil
}

// DeleteForPull removes all plan objects stored under the pull request prefix in S3.
func (s *S3PlanStore) DeleteForPull(owner, repo string, pullNum int) error {
	prefixParts := []string{}
	if s.prefix != "" {
		prefixParts = append(prefixParts, s.prefix)
	}
	prefixParts = append(prefixParts, owner, repo, strconv.Itoa(pullNum))
	listPrefix := strings.Join(prefixParts, "/") + "/"

	var deleted int
	var continuationToken *string
	for {
		listCtx, listCancel := s3Ctx()
		resp, err := s.client.ListObjectsV2(listCtx, &s3.ListObjectsV2Input{
			Bucket:            aws.String(s.bucket),
			Prefix:            aws.String(listPrefix),
			ContinuationToken: continuationToken,
		})
		listCancel()
		if err != nil {
			return fmt.Errorf("listing plans for deletion (prefix=%s): %w", listPrefix, err)
		}

		for _, obj := range resp.Contents {
			key := aws.ToString(obj.Key)
			delCtx, delCancel := s3Ctx()
			_, err := s.client.DeleteObject(delCtx, &s3.DeleteObjectInput{
				Bucket: aws.String(s.bucket),
				Key:    aws.String(key),
			})
			delCancel()
			if err != nil {
				s.logger.Warn("failed to delete plan from S3 (key=%s): %v", key, err)
				continue
			}
			deleted++
		}

		if !aws.ToBool(resp.IsTruncated) {
			break
		}
		continuationToken = resp.NextContinuationToken
	}

	if deleted > 0 {
		s.logger.Info("deleted %d plan(s) from S3 for %s/%s#%d", deleted, owner, repo, pullNum)
	}
	return nil
}

// s3Key builds a deterministic S3 object key from the ProjectContext and plan filename.
// Format: <prefix>/<owner>/<repo>/<pullNum>/<workspace>/<repoRelDir>/<planfilename>
func (s *S3PlanStore) s3Key(ctx command.ProjectContext, planPath string) string {
	parts := []string{}
	if s.prefix != "" {
		parts = append(parts, s.prefix)
	}
	parts = append(parts,
		ctx.BaseRepo.Owner,
		ctx.BaseRepo.Name,
		strconv.Itoa(ctx.Pull.Num),
		ctx.Workspace,
		ctx.RepoRelDir,
		filepath.Base(planPath),
	)
	return strings.Join(parts, "/")
}

func (s *S3PlanStore) managedPlanKey(ctx command.ProjectContext, planPath, generation, planHash string) string {
	return s.s3Key(ctx, planPath) + managedPlanVersionMarker + url.PathEscape(generation) + "/" + url.PathEscape(planHash)
}

// TestS3Key is exported for testing only.
func (s *S3PlanStore) TestS3Key(ctx command.ProjectContext, planPath string) string {
	return s.s3Key(ctx, planPath)
}

func (s *S3PlanStore) DeletePlanForProject(owner, repo string, pullNum int, workspace, repoRelDir, projectName string) error {
	var planFilename string
	if projectName == "" {
		planFilename = workspace + ".tfplan"
	} else {
		planFilename = strings.ReplaceAll(projectName, "/", "::") + "-" + workspace + ".tfplan"
	}
	parts := []string{}
	if s.prefix != "" {
		parts = append(parts, s.prefix)
	}
	parts = append(parts, owner, repo, strconv.Itoa(pullNum), workspace, repoRelDir, planFilename)
	key := strings.Join(parts, "/")
	versionPrefix := key + managedPlanVersionMarker

	var continuationToken *string
	for {
		listCtx, listCancel := s3Ctx()
		resp, err := s.client.ListObjectsV2(listCtx, &s3.ListObjectsV2Input{
			Bucket:            aws.String(s.bucket),
			Prefix:            aws.String(versionPrefix),
			ContinuationToken: continuationToken,
		})
		listCancel()
		if err != nil {
			s.logger.Warn("failed to list immutable managed plans for project cleanup (prefix=%s): %v", versionPrefix, err)
			break
		}
		for _, obj := range resp.Contents {
			versionKey := aws.ToString(obj.Key)
			deleteCtx, deleteCancel := s3Ctx()
			_, deleteErr := s.client.DeleteObject(deleteCtx, &s3.DeleteObjectInput{
				Bucket: aws.String(s.bucket),
				Key:    aws.String(versionKey),
			})
			deleteCancel()
			if deleteErr != nil {
				s.logger.Warn("failed to delete immutable managed plan from S3 (key=%s): %v", versionKey, deleteErr)
			}
		}
		if !aws.ToBool(resp.IsTruncated) {
			break
		}
		continuationToken = resp.NextContinuationToken
	}

	opCtx, opCancel := s3Ctx()
	defer opCancel()
	if _, err := s.client.DeleteObject(opCtx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}); err != nil {
		s.logger.Warn("failed to delete plan from S3 (key=%s): %v", key, err)
	} else {
		s.logger.Debug("deleted plan from s3://%s/%s", s.bucket, key)
	}
	return nil
}

// Ensure S3PlanStore satisfies PlanStore at compile time.
var _ PlanStore = (*S3PlanStore)(nil)

// Ensure the real S3 client satisfies our interface at compile time.
var _ S3Client = (*s3.Client)(nil)
