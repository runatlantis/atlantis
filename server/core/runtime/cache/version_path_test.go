package cache

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/hashicorp/go-version"
	"go.uber.org/mock/gomock"
	cache_mocks "github.com/runatlantis/atlantis/server/core/runtime/cache/mocks"
	"github.com/runatlantis/atlantis/server/core/runtime/models"
	models_mocks "github.com/runatlantis/atlantis/server/core/runtime/models/mocks"
	. "github.com/runatlantis/atlantis/testing"
)

func TestExecutionVersionDiskLayer(t *testing.T) {

	binaryVersion := "bin1.0"
	binaryName := "bin"

	expectedPath := "some/path/bin1.0"
	versionInput, _ := version.NewVersion("1.0")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFilePath := models_mocks.NewMockFilePath(ctrl)
	mockExec := models_mocks.NewMockExec(ctrl)
	mockSerializer := cache_mocks.NewMockKeySerializer(ctrl)

	t.Run("serializer error", func(t *testing.T) {
		subject := &ExecutionVersionDiskLayer{
			versionRootDir: mockFilePath,
			exec:           mockExec,
			loader: func(v *version.Version, destPath string) (models.FilePath, error) {
				if destPath == expectedPath && v == versionInput {
					return models.LocalFilePath(filepath.Join(destPath, "bin")), nil
				}

				t.Fatalf("unexpected inputs to loader")

				return models.LocalFilePath(""), nil
			},
			keySerializer: mockSerializer,
		}

		mockSerializer.EXPECT().Serialize(versionInput).Return("", errors.New("serializer error"))
		mockExec.EXPECT().LookPath(binaryVersion).Return(expectedPath, nil)

		_, err := subject.Get(versionInput)

		Assert(t, err != nil, "err is expected")

		// TODO: Convert Never() expectation: mockFilePath.EXPECT().Join(gomock.Any().Times(0))
		// TODO: Convert Never() expectation: mockFilePath.EXPECT().NotExists().Times(0)
		// TODO: Convert Never() expectation: mockFilePath.EXPECT().Resolve().Times(0)
		// TODO: Convert Never() expectation: mockExec.EXPECT().LookPath(gomock.Any().Times(0))
	})

	t.Run("finds in path", func(t *testing.T) {
		subject := &ExecutionVersionDiskLayer{
			versionRootDir: mockFilePath,
			exec:           mockExec,
			loader: func(v *version.Version, destPath string) (models.FilePath, error) {
				t.Fatalf("shouldn't be called")

				return models.LocalFilePath(""), nil
			},
			keySerializer: mockSerializer,
		}

		mockSerializer.EXPECT().Serialize(versionInput).Return(binaryVersion, nil)
		mockExec.EXPECT().LookPath(binaryVersion).Return(expectedPath, nil)

		resultPath, err := subject.Get(versionInput)

		Ok(t, err)

		Assert(t, resultPath == expectedPath, "path is expected")

		// TODO: Convert Never() expectation: mockFilePath.EXPECT().Join(gomock.Any().Times(0))
		// TODO: Convert Never() expectation: mockFilePath.EXPECT().Resolve().Times(0)
		// TODO: Convert Never() expectation: mockFilePath.EXPECT().NotExists().Times(0)
	})

	t.Run("finds in version root", func(t *testing.T) {
		subject := &ExecutionVersionDiskLayer{
			versionRootDir: mockFilePath,
			exec:           mockExec,
			loader: func(v *version.Version, destPath string) (models.FilePath, error) {

				t.Fatalf("shouldn't be called")

				return models.LocalFilePath(""), nil
			},
			keySerializer: mockSerializer,
		}

		mockSerializer.EXPECT().Serialize(versionInput).Return(binaryVersion, nil)
		mockExec.EXPECT().LookPath(binaryVersion).Return("", errors.New("error"))

		mockFilePath.EXPECT().Join(binaryVersion).Return(mockFilePath)

		mockFilePath.EXPECT().NotExists().Return(false)
		mockFilePath.EXPECT().Resolve().Return(expectedPath)

		resultPath, err := subject.Get(versionInput)

		Ok(t, err)

		Assert(t, resultPath == expectedPath, "path is expected")
	})

	t.Run("loads version", func(t *testing.T) {
		mockLoaderPath := models_mocks.NewMockFilePath(ctrl)
		mockSymlinkPath := models_mocks.NewMockFilePath(ctrl)
		mockLoadedBinaryPath := models_mocks.NewMockFilePath(ctrl)
		expectedLoaderPath := "/some/path/to/binary"
		expectedBinaryVersionPath := filepath.Join(expectedPath, binaryVersion)

		subject := &ExecutionVersionDiskLayer{
			versionRootDir: mockFilePath,
			exec:           mockExec,
			loader: func(v *version.Version, destPath string) (models.FilePath, error) {

				if destPath == expectedLoaderPath && v == versionInput {
					return mockLoadedBinaryPath, nil
				}

				t.Fatalf("unexpected inputs to loader")

				return models.LocalFilePath(""), nil
			},
			binaryName:    binaryName,
			keySerializer: mockSerializer,
		}

		mockSerializer.EXPECT().Serialize(versionInput).Return(binaryVersion, nil)
		mockExec.EXPECT().LookPath(binaryVersion).Return("", errors.New("error"))

		mockFilePath.EXPECT().Join(binaryVersion).Return(mockFilePath)
		mockFilePath.EXPECT().Resolve().Return(expectedBinaryVersionPath)

		mockFilePath.EXPECT().NotExists().Return(true)

		mockFilePath.EXPECT().Join(binaryName, "versions", versionInput.Original()).Return(mockLoaderPath)

		mockLoaderPath.EXPECT().Resolve().Return(expectedLoaderPath)
		mockLoadedBinaryPath.EXPECT().Symlink(expectedBinaryVersionPath).Return(mockSymlinkPath, nil)

		mockSymlinkPath.EXPECT().Resolve().Return(expectedPath)

		resultPath, err := subject.Get(versionInput)

		Ok(t, err)

		Assert(t, resultPath == expectedPath, "path is expected")
	})

	t.Run("loader error", func(t *testing.T) {
		mockLoaderPath := models_mocks.NewMockFilePath(ctrl)
		expectedLoaderPath := "/some/path/to/binary"
		subject := &ExecutionVersionDiskLayer{
			versionRootDir: mockFilePath,
			exec:           mockExec,
			loader: func(v *version.Version, destPath string) (models.FilePath, error) {

				if destPath == expectedLoaderPath && v == versionInput {
					return models.LocalFilePath(""), errors.New("error")
				}

				t.Fatalf("unexpected inputs to loader")

				return models.LocalFilePath(""), nil
			},
			keySerializer: mockSerializer,
			binaryName:    binaryName,
		}

		mockSerializer.EXPECT().Serialize(versionInput).Return(binaryVersion, nil)
		mockExec.EXPECT().LookPath(binaryVersion).Return("", errors.New("error"))

		mockFilePath.EXPECT().Join(binaryVersion).Return(mockFilePath)

		mockFilePath.EXPECT().NotExists().Return(true)

		mockFilePath.EXPECT().Join(binaryName, "versions", versionInput.Original()).Return(mockLoaderPath)

		mockLoaderPath.EXPECT().Resolve().Return(expectedLoaderPath)

		_, err := subject.Get(versionInput)

		Assert(t, err != nil, "path is expected")
	})
}

func TestExecutionVersionMemoryLayer(t *testing.T) {
	expectedPath := "some/path"
	versionInput, _ := version.NewVersion("1.0")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLayer := cache_mocks.NewMockExecutionVersionCache(ctrl)

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

		mockLayer.EXPECT().Get(versionInput).Return("", errors.New("error"))

		_, err := subject.Get(versionInput)

		Assert(t, err != nil, "error is expected")
	})

	t.Run("disk layer success", func(t *testing.T) {
		delete(cache, versionInput.String())

		mockLayer.EXPECT().Get(versionInput).Return(expectedPath, nil)

		resultPath, err := subject.Get(versionInput)

		Ok(t, err)

		Assert(t, resultPath == expectedPath, "path is expected")
		Assert(t, cache[versionInput.String()] == resultPath, "path is cached")
	})
}
