// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package planstore_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/runatlantis/atlantis/server/core/planstore"
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

type memoryS3Object struct {
	body     []byte
	metadata map[string]string
}

type memoryS3Client struct {
	mu      sync.Mutex
	objects map[string]memoryS3Object
	deleted []string
}

type blockingGenerationS3Client struct {
	*memoryS3Client
	generation string
	entered    chan struct{}
	release    chan struct{}
	once       sync.Once
}

func (c *blockingGenerationS3Client) PutObject(ctx context.Context, input *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	if input.Metadata["plan-generation"] == c.generation {
		c.once.Do(func() {
			close(c.entered)
			<-c.release
		})
	}
	return c.memoryS3Client.PutObject(ctx, input, optFns...)
}

func newMemoryS3Client() *memoryS3Client {
	return &memoryS3Client{objects: make(map[string]memoryS3Object)}
}

func (m *memoryS3Client) HeadBucket(context.Context, *s3.HeadBucketInput, ...func(*s3.Options)) (*s3.HeadBucketOutput, error) {
	return &s3.HeadBucketOutput{}, nil
}

func (m *memoryS3Client) PutObject(_ context.Context, input *s3.PutObjectInput, _ ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	body, err := io.ReadAll(input.Body)
	if err != nil {
		return nil, err
	}
	metadata := maps.Clone(input.Metadata)
	m.mu.Lock()
	defer m.mu.Unlock()
	m.objects[aws.ToString(input.Key)] = memoryS3Object{body: body, metadata: metadata}
	return &s3.PutObjectOutput{}, nil
}

func (m *memoryS3Client) GetObject(_ context.Context, input *s3.GetObjectInput, _ ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := aws.ToString(input.Key)
	object, ok := m.objects[key]
	if !ok {
		return nil, errors.New("no such key: " + key)
	}
	return &s3.GetObjectOutput{
		Body:     io.NopCloser(bytes.NewReader(append([]byte(nil), object.body...))),
		Metadata: object.metadata,
	}, nil
}

func (m *memoryS3Client) DeleteObject(_ context.Context, input *s3.DeleteObjectInput, _ ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := aws.ToString(input.Key)
	delete(m.objects, key)
	m.deleted = append(m.deleted, key)
	return &s3.DeleteObjectOutput{}, nil
}

func (m *memoryS3Client) ListObjectsV2(_ context.Context, input *s3.ListObjectsV2Input, _ ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	prefix := aws.ToString(input.Prefix)
	keys := make([]string, 0, len(m.objects))
	for key := range m.objects {
		if strings.HasPrefix(key, prefix) {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	contents := make([]s3types.Object, 0, len(keys))
	for _, key := range keys {
		contents = append(contents, s3types.Object{Key: aws.String(key)})
	}
	return &s3.ListObjectsV2Output{Contents: contents}, nil
}

func (m *memoryS3Client) keys() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	keys := make([]string, 0, len(m.objects))
	for key := range m.objects {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
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

func managedPlanContext(generation string, contents []byte) command.ProjectContext {
	ctx := testProjectContext()
	ctx.Pull.HeadCommit = "same-head"
	ctx.RequiresAtlantisManagedPlanFile = true
	ctx.PlanGeneration = generation
	planHash := fmt.Sprintf("%x", sha256.Sum256(contents))
	ctx.GeneratedPlanHash = &planHash
	return ctx
}

func acceptedManagedPlanContext(generation string, contents []byte) command.ProjectContext {
	ctx := testProjectContext()
	ctx.Pull.HeadCommit = "same-head"
	ctx.RequiresAtlantisManagedPlanFile = true
	ctx.AcceptedPlanGeneration = generation
	ctx.RecordedManagedPlanHash = fmt.Sprintf("%x", sha256.Sum256(contents))
	return ctx
}

func savePlanContents(t *testing.T, store *planstore.S3PlanStore, ctx command.ProjectContext, planPath string, contents []byte) {
	t.Helper()
	require.NoError(t, os.WriteFile(planPath, contents, 0o600))
	require.NoError(t, store.Save(ctx, planPath))
}

func TestS3Key_WithPrefix(t *testing.T) {
	store := planstore.NewS3PlanStoreWithClient(&mockS3Client{}, "bucket", "atlantis/plans", logging.NewNoopLogger(t))
	ctx := testProjectContext()

	key := store.TestS3Key(ctx, "/tmp/plans/myproject-default.tfplan")
	assert.Equal(t, "atlantis/plans/acme/infra/42/default/modules/vpc/myproject-default.tfplan", key)
}

func TestS3Key_WithoutPrefix(t *testing.T) {
	store := planstore.NewS3PlanStoreWithClient(&mockS3Client{}, "bucket", "", logging.NewNoopLogger(t))
	ctx := testProjectContext()

	key := store.TestS3Key(ctx, "/tmp/plans/myproject-default.tfplan")
	assert.Equal(t, "acme/infra/42/default/modules/vpc/myproject-default.tfplan", key)
}

func TestS3Key_NestedRepoRelDir(t *testing.T) {
	store := planstore.NewS3PlanStoreWithClient(&mockS3Client{}, "bucket", "pfx", logging.NewNoopLogger(t))
	ctx := testProjectContext()
	ctx.RepoRelDir = "envs/prod/us-east-1"

	key := store.TestS3Key(ctx, "/tmp/plan.tfplan")
	assert.Equal(t, "pfx/acme/infra/42/default/envs/prod/us-east-1/plan.tfplan", key)
}

func TestS3Key_TrailingSlashPrefix(t *testing.T) {
	store := planstore.NewS3PlanStoreWithClient(&mockS3Client{}, "bucket", "prefix/", logging.NewNoopLogger(t))
	ctx := testProjectContext()

	key := store.TestS3Key(ctx, "/tmp/plan.tfplan")
	assert.Equal(t, "prefix/acme/infra/42/default/modules/vpc/plan.tfplan", key)
}

func TestSave_Success(t *testing.T) {
	mock := &mockS3Client{}
	store := planstore.NewS3PlanStoreWithClient(mock, "my-bucket", "pfx", logging.NewNoopLogger(t))
	ctx := testProjectContext()
	ctx.Pull.HeadCommit = "abc123def456"
	ctx.PlanGeneration = "generation-42"
	ctx.AcceptedPlanGeneration = "generation-previously-accepted"
	managedPlanHash := "72852ea4d29c1a0726d30d75af3e47b77f650257f37777e698305bccce48f9b8"
	ctx.GeneratedPlanHash = &managedPlanHash

	planDir := t.TempDir()
	planPath := filepath.Join(planDir, "test.tfplan")
	require.NoError(t, os.WriteFile(planPath, []byte("plan-content"), 0o600))

	err := store.Save(ctx, planPath)
	require.NoError(t, err)

	assert.Equal(t, "my-bucket", *mock.putInput.Bucket)
	assert.Equal(t, "pfx/acme/infra/42/default/modules/vpc/test.tfplan", *mock.putInput.Key)
	assert.Equal(t, []byte("plan-content"), mock.putBody)
	assert.Equal(t, "abc123def456", mock.putInput.Metadata["head-commit"])
	assert.Equal(t, managedPlanHash, mock.putInput.Metadata["managed-plan-sha256"])
	assert.Equal(t, "generation-42", mock.putInput.Metadata["plan-generation"])
}

func TestSave_S3Error(t *testing.T) {
	mock := &mockS3Client{putErr: errors.New("access denied")}
	store := planstore.NewS3PlanStoreWithClient(mock, "bucket", "", logging.NewNoopLogger(t))
	ctx := testProjectContext()

	planDir := t.TempDir()
	planPath := filepath.Join(planDir, "test.tfplan")
	require.NoError(t, os.WriteFile(planPath, []byte("data"), 0o600))

	err := store.Save(ctx, planPath)
	assert.ErrorContains(t, err, "access denied")
}

func TestSave_FileOpenError(t *testing.T) {
	store := planstore.NewS3PlanStoreWithClient(&mockS3Client{}, "bucket", "", logging.NewNoopLogger(t))
	ctx := testProjectContext()

	err := store.Save(ctx, "/nonexistent/path/plan.tfplan")
	assert.ErrorContains(t, err, "opening plan file")
}

func TestManagedSave_SyntheticNonPRAPIUsesCanonicalArtifact(t *testing.T) {
	contents := []byte("synthetic-api-plan")
	objects := newMemoryS3Client()
	store := planstore.NewS3PlanStoreWithClient(objects, "bucket", "pfx", logging.NewNoopLogger(t))
	ctx := managedPlanContext("", contents)
	ctx.API = true
	ctx.Pull.Num = -1
	planPath := filepath.Join(t.TempDir(), "test.tfplan")

	savePlanContents(t, store, ctx, planPath, contents)

	assert.Equal(t, []string{"pfx/acme/infra/-1/default/modules/vpc/test.tfplan"}, objects.keys())

	loadCtx := ctx
	loadCtx.GeneratedPlanHash = nil
	loadCtx.RecordedManagedPlanHash = fmt.Sprintf("%x", sha256.Sum256(contents))
	restoredPath := filepath.Join(t.TempDir(), "test.tfplan")
	require.NoError(t, store.Load(loadCtx, restoredPath))
	restored, err := os.ReadFile(restoredPath)
	require.NoError(t, err)
	assert.Equal(t, contents, restored)
}

func TestManagedSave_PRBackedPlanRequiresGeneration(t *testing.T) {
	contents := []byte("managed-plan")
	store := planstore.NewS3PlanStoreWithClient(newMemoryS3Client(), "bucket", "pfx", logging.NewNoopLogger(t))
	ctx := managedPlanContext("", contents)
	planPath := filepath.Join(t.TempDir(), "test.tfplan")
	require.NoError(t, os.WriteFile(planPath, contents, 0o600))

	err := store.Save(ctx, planPath)

	assert.ErrorContains(t, err, "plan generation is empty")
}

func TestManagedSave_StaleSameHeadCanonicalOverwriteCannotReplaceAcceptedArtifact(t *testing.T) {
	objects := newMemoryS3Client()
	client := &blockingGenerationS3Client{
		memoryS3Client: objects,
		generation:     "generation-a",
		entered:        make(chan struct{}),
		release:        make(chan struct{}),
	}
	defer func() {
		select {
		case <-client.release:
		default:
			close(client.release)
		}
	}()
	store := planstore.NewS3PlanStoreWithClient(client, "bucket", "pfx", logging.NewNoopLogger(t))
	planPath := filepath.Join(t.TempDir(), "test.tfplan")
	planA := []byte("same-head-generation-a")
	planB := []byte("same-head-generation-b")
	ctxA := managedPlanContext("generation-a", planA)
	ctxB := managedPlanContext("generation-b", planB)

	require.NoError(t, os.WriteFile(planPath, planA, 0o600))
	staleSaveDone := make(chan error, 1)
	go func() {
		staleSaveDone <- store.Save(ctxA, planPath)
	}()
	select {
	case <-client.entered:
	case err := <-staleSaveDone:
		t.Fatalf("stale generation save returned before reaching the S3 barrier: %v", err)
	}

	savePlanContents(t, store, ctxB, planPath, planB)
	// Generation B is now the durable accepted plan. A finishes late and
	// overwrites only the canonical discovery object.
	close(client.release)
	require.NoError(t, <-staleSaveDone)

	legacyPath := filepath.Join(t.TempDir(), "test.tfplan")
	legacyCtx := testProjectContext()
	legacyCtx.Pull.HeadCommit = "same-head"
	require.NoError(t, store.Load(legacyCtx, legacyPath))
	legacyBytes, err := os.ReadFile(legacyPath)
	require.NoError(t, err)
	assert.Equal(t, planA, legacyBytes, "canonical discovery object should reproduce the stale overwrite")

	restartedStore := planstore.NewS3PlanStoreWithClient(objects, "bucket", "pfx", logging.NewNoopLogger(t))
	acceptedPath := filepath.Join(t.TempDir(), "test.tfplan")
	require.NoError(t, restartedStore.Load(acceptedManagedPlanContext("generation-b", planB), acceptedPath))
	acceptedBytes, err := os.ReadFile(acceptedPath)
	require.NoError(t, err)
	assert.Equal(t, planB, acceptedBytes)
	assert.Len(t, objects.keys(), 3, "expected one canonical and two immutable generation objects")
}

func TestManagedSave_ThreeGenerationsHostileCompletionOrderKeepsAcceptedExact(t *testing.T) {
	client := newMemoryS3Client()
	store := planstore.NewS3PlanStoreWithClient(client, "bucket", "pfx", logging.NewNoopLogger(t))
	planPath := filepath.Join(t.TempDir(), "test.tfplan")
	planA := []byte("generation-a")
	planB := []byte("generation-b-saved-but-completion-failed")
	planC := []byte("generation-c-accepted")

	// Generation order is A < B < C, but C publishes first and becomes the
	// durable accepted generation. A and then B finish their stale writes later.
	savePlanContents(t, store, managedPlanContext("generation-c", planC), planPath, planC)
	savePlanContents(t, store, managedPlanContext("generation-a", planA), planPath, planA)
	savePlanContents(t, store, managedPlanContext("generation-b", planB), planPath, planB)

	restartedStore := planstore.NewS3PlanStoreWithClient(client, "bucket", "pfx", logging.NewNoopLogger(t))
	restoredPath := filepath.Join(t.TempDir(), "test.tfplan")
	require.NoError(t, restartedStore.Load(acceptedManagedPlanContext("generation-c", planC), restoredPath))
	restored, err := os.ReadFile(restoredPath)
	require.NoError(t, err)
	assert.Equal(t, planC, restored)
	assert.Len(t, client.keys(), 4, "expected one canonical and three immutable generation objects")
}

func TestManagedRestore_CanonicalDiscoveryIsReplacedByAcceptedImmutableArtifact(t *testing.T) {
	client := newMemoryS3Client()
	store := planstore.NewS3PlanStoreWithClient(client, "bucket", "pfx", logging.NewNoopLogger(t))
	planPath := filepath.Join(t.TempDir(), "test.tfplan")
	planA := []byte("stale-discovery-plan")
	planB := []byte("accepted-plan")

	savePlanContents(t, store, managedPlanContext("generation-b", planB), planPath, planB)
	savePlanContents(t, store, managedPlanContext("generation-a", planA), planPath, planA)

	workspaces, err := store.ListWorkspaces("acme", "infra", 42)
	require.NoError(t, err)
	assert.Equal(t, []string{"default"}, workspaces)

	pullDir := t.TempDir()
	require.NoError(t, store.RestorePlans(pullDir, "acme", "infra", 42))
	restoredPath := filepath.Join(pullDir, "default", "modules", "vpc", "test.tfplan")
	discoveryBytes, err := os.ReadFile(restoredPath)
	require.NoError(t, err)
	assert.Equal(t, planA, discoveryBytes)

	require.NoError(t, store.Load(acceptedManagedPlanContext("generation-b", planB), restoredPath))
	acceptedBytes, err := os.ReadFile(restoredPath)
	require.NoError(t, err)
	assert.Equal(t, planB, acceptedBytes)
}

func TestRestorePlans_PrefixContainingManagedMarkerStillRestoresCanonicalPlan(t *testing.T) {
	client := newMemoryS3Client()
	store := planstore.NewS3PlanStoreWithClient(client, "bucket", "team/.atlantis-managed/plans", logging.NewNoopLogger(t))
	planPath := filepath.Join(t.TempDir(), "test.tfplan")
	ctx := testProjectContext()
	ctx.Pull.HeadCommit = "same-head"
	require.NoError(t, os.WriteFile(planPath, []byte("canonical-plan"), 0o600))
	require.NoError(t, store.Save(ctx, planPath))

	workspaces, err := store.ListWorkspaces("acme", "infra", 42)
	require.NoError(t, err)
	assert.Equal(t, []string{"default"}, workspaces)
	pullDir := t.TempDir()
	require.NoError(t, store.RestorePlans(pullDir, "acme", "infra", 42))
	restored, err := os.ReadFile(filepath.Join(pullDir, "default", "modules", "vpc", "test.tfplan"))
	require.NoError(t, err)
	assert.Equal(t, []byte("canonical-plan"), restored)
}

func TestSave_ManagedRejectsBytesDifferentFromRecordedHash(t *testing.T) {
	client := newMemoryS3Client()
	store := planstore.NewS3PlanStoreWithClient(client, "bucket", "pfx", logging.NewNoopLogger(t))
	planPath := filepath.Join(t.TempDir(), "test.tfplan")
	ctx := managedPlanContext("generation-a", []byte("expected-plan"))
	require.NoError(t, os.WriteFile(planPath, []byte("different-plan"), 0o600))

	err := store.Save(ctx, planPath)
	require.ErrorContains(t, err, "managed plan hash changed before S3 upload")
	assert.Empty(t, client.keys())
}

func TestLoad_SuccessWithoutManagedPlanMetadata(t *testing.T) {
	planContent := []byte("downloaded-plan-data")
	mock := &mockS3Client{
		getBody:     planContent,
		getMetadata: map[string]string{"Head-Commit": "abc123"},
	}
	store := planstore.NewS3PlanStoreWithClient(mock, "bucket", "pfx", logging.NewNoopLogger(t))
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

func TestLoad_RejectsPartialDurableManagedPlanIdentity(t *testing.T) {
	store := planstore.NewS3PlanStoreWithClient(newMemoryS3Client(), "bucket", "pfx", logging.NewNoopLogger(t))
	ctx := testProjectContext()
	ctx.RecordedManagedPlanHash = "hash-without-generation"

	err := store.Load(ctx, filepath.Join(t.TempDir(), "test.tfplan"))

	assert.ErrorContains(t, err, "durable plan hash and accepted generation must both be set")
}

func TestLoad_StalePlanRejected(t *testing.T) {
	mock := &mockS3Client{
		getBody:     []byte("old-plan"),
		getMetadata: map[string]string{"Head-Commit": "oldcommit"},
	}
	store := planstore.NewS3PlanStoreWithClient(mock, "bucket", "", logging.NewNoopLogger(t))
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
	store := planstore.NewS3PlanStoreWithClient(mock, "bucket", "", logging.NewNoopLogger(t))
	ctx := testProjectContext()
	ctx.Pull.HeadCommit = "abc123"

	err := store.Load(ctx, filepath.Join(t.TempDir(), "plan.tfplan"))
	assert.ErrorContains(t, err, "no head-commit metadata")
}

func TestLoad_S3Error(t *testing.T) {
	mock := &mockS3Client{getErr: errors.New("no such key")}
	store := planstore.NewS3PlanStoreWithClient(mock, "bucket", "", logging.NewNoopLogger(t))
	ctx := testProjectContext()

	err := store.Load(ctx, "/tmp/nonexistent/plan.tfplan")
	assert.ErrorContains(t, err, "no such key")
}

func TestRemove_DeletesLegacyArtifact(t *testing.T) {
	mock := &mockS3Client{}
	store := planstore.NewS3PlanStoreWithClient(mock, "my-bucket", "pfx", logging.NewNoopLogger(t))
	ctx := testProjectContext()

	planDir := t.TempDir()
	planPath := filepath.Join(planDir, "test.tfplan")
	require.NoError(t, os.WriteFile(planPath, []byte("data"), 0o600))

	err := store.Remove(ctx, planPath)
	require.NoError(t, err)

	assert.Equal(t, "my-bucket", *mock.deleteInput.Bucket)
	assert.Equal(t, "pfx/acme/infra/42/default/modules/vpc/test.tfplan", *mock.deleteInput.Key)
	assert.Nil(t, mock.deleteInput.IfMatch)

	_, statErr := os.Stat(planPath)
	assert.True(t, os.IsNotExist(statErr))
}

func TestRemove_S3Error(t *testing.T) {
	mock := &mockS3Client{deleteErr: errors.New("forbidden")}
	store := planstore.NewS3PlanStoreWithClient(mock, "bucket", "", logging.NewNoopLogger(t))
	ctx := testProjectContext()

	err := store.Remove(ctx, "/tmp/whatever.tfplan")
	assert.NoError(t, err)
}

func TestRemove_LocalFileAlreadyGone(t *testing.T) {
	mock := &mockS3Client{}
	store := planstore.NewS3PlanStoreWithClient(mock, "bucket", "", logging.NewNoopLogger(t))
	ctx := testProjectContext()

	err := store.Remove(ctx, "/tmp/nonexistent-plan-file.tfplan")
	require.NoError(t, err)
}

func TestRemove_DigestBoundMatchingArtifactIsRetained(t *testing.T) {
	planContents := []byte("accepted-plan")
	planHash := fmt.Sprintf("%x", sha256.Sum256(planContents))
	mock := &mockS3Client{}
	store := planstore.NewS3PlanStoreWithClient(mock, "bucket", "", logging.NewNoopLogger(t))
	ctx := testProjectContext()
	ctx.RecordedManagedPlanHash = planHash
	ctx.AcceptedPlanGeneration = "generation-accepted"
	ctx.PlanGeneration = "generation-active-unrelated"
	planPath := filepath.Join(t.TempDir(), "plan.tfplan")
	require.NoError(t, os.WriteFile(planPath, planContents, 0o600))

	require.NoError(t, store.Remove(ctx, planPath))

	assert.Nil(t, mock.deleteInput)
	_, err := os.Stat(planPath)
	require.NoError(t, err)
}

func TestRemove_DigestBoundNewerArtifactIsPreserved(t *testing.T) {
	acceptedPlan := []byte("accepted-plan")
	newerPlan := []byte("accepted-plan")
	mock := &mockS3Client{}
	store := planstore.NewS3PlanStoreWithClient(mock, "bucket", "", logging.NewNoopLogger(t))
	ctx := testProjectContext()
	ctx.RecordedManagedPlanHash = fmt.Sprintf("%x", sha256.Sum256(acceptedPlan))
	ctx.AcceptedPlanGeneration = "generation-old"
	ctx.PlanGeneration = "generation-newer"
	planPath := filepath.Join(t.TempDir(), "plan.tfplan")
	require.NoError(t, os.WriteFile(planPath, newerPlan, 0o600))

	require.NoError(t, store.Remove(ctx, planPath))

	assert.Nil(t, mock.deleteInput)
	got, err := os.ReadFile(planPath)
	require.NoError(t, err)
	assert.Equal(t, newerPlan, got)
}

func TestRemove_DigestBoundWithoutAcceptedGenerationDoesNotAttemptAutomaticDelete(t *testing.T) {
	planContents := []byte("accepted-plan")
	planHash := fmt.Sprintf("%x", sha256.Sum256(planContents))
	mock := &mockS3Client{}
	store := planstore.NewS3PlanStoreWithClient(mock, "bucket", "", logging.NewNoopLogger(t))
	ctx := testProjectContext()
	ctx.RecordedManagedPlanHash = planHash
	planPath := filepath.Join(t.TempDir(), "plan.tfplan")
	require.NoError(t, os.WriteFile(planPath, planContents, 0o600))

	require.NoError(t, store.Remove(ctx, planPath))

	assert.Nil(t, mock.deleteInput)
	_, err := os.Stat(planPath)
	require.NoError(t, err)
}

func TestRemove_SyntheticNonPRAPIDeletesDigestBoundCanonicalArtifact(t *testing.T) {
	planContents := []byte("synthetic-api-plan")
	mock := &mockS3Client{}
	store := planstore.NewS3PlanStoreWithClient(mock, "bucket", "pfx", logging.NewNoopLogger(t))
	ctx := testProjectContext()
	ctx.API = true
	ctx.Pull.Num = -1
	ctx.RecordedManagedPlanHash = fmt.Sprintf("%x", sha256.Sum256(planContents))
	planPath := filepath.Join(t.TempDir(), "plan.tfplan")
	require.NoError(t, os.WriteFile(planPath, planContents, 0o600))

	require.NoError(t, store.Remove(ctx, planPath))

	require.NotNil(t, mock.deleteInput)
	assert.Equal(t, "pfx/acme/infra/-1/default/modules/vpc/plan.tfplan", aws.ToString(mock.deleteInput.Key))
	_, err := os.Stat(planPath)
	assert.True(t, os.IsNotExist(err))
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
	store := planstore.NewS3PlanStoreWithClient(mock, "bucket", "pfx", logging.NewNoopLogger(t))
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
	store := planstore.NewS3PlanStoreWithClient(mock, "bucket", "pfx", logging.NewNoopLogger(t))

	err := store.RestorePlans(t.TempDir(), "acme", "infra", 42)
	require.NoError(t, err)
}

func TestRestorePlans_ListError(t *testing.T) {
	mock := &mockS3Client{listErr: errors.New("access denied")}
	store := planstore.NewS3PlanStoreWithClient(mock, "bucket", "pfx", logging.NewNoopLogger(t))

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
	store := planstore.NewS3PlanStoreWithClient(mock, "bucket", "", logging.NewNoopLogger(t))
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
	store := planstore.NewS3PlanStoreWithClient(mock, "bucket", "pfx", logging.NewNoopLogger(t))

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
	store := planstore.NewS3PlanStoreWithClient(mock, "bucket", "pfx", logging.NewNoopLogger(t))

	err := store.DeleteForPull("acme", "infra", 42)
	require.NoError(t, err)
	assert.Empty(t, mock.deletedKeys)
}

func TestDeleteForPull_ListError(t *testing.T) {
	mock := &mockS3Client{listErr: errors.New("access denied")}
	store := planstore.NewS3PlanStoreWithClient(mock, "bucket", "pfx", logging.NewNoopLogger(t))

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
	store := planstore.NewS3PlanStoreWithClient(mock, "bucket", "pfx", logging.NewNoopLogger(t))

	// S3 delete errors during cleanup are logged but not returned (soft-fail).
	err := store.DeleteForPull("acme", "infra", 42)
	assert.NoError(t, err)
}

func TestManagedPlanCleanup_DeletesCanonicalAndImmutableArtifacts(t *testing.T) {
	client := newMemoryS3Client()
	store := planstore.NewS3PlanStoreWithClient(client, "bucket", "pfx", logging.NewNoopLogger(t))
	planPath := filepath.Join(t.TempDir(), "test-default.tfplan")
	planA := []byte("project-one-generation-a")
	planB := []byte("project-one-generation-b")
	ctxA := managedPlanContext("generation-a", planA)
	ctxA.ProjectName = "test"
	ctxB := managedPlanContext("generation-b", planB)
	ctxB.ProjectName = "test"
	savePlanContents(t, store, ctxA, planPath, planA)
	savePlanContents(t, store, ctxB, planPath, planB)

	otherCtx := managedPlanContext("generation-c", []byte("project-two-generation-c"))
	otherCtx.RepoRelDir = "modules/rds"
	otherCtx.ProjectName = "other"
	otherPath := filepath.Join(t.TempDir(), "other-default.tfplan")
	savePlanContents(t, store, otherCtx, otherPath, []byte("project-two-generation-c"))

	require.NoError(t, store.DeletePlanForProject("acme", "infra", 42, "default", "modules/vpc", "test"))
	remaining := client.keys()
	require.Len(t, remaining, 2, "other project should retain its canonical and immutable objects")
	for _, key := range remaining {
		assert.Contains(t, key, "modules/rds/other-default.tfplan")
	}

	require.NoError(t, store.DeleteForPull("acme", "infra", 42))
	assert.Empty(t, client.keys())
}
