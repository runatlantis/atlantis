package jobs_test

import (
	"fmt"
	"testing"

	"github.com/uber-go/tally/v4"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/jobs"
	"github.com/runatlantis/atlantis/server/jobs/mocks"
	"github.com/runatlantis/atlantis/server/jobs/mocks/matchers"
	"github.com/stretchr/testify/assert"

	. "github.com/petergtz/pegomock"
	. "github.com/runatlantis/atlantis/testing"
)

func TestJobStore_Get(t *testing.T) {
	t.Run("load from memory", func(t *testing.T) {
		// Setup job store
		storageBackend := mocks.NewMockStorageBackend()
		expectedJob := &jobs.Job{
			Output: []string{"a"},
			Status: jobs.Complete,
		}
		jobsMap := make(map[string]*jobs.Job)
		jobsMap["1234"] = expectedJob
		jobStore := jobs.NewTestJobStore(storageBackend, jobsMap)

		// Assert job
		gotJob, err := jobStore.Get("1234")
		assert.NoError(t, err)
		assert.Equal(t, expectedJob.Output, gotJob.Output)
		assert.Equal(t, expectedJob.Status, gotJob.Status)
	})

	t.Run("load from storage backend when not in memory", func(t *testing.T) {
		// Setup job store
		storageBackend := mocks.NewMockStorageBackend()
		expectedLogs := []string{"a", "b"}
		expectedJob := jobs.Job{
			Output: expectedLogs,
			Status: jobs.Complete,
		}
		When(storageBackend.Read(AnyString())).ThenReturn(expectedLogs, nil)

		// Assert job
		jobStore := jobs.NewJobStore(storageBackend, tally.NewTestScope("test", map[string]string{}))
		gotJob, err := jobStore.Get("1234")
		assert.NoError(t, err)
		assert.Equal(t, expectedJob.Output, gotJob.Output)
		assert.Equal(t, expectedJob.Status, gotJob.Status)
	})

	t.Run("error when reading from storage backend fails", func(t *testing.T) {
		// Setup job store
		storageBackend := mocks.NewMockStorageBackend()
		expectedError := fmt.Errorf("reading from backend storage: error")
		When(storageBackend.Read(AnyString())).ThenReturn([]string{}, errors.New("error"))

		// Assert job
		jobStore := jobs.NewJobStore(storageBackend, tally.NewTestScope("test", map[string]string{}))
		gotJob, err := jobStore.Get("1234")
		assert.Empty(t, gotJob)
		assert.EqualError(t, expectedError, err.Error())
	})
}

func TestJobStore_AppendOutput(t *testing.T) {

	t.Run("append output when new job", func(t *testing.T) {
		// Setup job store
		storageBackend := mocks.NewMockStorageBackend()
		jobStore := jobs.NewJobStore(storageBackend, tally.NewTestScope("test", map[string]string{}))
		jobID := "1234"
		output := "Test log message"

		err := jobStore.AppendOutput(jobID, output)
		assert.NoError(t, err)

		// Assert job
		job, err := jobStore.Get(jobID)
		Ok(t, err)
		assert.Equal(t, job.Output, []string{output})
		assert.Equal(t, job.Status, jobs.Processing)
	})

	t.Run("append output when existing job", func(t *testing.T) {
		// Setup job store
		storageBackend := mocks.NewMockStorageBackend()
		jobStore := jobs.NewJobStore(storageBackend, tally.NewTestScope("test", map[string]string{}))
		jobID := "1234"
		output := []string{"Test log message", "Test log message 2"}

		err := jobStore.AppendOutput(jobID, output[0])
		assert.NoError(t, err)

		err = jobStore.AppendOutput(jobID, output[1])
		assert.NoError(t, err)

		// Assert job
		job, err := jobStore.Get(jobID)
		Ok(t, err)
		assert.Equal(t, job.Output, output)
		assert.Equal(t, job.Status, jobs.Processing)
	})

	t.Run("error when job status complete", func(t *testing.T) {
		// Setup job store
		storageBackend := mocks.NewMockStorageBackend()
		jobID := "1234"
		job := &jobs.Job{
			Output: []string{"a"},
			Status: jobs.Complete,
		}

		// Add complete to job in store
		jobsMap := make(map[string]*jobs.Job)
		jobsMap[jobID] = job
		jobStore := jobs.NewTestJobStore(storageBackend, jobsMap)

		// Assert error
		err := jobStore.AppendOutput(jobID, "test message")
		assert.Error(t, err)
	})
}

func TestJobStore_UpdateJobStatus(t *testing.T) {

	t.Run("retain job in memory when persist fails", func(t *testing.T) {
		// Create new job and add it to store
		jobID := "1234"
		job := &jobs.Job{
			Output: []string{"a"},
			Status: jobs.Processing,
		}
		jobsMap := make(map[string]*jobs.Job)
		jobsMap[jobID] = job
		storageBackendErr := fmt.Errorf("random error")
		expecterErr := errors.Wrapf(storageBackendErr, "persisting job: %s", jobID)

		// Setup storage backend
		storageBackend := mocks.NewMockStorageBackend()
		When(storageBackend.Write(AnyString(), matchers.AnySliceOfString(), AnyString())).ThenReturn(false, storageBackendErr)
		jobStore := jobs.NewTestJobStore(storageBackend, jobsMap)
		err := jobStore.SetJobCompleteStatus(jobID, "test-repo", jobs.Complete)

		// Assert storage backend error
		assert.EqualError(t, err, expecterErr.Error())

		// Assert the job is in memory
		jobInMem, err := jobStore.Get(jobID)
		Ok(t, err)
		assert.Equal(t, jobInMem.Output, job.Output)
		assert.Equal(t, job.Status, jobs.Complete)
	})

	t.Run("retain job in memory when storage backend not configured", func(t *testing.T) {
		// Create new job and add it to store
		jobID := "1234"
		job := &jobs.Job{
			Output: []string{"a"},
			Status: jobs.Processing,
		}
		jobsMap := make(map[string]*jobs.Job)
		jobsMap[jobID] = job

		// Setup storage backend
		storageBackend := &jobs.NoopStorageBackend{}
		jobStore := jobs.NewTestJobStore(storageBackend, jobsMap)
		err := jobStore.SetJobCompleteStatus(jobID, "test-repo", jobs.Complete)

		assert.Nil(t, err)

		// Assert the job is in memory
		jobInMem, err := jobStore.Get(jobID)
		Ok(t, err)
		assert.Equal(t, jobInMem.Output, job.Output)
		assert.Equal(t, job.Status, jobs.Complete)
	})

	t.Run("delete from memory when persist succeeds", func(t *testing.T) {
		// Create new job and add it to store
		jobID := "1234"
		job := &jobs.Job{
			Output: []string{"a"},
			Status: jobs.Processing,
		}
		jobsMap := make(map[string]*jobs.Job)
		jobsMap[jobID] = job

		// Setup storage backend
		storageBackend := mocks.NewMockStorageBackend()
		When(storageBackend.Write(AnyString(), matchers.AnySliceOfString(), AnyString())).ThenReturn(true, nil)
		jobStore := jobs.NewTestJobStore(storageBackend, jobsMap)
		err := jobStore.SetJobCompleteStatus(jobID, "test-repo", jobs.Complete)
		assert.Nil(t, err)

		When(storageBackend.Read(jobID)).ThenReturn([]string{}, nil)
		gotJob, err := jobStore.Get(jobID)
		assert.Nil(t, err)
		assert.Empty(t, gotJob.Output)
	})

	t.Run("error when job does not exist", func(t *testing.T) {
		storageBackend := mocks.NewMockStorageBackend()
		jobStore := jobs.NewJobStore(storageBackend, tally.NewTestScope("test", map[string]string{}))
		jobID := "1234"
		expectedErrString := fmt.Sprintf("job: %s does not exist", jobID)

		err := jobStore.SetJobCompleteStatus(jobID, "test-repo", jobs.Complete)
		assert.EqualError(t, err, expectedErrString)

	})
}
