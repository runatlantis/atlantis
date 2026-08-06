// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package planstore

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/runatlantis/atlantis/server/core/backends"
	"github.com/runatlantis/atlantis/server/logging"
)

// s3StoreSegment namespaces this store's objects under the backend's
// prefix, so stores sharing the backend cannot collide.
const s3StoreSegment = "plans"

// s3PlanPrefix appends this store's segment to the backend's prefix.
func s3PlanPrefix(backendPrefix string) string {
	if p := strings.Trim(backendPrefix, "/"); p != "" {
		return p + "/" + s3StoreSegment
	}
	return s3StoreSegment
}

// New builds the plan store on the given backend.
func New(logger logging.SimpleLogging, backend backends.Backend) (PlanStore, error) {
	switch b := backend.(type) {
	case *backends.FilesystemBackend:
		return &LocalPlanStore{}, nil
	case *backends.S3Backend:
		prefix := s3PlanPrefix(b.Prefix)
		logger.Info("using S3 plan store (bucket=%s, prefix=%s)", b.Bucket, prefix)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if _, err := b.Client.HeadBucket(ctx, &s3.HeadBucketInput{
			Bucket: aws.String(b.Bucket),
		}); err != nil {
			return nil, fmt.Errorf("validating S3 plan store bucket %q: %w", b.Bucket, err)
		}
		return NewS3PlanStoreWithClient(b.Client, b.Bucket, prefix, logger), nil
	default:
		return nil, fmt.Errorf("the plan store has no %s driver; supported backends: filesystem, s3", backend.Kind())
	}
}
