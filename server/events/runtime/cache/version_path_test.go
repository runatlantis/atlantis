package cache

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/hashicorp/go-version"
	. "github.com/petergtz/pegomock"
	cache_mocks "github.com/runatlantis/atlantis/server/events/runtime/cache/mocks"
	models_mocks "github.com/runatlantis/atlantis/server/events/runtime/models/mocks"
	. "github.com/runatlantis/atlantis/testing"
)

func TestExecutionVersionDiskLayer(t *testing.T) {

	binaryVersion := "bin1.0"
	binaryName := "bin"

	expectedPath := "some/path"
	versionInput, _ := version.NewVersion("1.0")

	RegisterMockTestingT(t)

	mockFilePath := models_mocks.NewMockFilePath()
	mockExec := models_mocks.NewMockExec()
	mockSerializer := cache_mocks.NewMockKeySerializer()

	t.Run("serializer error", func(t *testing.T) {
		subject := &ExecutionVersionDiskLayer{
			versionRootDir: mockFilePath,
			exec:           mockExec,
			loader: func(v *version.Version, destPath string) (string, error) {
				if destPath == expectedPath && v == versionInput {
					return filepath.Join(destPath, "bin"), nil
				}

				t.Fatalf("unexpected inputs to loader")

				return "", nil
			},
			keySerializer: mockSerializer,
		}

		When(mockSerializer.Serialize(versionInput)).ThenReturn("", errors.New("serializer error"))
		When(mockExec.LookPath(binaryVersion)).ThenReturn(expectedPath, nil)

		_, err := subject.Get(versionInput)

		Assert(t, err != nil, "err is expected")

		mockFilePath.VerifyWasCalled(Never()).Join(AnyString())
		mockFilePath.VerifyWasCalled(Never()).NotExists()
		mockFilePath.VerifyWasCalled(Never()).Resolve()
		mockExec.VerifyWasCalled(Never()).LookPath(AnyString())
	})

	t.Run("finds in path", func(t *testing.T) {
		subject := &ExecutionVersionDiskLayer{
			versionRootDir: mockFilePath,
			exec:           mockExec,
			loader: func(v *version.Version, destPath string) (string, error) {
				t.Fatalf("shouldn't be called")

				return "", nil
			},
			keySerializer: mockSerializer,
		}

		When(mockSerializer.Serialize(versionInput)).ThenReturn(binaryVersion, nil)
		When(mockExec.LookPath(binaryVersion)).ThenReturn(expectedPath, nil)

		resultPath, err := subject.Get(versionInput)

		Ok(t, err)

		Assert(t, resultPath == expectedPath, "path is expected")

		mockFilePath.VerifyWasCalled(Never()).Join(AnyString())
		mockFilePath.VerifyWasCalled(Never()).Resolve()
		mockFilePath.VerifyWasCalled(Never()).NotExists()
	})

	t.Run("finds in version root", func(t *testing.T) {
		subject := &ExecutionVersionDiskLayer{
			versionRootDir: mockFilePath,
			exec:           mockExec,
			loader: func(v *version.Version, destPath string) (string, error) {

				t.Fatalf("shouldn't be called")

				return "", nil
			},
			keySerializer: mockSerializer,
		}

		When(mockSerializer.Serialize(versionInput)).ThenReturn(binaryVersion, nil)
		When(mockExec.LookPath(binaryVersion)).ThenReturn("", errors.New("error"))

		When(mockFilePath.Join(binaryVersion)).ThenReturn(mockFilePath)

		When(mockFilePath.NotExists()).ThenReturn(false)
		When(mockFilePath.Resolve()).ThenReturn(expectedPath)

		resultPath, err := subject.Get(versionInput)

		Ok(t, err)

		Assert(t, resultPath == expectedPath, "path is expected")
	})

	t.Run("loads version", func(t *testing.T) {
		mockLoaderPath := models_mocks.NewMockFilePath()
		mockSymlinkPath := models_mocks.NewMockFilePath()
		expectedLoaderPath := "/some/path/to/binary"

		subject := &ExecutionVersionDiskLayer{
			versionRootDir: mockFilePath,
			exec:           mockExec,
			loader: func(v *version.Version, destPath string) (string, error) {

				if destPath == expectedLoaderPath && v == versionInput {
					return filepath.Join(destPath, "bin"), nil
				}

				t.Fatalf("unexpected inputs to loader")

				return "", nil
			},
			binaryName:    binaryName,
			keySerializer: mockSerializer,
		}

		When(mockSerializer.Serialize(versionInput)).ThenReturn(binaryVersion, nil)
		When(mockExec.LookPath(binaryVersion)).ThenReturn("", errors.New("error"))

		When(mockFilePath.Join(binaryVersion)).ThenReturn(mockFilePath)

		When(mockFilePath.NotExists()).ThenReturn(true)

		When(mockFilePath.Join(binaryName, "versions", versionInput.Original())).ThenReturn(mockLoaderPath)

		When(mockLoaderPath.Resolve()).ThenReturn(expectedLoaderPath)
		When(mockLoaderPath.Symlink(filepath.Join(expectedLoaderPath, "bin"))).ThenReturn(mockSymlinkPath, nil)

		When(mockSymlinkPath.Resolve()).ThenReturn(expectedPath)

		resultPath, err := subject.Get(versionInput)

		Ok(t, err)

		Assert(t, resultPath == expectedPath, "path is expected")
	})

	t.Run("loader error", func(t *testing.T) {
		mockLoaderPath := models_mocks.NewMockFilePath()
		expectedLoaderPath := "/some/path/to/binary"
		subject := &ExecutionVersionDiskLayer{
			versionRootDir: mockFilePath,
			exec:           mockExec,
			loader: func(v *version.Version, destPath string) (string, error) {

				if destPath == expectedLoaderPath && v == versionInput {
					return "", errors.New("error")
				}

				t.Fatalf("unexpected inputs to loader")

				return "", nil
			},
			keySerializer: mockSerializer,
			binaryName:    binaryName,
		}

		When(mockSerializer.Serialize(versionInput)).ThenReturn(binaryVersion, nil)
		When(mockExec.LookPath(binaryVersion)).ThenReturn("", errors.New("error"))

		When(mockFilePath.Join(binaryVersion)).ThenReturn(mockFilePath)

		When(mockFilePath.NotExists()).ThenReturn(true)

		When(mockFilePath.Join(binaryName, "versions", versionInput.Original())).ThenReturn(mockLoaderPath)

		When(mockLoaderPath.Resolve()).ThenReturn(expectedLoaderPath)

		_, err := subject.Get(versionInput)

		Assert(t, err != nil, "path is expected")
	})
}

func TestExecutionVersionMemoryLayer(t *testing.T) {
	expectedPath := "some/path"
	versionInput, _ := version.NewVersion("1.0")

	RegisterMockTestingT(t)

	mockLayer := cache_mocks.NewMockExecutionVersionCache()

	cache := make(map[string]string)

	subject := &ExecutionVersionMemoryLayer{
		diskLayer: mockLayer,
		cache:     cache,
	}

	t.Run("exists in cache", func(t *testing.T) {
		cache[versionInput.String()] = expectedPath

		resultPath, err := subject.Get(versionInput)

		Ok(t, err)

		Assert(t, resultPath == expectedPath, "path is expected")
	})

	t.Run("disk layer error", func(t *testing.T) {
		delete(cache, versionInput.String())

		When(mockLayer.Get(versionInput)).ThenReturn("", errors.New("error"))

		_, err := subject.Get(versionInput)

		Assert(t, err != nil, "error is expected")
	})

	t.Run("disk layer success", func(t *testing.T) {
		delete(cache, versionInput.String())

		When(mockLayer.Get(versionInput)).ThenReturn(expectedPath, nil)

		resultPath, err := subject.Get(versionInput)

		Ok(t, err)

		Assert(t, resultPath == expectedPath, "path is expected")
		Assert(t, cache[versionInput.String()] == resultPath, "path is cached")
	})
}
