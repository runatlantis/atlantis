package cache

import (
	"os/exec"

	"github.com/hashicorp/go-version"
)

func NewLocalBinaryCache(binaryName string) ExecutionVersionCache {
	return &localFS{
		binaryName: binaryName,
	}
}

// LocalFS is a basic implementation that just looks up
// the binary in the path. Primarily used for testing.
type localFS struct {
	binaryName string
}

func (m *localFS) Get(key *version.Version) (string, error) {
	return exec.LookPath(m.binaryName)
}
