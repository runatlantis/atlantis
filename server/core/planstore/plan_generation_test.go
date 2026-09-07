// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package planstore_test

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/core/planstore"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/stretchr/testify/require"
)

type generationS3Client struct {
	mockS3Client
	beforeUpload func()
}

func (m *generationS3Client) PutObject(_ context.Context, input *s3.PutObjectInput, _ ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	if m.beforeUpload != nil {
		m.beforeUpload()
	}
	if m.putErr != nil {
		return nil, m.putErr
	}
	body, err := io.ReadAll(input.Body)
	if err != nil {
		return nil, err
	}
	m.getObjects[aws.ToString(input.Key)] = body
	return &s3.PutObjectOutput{}, nil
}
func (m *generationS3Client) DeleteObject(ctx context.Context, input *s3.DeleteObjectInput, options ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	delete(m.getObjects, aws.ToString(input.Key))
	return m.mockS3Client.DeleteObject(ctx, input, options...)
}

func TestManagedGeneration_LateSaveCannotReplaceAcceptedArtifact(t *testing.T) {
	for _, external := range []bool{false, true} {
		name := "local"
		if external {
			name = "s3"
		}
		t.Run(name, func(t *testing.T) {
			client := &generationS3Client{mockS3Client: mockS3Client{getObjects: map[string][]byte{}}}
			var store planstore.PlanStore = &planstore.LocalPlanStore{}
			if external {
				store = planstore.NewS3PlanStoreWithClient(client, "bucket", "", logging.NewNoopLogger(t))
			}
			canonical := filepath.Join(t.TempDir(), "default.tfplan")
			ctx := testProjectContext()
			ctx.Pull.HeadCommit = "same-pr-head-for-both-generations"
			ctx.RequiresAtlantisManagedPlanFile = true
			ctx.CanonicalPlanPath = canonical
			g1, err := db.BeginPlanGeneration(nil, ctx.Pull, "G1", []command.ProjectContext{ctx}, true)
			require.NoError(t, err)
			ctx.PlanGeneration = "G1"
			staging1, cleanup1, err := planstore.StagePlan(ctx, canonical)
			require.NoError(t, err)
			defer cleanup1()
			require.NoError(t, os.WriteFile(staging1, []byte("G1 bytes"), 0o600))
			g2, err := db.BeginPlanGeneration(&g1.PullStatus, ctx.Pull, "G2", []command.ProjectContext{ctx}, true)
			require.NoError(t, err)
			ctx.PlanGeneration = "G2"
			staging2, cleanup2, err := planstore.StagePlan(ctx, canonical)
			require.NoError(t, err)
			defer cleanup2()
			require.NoError(t, os.WriteFile(staging2, []byte("G2 accepted bytes"), 0o600))
			var hash2 string
			ctx.SavedPlanHash = &hash2
			require.NoError(t, store.Save(ctx, staging2))
			result := command.ProjectResult{Command: command.Plan, Workspace: ctx.Workspace, RepoRelDir: ctx.RepoRelDir, PlanGeneration: "G2", ManagedPlanHash: hash2, ProjectCommandOutput: command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{}}}
			accepted, err := db.MergePullResults(&g2.PullStatus, ctx.Pull, []command.ProjectResult{result})
			require.NoError(t, err)
			ctx.PlanGeneration = "G1"
			var hash1 string
			ctx.SavedPlanHash = &hash1
			require.NoError(t, store.Save(ctx, staging1))
			result.PlanGeneration, result.ManagedPlanHash = "G1", hash1
			_, err = db.MergePullResults(&accepted, ctx.Pull, []command.ProjectResult{result})
			require.ErrorIs(t, err, db.ErrPlanGenerationSuperseded)
			// Reaping the rejected artifact must preserve the accepted G2 identity.
			ctx.AcceptedPlanGeneration, ctx.ExpectedPlanHash = "G1", hash1
			require.NoError(t, store.Remove(ctx, canonical))
			if external {
				// A fresh replica has no local plan or generation cache at all.
				canonical = filepath.Join(t.TempDir(), "default.tfplan")
				store = planstore.NewS3PlanStoreWithClient(client, "bucket", "", logging.NewNoopLogger(t))
			}
			ctx.CanonicalPlanPath = canonical
			ctx.PlanGeneration = accepted.Projects[0].PlanGeneration
			ctx.AcceptedPlanGeneration = accepted.Projects[0].AcceptedPlanGeneration
			ctx.ExpectedPlanHash = accepted.Projects[0].ManagedPlanHash
			require.NoError(t, store.Load(ctx, canonical))
			contents, err := os.ReadFile(canonical)
			require.NoError(t, err)
			require.Equal(t, "G2 accepted bytes", string(contents))
			info, err := os.Stat(canonical)
			require.NoError(t, err)
			require.Equal(t, os.FileMode(0o600), info.Mode().Perm())
			ctx.ExpectedPlanHash = ""
			require.Error(t, store.Load(ctx, canonical))
		})
	}
}

func TestManagedGeneration_UploadUsesHashedBuffer(t *testing.T) {
	client := &generationS3Client{mockS3Client: mockS3Client{getObjects: map[string][]byte{}}}
	store := planstore.NewS3PlanStoreWithClient(client, "bucket", "", logging.NewNoopLogger(t))
	ctx := testProjectContext()
	ctx.PlanGeneration = "G1"
	ctx.CanonicalPlanPath = filepath.Join(t.TempDir(), "default.tfplan")
	var digest string
	ctx.SavedPlanHash = &digest
	staging, cleanup, err := planstore.StagePlan(ctx, ctx.CanonicalPlanPath)
	require.NoError(t, err)
	defer cleanup()
	require.NoError(t, os.WriteFile(staging, []byte("original"), 0o600))
	client.beforeUpload = func() { require.NoError(t, os.WriteFile(staging, []byte("mutated"), 0o600)) }
	require.NoError(t, store.Save(ctx, staging))
	ctx.AcceptedPlanGeneration, ctx.ExpectedPlanHash = "G1", digest
	require.NoError(t, store.Load(ctx, ctx.CanonicalPlanPath))
	contents, err := os.ReadFile(ctx.CanonicalPlanPath)
	require.NoError(t, err)
	require.Equal(t, "original", string(contents))
	client.putErr = errors.New("upload unavailable")
	require.ErrorContains(t, store.Save(ctx, staging), "upload unavailable")
	require.Empty(t, digest, "failed save must not install an accepted digest")
}

func TestManagedGeneration_ProjectCleanupPreservesSibling(t *testing.T) {
	client := &generationS3Client{mockS3Client: mockS3Client{getObjects: map[string][]byte{}}}
	store := planstore.NewS3PlanStoreWithClient(client, "bucket", "", logging.NewNoopLogger(t))
	ctx := testProjectContext()
	ctx.PlanGeneration = "G1"
	dir := t.TempDir()
	paths := []string{filepath.Join(dir, "a-default.tfplan"), filepath.Join(dir, "b-default.tfplan")}
	digests := make([]string, len(paths))
	for i, path := range paths {
		ctx.CanonicalPlanPath = path
		ctx.SavedPlanHash = &digests[i]
		require.NoError(t, os.WriteFile(path, []byte(path), 0o600))
		require.NoError(t, store.Save(ctx, path))
	}
	ctx.CanonicalPlanPath = paths[0]
	ctx.AcceptedPlanGeneration, ctx.ExpectedPlanHash = "G1", digests[0]
	require.NoError(t, store.Remove(ctx, paths[0]))
	require.Error(t, store.Load(ctx, paths[0]))
	ctx.CanonicalPlanPath = paths[1]
	ctx.ExpectedPlanHash = digests[1]
	require.NoError(t, store.Load(ctx, paths[1]))
	contents, err := os.ReadFile(paths[1])
	require.NoError(t, err)
	require.Equal(t, paths[1], string(contents))
}

func TestLegacyPullCleanupPreservesAcceptedGenerationObjects(t *testing.T) {
	client := &generationS3Client{mockS3Client: mockS3Client{getObjects: map[string][]byte{}}}
	store := planstore.NewS3PlanStoreWithClient(client, "bucket", "", logging.NewNoopLogger(t))
	ctx := testProjectContext()
	ctx.PlanGeneration = "G2"
	ctx.SavedPlanHash = new(string)
	path := filepath.Join(t.TempDir(), "default.tfplan")
	require.NoError(t, os.WriteFile(path, []byte("accepted G2"), 0o600))
	require.NoError(t, store.Save(ctx, path))
	legacyKey := store.TestS3Key(ctx, path)
	client.getObjects[legacyKey] = []byte("late legacy bytes")
	client.listOutput = &s3.ListObjectsV2Output{}
	for key := range client.getObjects {
		client.listOutput.Contents = append(client.listOutput.Contents, s3types.Object{Key: aws.String(key)})
	}
	require.NoError(t, store.DeleteLegacyForPull(ctx.BaseRepo.Owner, ctx.BaseRepo.Name, ctx.Pull.Num))
	require.NotContains(t, client.getObjects, legacyKey)
	ctx.AcceptedPlanGeneration, ctx.ExpectedPlanHash = "G2", *ctx.SavedPlanHash
	require.NoError(t, store.Load(ctx, path))
	contents, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, "accepted G2", string(contents))
	// Pull close still reaps everything, including generation identities.
	require.NoError(t, store.DeleteForPull(ctx.BaseRepo.Owner, ctx.BaseRepo.Name, ctx.Pull.Num))
	require.Empty(t, client.getObjects)
}

func TestManagedGeneration_RemediationWithoutAcceptedArtifact(t *testing.T) {
	client := &generationS3Client{mockS3Client: mockS3Client{getObjects: map[string][]byte{}}}
	for name, store := range map[string]planstore.PlanStore{
		"local": &planstore.LocalPlanStore{},
		"s3":    planstore.NewS3PlanStoreWithClient(client, "bucket", "", logging.NewNoopLogger(t)),
	} {
		t.Run(name, func(t *testing.T) {
			ctx := testProjectContext()
			ctx.PlanGeneration = "failed-generation"
			path := filepath.Join(t.TempDir(), "default.tfplan")
			require.NoError(t, os.WriteFile(path, []byte("incomplete plan"), 0600))
			require.NoError(t, store.Remove(ctx, path))
			_, err := os.Stat(path)
			require.True(t, os.IsNotExist(err))
			require.Empty(t, client.deletedKeys)
		})
	}
}
