package event_test

import (
	"bytes"
	"context"
	"github.com/runatlantis/atlantis/server/events/models"
	buffered "github.com/runatlantis/atlantis/server/http"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/lyft/feature"
	"github.com/runatlantis/atlantis/server/neptune/gateway/event"
	"github.com/runatlantis/atlantis/server/neptune/sync"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"testing"
)

const repoFullName = "repo"

func buildRequest(t *testing.T) *buffered.BufferedRequest {
	requestBody := "body"
	rawRequest, err := http.NewRequest(http.MethodPost, "", io.NopCloser(bytes.NewBuffer([]byte(requestBody))))
	assert.NoError(t, err)
	r, err := buffered.NewBufferedRequest(rawRequest)
	assert.NoError(t, err)
	return r
}

func TestPullRequestReviewWorkerProxy_HandleSuccessWithFailedPolicies(t *testing.T) {
	writer := &mockSnsWriter{}
	allocator := &testAllocator{
		t:                 t,
		expectedFeatureID: feature.PolicyV2,
		expectedFeatureCtx: feature.FeatureContext{
			RepoName: repoFullName,
		},
		expectedAllocation: true,
	}
	mockFetcher := &mockCheckRunFetcher{
		failedPolicies: []string{"failed policy"},
	}
	logger := logging.NewNoopCtxLogger(t)
	proxy := event.PullRequestReviewWorkerProxy{
		Scheduler:       &sync.SynchronousScheduler{Logger: logger},
		SnsWriter:       writer,
		Logger:          logger,
		Allocator:       allocator,
		CheckRunFetcher: mockFetcher,
	}
	prrEvent := event.PullRequestReview{
		State: event.Approved,
		Repo:  models.Repo{FullName: repoFullName},
	}
	err := proxy.Handle(context.Background(), prrEvent, buildRequest(t))
	assert.NoError(t, err)
	assert.True(t, writer.isCalled)
	assert.True(t, mockFetcher.isCalled)
}

func TestPullRequestReviewWorkerProxy_HandleSuccessNoFailedPolicies(t *testing.T) {
	writer := &mockSnsWriter{}
	allocator := &testAllocator{
		t:                 t,
		expectedFeatureID: feature.PolicyV2,
		expectedFeatureCtx: feature.FeatureContext{
			RepoName: repoFullName,
		},
		expectedAllocation: true,
	}
	mockFetcher := &mockCheckRunFetcher{}
	logger := logging.NewNoopCtxLogger(t)
	proxy := event.PullRequestReviewWorkerProxy{
		Scheduler:       &sync.SynchronousScheduler{Logger: logger},
		SnsWriter:       writer,
		Logger:          logger,
		Allocator:       allocator,
		CheckRunFetcher: mockFetcher,
	}
	prrEvent := event.PullRequestReview{
		State: event.Approved,
		Repo:  models.Repo{FullName: repoFullName},
	}
	err := proxy.Handle(context.Background(), prrEvent, buildRequest(t))
	assert.NoError(t, err)
	assert.False(t, writer.isCalled)
	assert.True(t, mockFetcher.isCalled)
}

func TestPullRequestReviewWorkerProxy_AllocationError(t *testing.T) {
	writer := &mockSnsWriter{}
	allocator := &testAllocator{
		t:                 t,
		expectedFeatureID: feature.PolicyV2,
		expectedFeatureCtx: feature.FeatureContext{
			RepoName: repoFullName,
		},
		expectedError: assert.AnError,
	}
	mockFetcher := &mockCheckRunFetcher{}
	logger := logging.NewNoopCtxLogger(t)
	proxy := event.PullRequestReviewWorkerProxy{
		Scheduler:       &sync.SynchronousScheduler{Logger: logger},
		SnsWriter:       writer,
		Logger:          logger,
		Allocator:       allocator,
		CheckRunFetcher: mockFetcher,
	}
	prrEvent := event.PullRequestReview{
		State: event.Approved,
		Repo:  models.Repo{FullName: repoFullName},
	}
	err := proxy.Handle(context.Background(), prrEvent, buildRequest(t))
	assert.Error(t, err)
	assert.False(t, writer.isCalled)
	assert.False(t, mockFetcher.isCalled)
}

func TestPullRequestReviewWorkerProxy_AllocationFalse(t *testing.T) {
	writer := &mockSnsWriter{}
	allocator := &testAllocator{
		t:                 t,
		expectedFeatureID: feature.PolicyV2,
		expectedFeatureCtx: feature.FeatureContext{
			RepoName: repoFullName,
		},
	}
	mockFetcher := &mockCheckRunFetcher{}
	logger := logging.NewNoopCtxLogger(t)
	proxy := event.PullRequestReviewWorkerProxy{
		Scheduler:       &sync.SynchronousScheduler{Logger: logger},
		SnsWriter:       writer,
		Logger:          logger,
		Allocator:       allocator,
		CheckRunFetcher: mockFetcher,
	}
	prrEvent := event.PullRequestReview{
		State: event.Approved,
		Repo:  models.Repo{FullName: repoFullName},
	}
	err := proxy.Handle(context.Background(), prrEvent, buildRequest(t))
	assert.NoError(t, err)
	assert.False(t, writer.isCalled)
	assert.False(t, mockFetcher.isCalled)
}

func TestPullRequestReviewWorkerProxy_NotApprovalEvent(t *testing.T) {
	writer := &mockSnsWriter{}
	allocator := &testAllocator{
		t:                 t,
		expectedFeatureID: feature.PolicyV2,
		expectedFeatureCtx: feature.FeatureContext{
			RepoName: repoFullName,
		},
		expectedAllocation: true,
	}
	mockFetcher := &mockCheckRunFetcher{}
	logger := logging.NewNoopCtxLogger(t)
	proxy := event.PullRequestReviewWorkerProxy{
		Scheduler:       &sync.SynchronousScheduler{Logger: logger},
		SnsWriter:       writer,
		Logger:          logger,
		Allocator:       allocator,
		CheckRunFetcher: mockFetcher,
	}
	prrEvent := event.PullRequestReview{
		State: "something else",
		Repo:  models.Repo{FullName: repoFullName},
	}
	err := proxy.Handle(context.Background(), prrEvent, buildRequest(t))
	assert.NoError(t, err)
	assert.False(t, writer.isCalled)
	assert.False(t, mockFetcher.isCalled)
}

func TestPullRequestReviewWorkerProxy_FetcherError(t *testing.T) {
	writer := &mockSnsWriter{}
	allocator := &testAllocator{
		t:                 t,
		expectedFeatureID: feature.PolicyV2,
		expectedFeatureCtx: feature.FeatureContext{
			RepoName: repoFullName,
		},
		expectedAllocation: true,
	}
	mockFetcher := &mockCheckRunFetcher{
		err: assert.AnError,
	}
	logger := logging.NewNoopCtxLogger(t)
	proxy := event.PullRequestReviewWorkerProxy{
		Scheduler:       &sync.SynchronousScheduler{Logger: logger},
		SnsWriter:       writer,
		Logger:          logger,
		Allocator:       allocator,
		CheckRunFetcher: mockFetcher,
	}
	prrEvent := event.PullRequestReview{
		State: event.Approved,
		Repo:  models.Repo{FullName: repoFullName},
	}
	err := proxy.Handle(context.Background(), prrEvent, buildRequest(t))
	assert.Error(t, err)
	assert.False(t, writer.isCalled)
	assert.True(t, mockFetcher.isCalled)
}

func TestPullRequestReviewWorkerProxy_SNSError(t *testing.T) {
	writer := &mockSnsWriter{}
	allocator := &testAllocator{
		t:                 t,
		expectedFeatureID: feature.PolicyV2,
		expectedFeatureCtx: feature.FeatureContext{
			RepoName: repoFullName,
		},
		expectedAllocation: true,
	}
	mockFetcher := &mockCheckRunFetcher{
		failedPolicies: []string{"failed policy"},
	}
	logger := logging.NewNoopCtxLogger(t)
	proxy := event.PullRequestReviewWorkerProxy{
		Scheduler:       &sync.SynchronousScheduler{Logger: logger},
		SnsWriter:       writer,
		Logger:          logger,
		Allocator:       allocator,
		CheckRunFetcher: mockFetcher,
	}
	prrEvent := event.PullRequestReview{
		State: event.Approved,
		Repo:  models.Repo{FullName: repoFullName},
	}

	err := proxy.Handle(context.Background(), prrEvent, buildRequest(t))
	assert.NoError(t, err)
	assert.True(t, writer.isCalled)
	assert.True(t, mockFetcher.isCalled)
}

type mockSnsWriter struct {
	err      error
	isCalled bool
}

func (s *mockSnsWriter) WriteWithContext(ctx context.Context, payload []byte) error {
	s.isCalled = true
	return s.err
}

type mockCheckRunFetcher struct {
	failedPolicies []string
	err            error
	isCalled       bool
}

func (f *mockCheckRunFetcher) ListFailedPolicyCheckRuns(_ context.Context, _ int64, _ models.Repo, _ string) ([]string, error) {
	f.isCalled = true
	return f.failedPolicies, f.err
}
