package cache

import (
	"fmt"
	"sync"

	"github.com/hashicorp/go-version"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/runtime/models"
)

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_version_path.go ExecutionVersionCache
//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_key_serializer.go KeySerializer

type ExecutionVersionCache interface {
	Get(key *version.Version) (string, error)
}

type KeySerializer interface {
	Serialize(key *version.Version) (string, error)
}

type DefaultDiskLookupKeySerializer struct {
	binaryName string
}

func (s *DefaultDiskLookupKeySerializer) Serialize(key *version.Version) (string, error) {
	return fmt.Sprintf("%s%s", s.binaryName, key.Original()), nil
}

// ExecutionVersionDiskLayer is a cache layer which attempts to find the the version on disk,
// before calling the configured loading function.
type ExecutionVersionDiskLayer struct {
	versionRootDir models.FilePath
	exec           models.Exec
	keySerializer  KeySerializer
	loader         func(v *version.Version, destPath string) error
}

// Gets a path from cache
func (v *ExecutionVersionDiskLayer) Get(key *version.Version) (string, error) {
	binaryName, err := v.keySerializer.Serialize(key)

	if err != nil {
		return "", errors.Wrapf(err, "serializing key for disk lookup")
	}

	// first check for the binary in our path
	path, err := v.exec.LookPath(binaryName)

	if err == nil {
		return path, nil
	}

	// if the binary is not in our path, let's look in the version root directory
	binaryPath := v.versionRootDir.Join(binaryName)
	resolvedPath := binaryPath.Resolve()

	// if the binary doesn't exist there, we need to load it.
	if binaryPath.NotExists() {
		if err = v.loader(key, resolvedPath); err != nil {
			return "", errors.Wrapf(err, "loading %s", binaryPath)
		}
	}

	return resolvedPath, nil
}

// ExecutionVersionMemoryLayer is an in-memory cache which delegates to a disk layer
// if a version's path doesn't exist yet.
type ExecutionVersionMemoryLayer struct {
	// RWMutex allows us to have separation between reader locks/writer locks which is great
	// since writing of data shouldn't happen too often
	lock      sync.RWMutex
	diskLayer ExecutionVersionCache
	cache     map[string]string
}

func (v *ExecutionVersionMemoryLayer) Get(key *version.Version) (string, error) {

	// If we need to we can rip this out into a KeySerializer impl, for now this
	// seems overkill
	serializedKey := key.String()

	v.lock.RLock()
	_, ok := v.cache[serializedKey]
	v.lock.RUnlock()

	if !ok {
		v.lock.Lock()
		defer v.lock.Unlock()
		value, err := v.diskLayer.Get(key)

		if err != nil {
			return "", errors.Wrapf(err, "fetching %s from cache", serializedKey)
		}
		v.cache[serializedKey] = value
	}
	return v.cache[serializedKey], nil
}

func NewExecutionVersionLayeredLoadingCache(
	binaryName string,
	versionRootDir string,
	loader func(v *version.Version, destPath string) error,
) ExecutionVersionCache {

	diskLayer := &ExecutionVersionDiskLayer{
		exec:           models.LocalExec{},
		versionRootDir: models.LocalFilePath(versionRootDir),
		keySerializer:  &DefaultDiskLookupKeySerializer{binaryName: binaryName},
		loader:         loader,
	}

	return &ExecutionVersionMemoryLayer{
		diskLayer: diskLayer,
		cache:     make(map[string]string),
	}
}
