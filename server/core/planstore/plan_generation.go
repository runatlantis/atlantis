// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package planstore

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/runatlantis/atlantis/server/events/command"
)

// generationSuffix is below an individual plan filename, preserving sibling
// projects even when their workspace and directory are the same. Its objects
// are not convention .tfplan files and must not be discovered as pending plans.
const generationSuffix = ".atlantis-managed"

func generationPath(canonical, generation, digest string) (string, error) {
	if generation == "" || strings.ContainsAny(generation, `/\\`) || generation == "." || generation == ".." {
		return "", fmt.Errorf("invalid managed plan generation")
	}
	decoded, err := hex.DecodeString(digest)
	if err != nil || len(decoded) != sha256.Size {
		return "", fmt.Errorf("accepted managed plan hash is missing or invalid; run `atlantis plan` again")
	}
	return filepath.Join(canonical+generationSuffix, generation, digest), nil
}

func canonicalPlanPath(ctx command.ProjectContext, planPath string) string {
	if ctx.CanonicalPlanPath != "" {
		return ctx.CanonicalPlanPath
	}
	return planPath
}

// StagePlan gives Terraform an operation-owned file. Custom scripts continue to
// receive the convention PLANFILE; their arbitrary reads/writes are not isolated
// by this helper. The saved artifact is captured from this staging file only.
func StagePlan(ctx command.ProjectContext, canonical string) (string, func(), error) {
	if ctx.PlanGeneration == "" {
		return canonical, func() {}, nil
	}
	parent, err := generationPath(canonical, ctx.PlanGeneration, strings.Repeat("0", sha256.Size*2))
	if err != nil {
		return "", nil, err
	}
	if err := os.MkdirAll(filepath.Dir(parent), 0o700); err != nil {
		return "", nil, fmt.Errorf("creating managed plan directory: %w", err)
	}
	dir, err := os.MkdirTemp(filepath.Dir(parent), "staging-")
	if err != nil {
		return "", nil, fmt.Errorf("creating plan staging directory: %w", err)
	}
	return filepath.Join(dir, "plan.tfplan"), func() { _ = os.RemoveAll(dir) }, nil
}

func readGenerationPlan(ctx command.ProjectContext, planPath string) ([]byte, string, string, error) {
	if ctx.SavedPlanHash == nil {
		return nil, "", "", fmt.Errorf("managed plan command has no saved digest receiver")
	}
	*ctx.SavedPlanHash = ""
	contents, err := os.ReadFile(planPath)
	if err != nil {
		return nil, "", "", fmt.Errorf("reading managed plan for storage: %w", err)
	}
	sum := sha256.Sum256(contents)
	digest := hex.EncodeToString(sum[:])
	identityPath, err := generationPath(canonicalPlanPath(ctx, planPath), ctx.PlanGeneration, digest)
	return contents, digest, identityPath, err
}

// replacePlan installs one complete buffer. It never exposes a partially written
// plan and preserves a private 0600 file mode for Terraform's sensitive data.
func replacePlan(path string, contents []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".plan-")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err := f.Write(contents); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err := os.Rename(f.Name(), path); err != nil {
		return err
	}
	parent, err := os.Open(filepath.Dir(path))
	if err != nil {
		return err
	}
	defer parent.Close()
	return parent.Sync()
}

func verifyGenerationPlan(contents []byte, expected string) error {
	sum := sha256.Sum256(contents)
	if expected == "" || hex.EncodeToString(sum[:]) != expected {
		return fmt.Errorf("accepted managed plan bytes do not match durable hash; run `atlantis plan` again")
	}
	return nil
}

func saveLocalGeneration(ctx command.ProjectContext, planPath string) error {
	contents, digest, identityPath, err := readGenerationPlan(ctx, planPath)
	if err != nil {
		return err
	}
	if err := replacePlan(identityPath, contents); err != nil {
		return fmt.Errorf("saving immutable managed plan: %w", err)
	}
	if err := replacePlan(canonicalPlanPath(ctx, planPath), contents); err != nil {
		return fmt.Errorf("publishing convention plan file: %w", err)
	}
	*ctx.SavedPlanHash = digest
	return nil
}

func loadLocalGeneration(ctx command.ProjectContext, planPath string) error {
	identityPath, err := generationPath(planPath, ctx.AcceptedPlanGeneration, ctx.ExpectedPlanHash)
	if err != nil {
		return err
	}
	contents, err := os.ReadFile(identityPath)
	if err != nil {
		return fmt.Errorf("reading accepted managed plan: %w", err)
	}
	if err := verifyGenerationPlan(contents, ctx.ExpectedPlanHash); err != nil {
		return err
	}
	return replacePlan(planPath, contents)
}
