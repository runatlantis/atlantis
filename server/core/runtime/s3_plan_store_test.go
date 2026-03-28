// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package runtime_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/runatlantis/atlantis/server/core/runtime"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockS3Client records calls and returns configured responses.
type mockS3Client struct {
	putInput    *s3.PutObjectInput
	putBody     []byte
	putErr      error
	getBody     []byte
	getMetadata map[string]string
	getErr      error
	deleteInput *s3.DeleteObjectInput
	deleteErr   error
	// deletedKeys tracks all keys passed to DeleteObject
	deletedKeys []string

	// For HeadBucket startup validation
	headBucketErr error

	// For ListObjectsV2 / RestorePlans testing
	listOutput *s3.ListObjectsV2Output
	listErr    error
	// getObjects maps S3 key to body content for multi-key GetObject calls
	getObjects map[string][]byte
}

func (m *mockS3Client) HeadBucket(_ context.Context, _ *s3.HeadBucketInput, _ ...func(*s3.Options)) (*s3.HeadBucketOutput, error) {
	return &s3.HeadBucketOutput{}, m.headBucketErr
}

func (m *mockS3Client) PutObject(_ context.Context, input *s3.PutObjectInput, _ ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	m.putInput = input
	if input.Body != nil {
		b, _ := io.ReadAll(input.Body)
		m.putBody = b
	}
	return &s3.PutObjectOutput{}, m.putErr
}

func (m *mockS3Client) GetObject(_ context.Context, input *s3.GetObjectInput, _ ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	// Support per-key bodies for RestorePlans testing.
	if m.getObjects != nil {
		key := *input.Key
		if body, ok := m.getObjects[key]; ok {
			return &s3.GetObjectOutput{
				Body: io.NopCloser(bytes.NewReader(body)),
			}, nil
		}
		return nil, errors.New("no such key: " + key)
	}
	return &s3.GetObjectOutput{
		Body:     io.NopCloser(bytes.NewReader(m.getBody)),
		Metadata: m.getMetadata,
	}, nil
}

func (m *mockS3Client) DeleteObject(_ context.Context, input *s3.DeleteObjectInput, _ ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	m.deleteInput = input
	m.deletedKeys = append(m.deletedKeys, aws.ToString(input.Key))
	return &s3.DeleteObjectOutput{}, m.deleteErr
}

func (m *mockS3Client) ListObjectsV2(_ context.Context, _ *s3.ListObjectsV2Input, _ ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	if m.listOutput != nil {
		return m.listOutput, nil
	}
	return &s3.ListObjectsV2Output{}, nil
}

func testProjectContext() command.ProjectContext {
	return command.ProjectContext{
		BaseRepo: models.Repo{
			Owner: "acme",
			Name:  "infra",
		},
		Pull: models.PullRequest{
			Num: 42,
		},
		Workspace:  "default",
		RepoRelDir: "modules/vpc",
	}
}

func TestS3Key_WithPrefix(t *testing.T) {
	store := runtime.NewS3PlanStoreWithClient(&mockS3Client{}, "bucket", "atlantis/plans", logging.NewNoopLogger(t))
	ctx := testProjectContext()

	key := store.TestS3Key(ctx, "/tmp/plans/myproject-default.tfplan")
	assert.Equal(t, "atlantis/plans/acme/infra/42/default/modules/vpc/myproject-default.tfplan", key)
}

func TestS3Key_WithoutPrefix(t *testing.T) {
	store := runtime.NewS3PlanStoreWithClient(&mockS3Client{}, "bucket", "", logging.NewNoopLogger(t))
	ctx := testProjectContext()

	key := store.TestS3Key(ctx, "/tmp/plans/myproject-default.tfplan")
	assert.Equal(t, "acme/infra/42/default/modules/vpc/myproject-default.tfplan", key)
}

func TestS3Key_NestedRepoRelDir(t *testing.T) {
	store := runtime.NewS3PlanStoreWithClient(&mockS3Client{}, "bucket", "pfx", logging.NewNoopLogger(t))
	ctx := testProjectContext()
	ctx.RepoRelDir = "envs/prod/us-east-1"

	key := store.TestS3Key(ctx, "/tmp/plan.tfplan")
	assert.Equal(t, "pfx/acme/infra/42/default/envs/prod/us-east-1/plan.tfplan", key)
}

func TestS3Key_TrailingSlashPrefix(t *testing.T) {
	store := runtime.NewS3PlanStoreWithClient(&mockS3Client{}, "bucket", "prefix/", logging.NewNoopLogger(t))
	ctx := testProjectContext()

	key := store.TestS3Key(ctx, "/tmp/plan.tfplan")
	assert.Equal(t, "prefix/acme/infra/42/default/modules/vpc/plan.tfplan", key)
}

func TestSave_Success(t *testing.T) {
	mock := &mockS3Client{}
	store := runtime.NewS3PlanStoreWithClient(mock, "my-bucket", "pfx", logging.NewNoopLogger(t))
	ctx := testProjectContext()
	ctx.Pull.HeadCommit = "abc123def456"

	planDir := t.TempDir()
	planPath := filepath.Join(planDir, "test.tfplan")
	require.NoError(t, os.WriteFile(planPath, []byte("plan-content"), 0o644))

	err := store.Save(ctx, planPath)
	require.NoError(t, err)

	assert.Equal(t, "my-bucket", *mock.putInput.Bucket)
	assert.Equal(t, "pfx/acme/infra/42/default/modules/vpc/test.tfplan", *mock.putInput.Key)
	assert.Equal(t, []byte("plan-content"), mock.putBody)
	assert.Equal(t, "abc123def456", mock.putInput.Metadata["head-commit"])
}

func TestSave_S3Error(t *testing.T) {
	mock := &mockS3Client{putErr: errors.New("access denied")}
	store := runtime.NewS3PlanStoreWithClient(mock, "bucket", "", logging.NewNoopLogger(t))
	ctx := testProjectContext()

	planDir := t.TempDir()
	planPath := filepath.Join(planDir, "test.tfplan")
	require.NoError(t, os.WriteFile(planPath, []byte("data"), 0o644))

	err := store.Save(ctx, planPath)
	assert.ErrorContains(t, err, "access denied")
}

func TestSave_FileOpenError(t *testing.T) {
	store := runtime.NewS3PlanStoreWithClient(&mockS3Client{}, "bucket", "", logging.NewNoopLogger(t))
	ctx := testProjectContext()

	err := store.Save(ctx, "/nonexistent/path/plan.tfplan")
	assert.ErrorContains(t, err, "opening plan file")
}

func TestLoad_Success(t *testing.T) {
	planContent := []byte("downloaded-plan-data")
	mock := &mockS3Client{
		getBody:     planContent,
		getMetadata: map[string]string{"Head-Commit": "abc123"},
	}
	store := runtime.NewS3PlanStoreWithClient(mock, "bucket", "pfx", logging.NewNoopLogger(t))
	ctx := testProjectContext()
	ctx.Pull.HeadCommit = "abc123"

	planDir := t.TempDir()
	planPath := filepath.Join(planDir, "subdir", "test.tfplan")

	err := store.Load(ctx, planPath)
	require.NoError(t, err)

	got, err := os.ReadFile(planPath)
	require.NoError(t, err)
	assert.Equal(t, planContent, got)
}

func TestLoad_StalePlanRejected(t *testing.T) {
	mock := &mockS3Client{
		getBody:     []byte("old-plan"),
		getMetadata: map[string]string{"Head-Commit": "oldcommit"},
	}
	store := runtime.NewS3PlanStoreWithClient(mock, "bucket", "", logging.NewNoopLogger(t))
	ctx := testProjectContext()
	ctx.Pull.HeadCommit = "newcommit"

	err := store.Load(ctx, filepath.Join(t.TempDir(), "plan.tfplan"))
	assert.ErrorContains(t, err, "plan was created at commit oldcommi but PR is now at newcommi")
}

func TestLoad_MissingMetadataRejected(t *testing.T) {
	mock := &mockS3Client{
		getBody:     []byte("plan-data"),
		getMetadata: map[string]string{},
	}
	store := runtime.NewS3PlanStoreWithClient(mock, "bucket", "", logging.NewNoopLogger(t))
	ctx := testProjectContext()
	ctx.Pull.HeadCommit = "abc123"

	err := store.Load(ctx, filepath.Join(t.TempDir(), "plan.tfplan"))
	assert.ErrorContains(t, err, "no head-commit metadata")
}

func TestLoad_S3Error(t *testing.T) {
	mock := &mockS3Client{getErr: errors.New("no such key")}
	store := runtime.NewS3PlanStoreWithClient(mock, "bucket", "", logging.NewNoopLogger(t))
	ctx := testProjectContext()

	err := store.Load(ctx, "/tmp/nonexistent/plan.tfplan")
	assert.ErrorContains(t, err, "no such key")
}

func TestRemove_Success(t *testing.T) {
	mock := &mockS3Client{}
	store := runtime.NewS3PlanStoreWithClient(mock, "my-bucket", "pfx", logging.NewNoopLogger(t))
	ctx := testProjectContext()

	planDir := t.TempDir()
	planPath := filepath.Join(planDir, "test.tfplan")
	require.NoError(t, os.WriteFile(planPath, []byte("data"), 0o644))

	err := store.Remove(ctx, planPath)
	require.NoError(t, err)

	assert.Equal(t, "my-bucket", *mock.deleteInput.Bucket)
	assert.Equal(t, "pfx/acme/infra/42/default/modules/vpc/test.tfplan", *mock.deleteInput.Key)

	_, statErr := os.Stat(planPath)
	assert.True(t, os.IsNotExist(statErr))
}

func TestRemove_S3Error(t *testing.T) {
	mock := &mockS3Client{deleteErr: errors.New("forbidden")}
	store := runtime.NewS3PlanStoreWithClient(mock, "bucket", "", logging.NewNoopLogger(t))
	ctx := testProjectContext()

	// S3 delete errors are logged but not returned (soft-fail).
	err := store.Remove(ctx, "/tmp/whatever.tfplan")
	assert.NoError(t, err)
}

func TestRemove_LocalFileAlreadyGone(t *testing.T) {
	mock := &mockS3Client{}
	store := runtime.NewS3PlanStoreWithClient(mock, "bucket", "", logging.NewNoopLogger(t))
	ctx := testProjectContext()

	err := store.Remove(ctx, "/tmp/nonexistent-plan-file.tfplan")
	require.NoError(t, err)
}

func TestRestorePlans_Success(t *testing.T) {
	mock := &mockS3Client{
		listOutput: &s3.ListObjectsV2Output{
			Contents: []s3types.Object{
				{Key: aws.String("pfx/acme/infra/42/default/modules/vpc/plan.tfplan")},
				{Key: aws.String("pfx/acme/infra/42/staging/modules/rds/plan.tfplan")},
				{Key: aws.String("pfx/acme/infra/42/default/some-other-file.txt")}, // not a .tfplan — skipped
			},
		},
		getObjects: map[string][]byte{
			"pfx/acme/infra/42/default/modules/vpc/plan.tfplan": []byte("plan-vpc"),
			"pfx/acme/infra/42/staging/modules/rds/plan.tfplan": []byte("plan-rds"),
		},
	}
	store := runtime.NewS3PlanStoreWithClient(mock, "bucket", "pfx", logging.NewNoopLogger(t))
	pullDir := t.TempDir()

	err := store.RestorePlans(pullDir, "acme", "infra", 42)
	require.NoError(t, err)

	// Verify files were written to the correct paths.
	got1, err := os.ReadFile(filepath.Join(pullDir, "default", "modules", "vpc", "plan.tfplan"))
	require.NoError(t, err)
	assert.Equal(t, []byte("plan-vpc"), got1)

	got2, err := os.ReadFile(filepath.Join(pullDir, "staging", "modules", "rds", "plan.tfplan"))
	require.NoError(t, err)
	assert.Equal(t, []byte("plan-rds"), got2)
}

func TestRestorePlans_NoPlansFound(t *testing.T) {
	mock := &mockS3Client{
		listOutput: &s3.ListObjectsV2Output{
			Contents: []s3types.Object{},
		},
	}
	store := runtime.NewS3PlanStoreWithClient(mock, "bucket", "pfx", logging.NewNoopLogger(t))

	err := store.RestorePlans(t.TempDir(), "acme", "infra", 42)
	require.NoError(t, err)
}

func TestRestorePlans_ListError(t *testing.T) {
	mock := &mockS3Client{listErr: errors.New("access denied")}
	store := runtime.NewS3PlanStoreWithClient(mock, "bucket", "pfx", logging.NewNoopLogger(t))

	err := store.RestorePlans(t.TempDir(), "acme", "infra", 42)
	assert.ErrorContains(t, err, "access denied")
}

func TestRestorePlans_WithoutPrefix(t *testing.T) {
	mock := &mockS3Client{
		listOutput: &s3.ListObjectsV2Output{
			Contents: []s3types.Object{
				{Key: aws.String("acme/infra/42/default/plan.tfplan")},
			},
		},
		getObjects: map[string][]byte{
			"acme/infra/42/default/plan.tfplan": []byte("plan-data"),
		},
	}
	store := runtime.NewS3PlanStoreWithClient(mock, "bucket", "", logging.NewNoopLogger(t))
	pullDir := t.TempDir()

	err := store.RestorePlans(pullDir, "acme", "infra", 42)
	require.NoError(t, err)

	got, err := os.ReadFile(filepath.Join(pullDir, "default", "plan.tfplan"))
	require.NoError(t, err)
	assert.Equal(t, []byte("plan-data"), got)
}

func TestDeleteForPull_Success(t *testing.T) {
	mock := &mockS3Client{
		listOutput: &s3.ListObjectsV2Output{
			Contents: []s3types.Object{
				{Key: aws.String("pfx/acme/infra/42/default/modules/vpc/plan.tfplan")},
				{Key: aws.String("pfx/acme/infra/42/staging/modules/rds/plan.tfplan")},
			},
		},
	}
	store := runtime.NewS3PlanStoreWithClient(mock, "bucket", "pfx", logging.NewNoopLogger(t))

	err := store.DeleteForPull("acme", "infra", 42)
	require.NoError(t, err)

	assert.Equal(t, []string{
		"pfx/acme/infra/42/default/modules/vpc/plan.tfplan",
		"pfx/acme/infra/42/staging/modules/rds/plan.tfplan",
	}, mock.deletedKeys)
}

func TestDeleteForPull_NoObjects(t *testing.T) {
	mock := &mockS3Client{
		listOutput: &s3.ListObjectsV2Output{
			Contents: []s3types.Object{},
		},
	}
	store := runtime.NewS3PlanStoreWithClient(mock, "bucket", "pfx", logging.NewNoopLogger(t))

	err := store.DeleteForPull("acme", "infra", 42)
	require.NoError(t, err)
	assert.Empty(t, mock.deletedKeys)
}

func TestDeleteForPull_ListError(t *testing.T) {
	mock := &mockS3Client{listErr: errors.New("access denied")}
	store := runtime.NewS3PlanStoreWithClient(mock, "bucket", "pfx", logging.NewNoopLogger(t))

	err := store.DeleteForPull("acme", "infra", 42)
	assert.ErrorContains(t, err, "access denied")
}

func TestDeleteForPull_DeleteError(t *testing.T) {
	mock := &mockS3Client{
		listOutput: &s3.ListObjectsV2Output{
			Contents: []s3types.Object{
				{Key: aws.String("pfx/acme/infra/42/default/plan.tfplan")},
			},
		},
		deleteErr: errors.New("forbidden"),
	}
	store := runtime.NewS3PlanStoreWithClient(mock, "bucket", "pfx", logging.NewNoopLogger(t))

	// S3 delete errors during cleanup are logged but not returned (soft-fail).
	err := store.DeleteForPull("acme", "infra", 42)
	assert.NoError(t, err)
}
