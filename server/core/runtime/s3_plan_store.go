// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/utils"
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
	key := s.s3Key(ctx, planPath)

	f, err := os.Open(planPath)
	if err != nil {
		return fmt.Errorf("opening plan file for S3 upload: %w", err)
	}
	defer f.Close()

	metadata := map[string]string{}
	if ctx.Pull.HeadCommit != "" {
		metadata["head-commit"] = ctx.Pull.HeadCommit
	}
	if ctx.User.Username != "" {
		metadata["planned-by"] = ctx.User.Username
	}

	_, err = s.client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket:   aws.String(s.bucket),
		Key:      aws.String(key),
		Body:     f,
		Metadata: metadata,
	})
	if err != nil {
		return fmt.Errorf("uploading plan to S3 (key=%s): %w", key, err)
	}

	s.logger.Info("uploaded plan to s3://%s/%s", s.bucket, key)
	return nil
}

// Load downloads the plan file from S3 and writes it to planPath.
func (s *S3PlanStore) Load(ctx command.ProjectContext, planPath string) error {
	key := s.s3Key(ctx, planPath)

	resp, err := s.client.GetObject(context.Background(), &s3.GetObjectInput{
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

// Remove deletes the plan file from S3 and locally.
func (s *S3PlanStore) Remove(ctx command.ProjectContext, planPath string) error {
	key := s.s3Key(ctx, planPath)

	if _, err := s.client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}); err != nil {
		s.logger.Warn("failed to delete plan from S3 (key=%s): %v", key, err)
	} else {
		s.logger.Debug("deleted plan from s3://%s/%s", s.bucket, key)
	}

	return utils.RemoveIgnoreNonExistent(planPath)
}

// RestorePlans lists all plan files for a pull request in S3 (via prefix scan)
// and downloads them into pullDir so PendingPlanFinder can discover them.
// Only called from the "apply all" path where we don't know which projects
// were planned. The single-project path skips this and uses Load directly.
//
// Note: plans downloaded here will be re-downloaded by Load() in
// ApplyStepRunner, which also validates head-commit metadata. This means
// each plan is fetched from S3 twice in the "apply all" path. Acceptable
// since plan files are small; eliminating it would require shared state
// between RestorePlans and Load.
func (s *S3PlanStore) RestorePlans(pullDir, owner, repo string, pullNum int) error {
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
		resp, err := s.client.ListObjectsV2(context.Background(), &s3.ListObjectsV2Input{
			Bucket:            aws.String(s.bucket),
			Prefix:            aws.String(listPrefix),
			ContinuationToken: continuationToken,
		})
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
			relPath = filepath.Clean(relPath)
			if relPath == "." || relPath == string(os.PathSeparator) {
				s.logger.Info("skipping S3 object with empty relative path (key=%s, prefix=%s)", key, listPrefix)
				continue
			}
			if filepath.IsAbs(relPath) || relPath == ".." || strings.HasPrefix(relPath, ".."+string(os.PathSeparator)) {
				return fmt.Errorf("refusing to restore plan outside pull dir (key=%s, relPath=%s)", key, relPath)
			}
			localPath := filepath.Join(pullDir, relPath)

			if err := os.MkdirAll(filepath.Dir(localPath), 0o700); err != nil {
				return fmt.Errorf("creating directory for restored plan: %w", err)
			}

			getResp, err := s.client.GetObject(context.Background(), &s3.GetObjectInput{
				Bucket: aws.String(s.bucket),
				Key:    aws.String(key),
			})
			if err != nil {
				return fmt.Errorf("downloading plan from S3 (key=%s): %w", key, err)
			}

			f, err := os.Create(localPath)
			if err != nil {
				getResp.Body.Close()
				return fmt.Errorf("creating local plan file %s: %w", localPath, err)
			}

			_, copyErr := io.Copy(f, getResp.Body)
			f.Close()
			getResp.Body.Close()
			if copyErr != nil {
				return fmt.Errorf("writing restored plan file %s: %w", localPath, copyErr)
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
		resp, err := s.client.ListObjectsV2(context.Background(), &s3.ListObjectsV2Input{
			Bucket:            aws.String(s.bucket),
			Prefix:            aws.String(listPrefix),
			ContinuationToken: continuationToken,
		})
		if err != nil {
			return fmt.Errorf("listing plans for deletion (prefix=%s): %w", listPrefix, err)
		}

		for _, obj := range resp.Contents {
			key := aws.ToString(obj.Key)
			if _, err := s.client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
				Bucket: aws.String(s.bucket),
				Key:    aws.String(key),
			}); err != nil {
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

// TestS3Key is exported for testing only.
func (s *S3PlanStore) TestS3Key(ctx command.ProjectContext, planPath string) string {
	return s.s3Key(ctx, planPath)
}

// Ensure S3PlanStore satisfies PlanStore at compile time.
var _ PlanStore = (*S3PlanStore)(nil)

// Ensure the real S3 client satisfies our interface at compile time.
var _ S3Client = (*s3.Client)(nil)

