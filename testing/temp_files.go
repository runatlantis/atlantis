package testing

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

// TempDir creates a temporary directory and returns its path along
// with a cleanup function to be called via defer, ex:
//   dir, cleanup := TempDir()
//   defer cleanup()
func TempDir(t *testing.T) (string, func()) {
	tmpDir, err := ioutil.TempDir("", "")
	Ok(t, err)
	return tmpDir, func() {
		os.RemoveAll(tmpDir) // nolint: errcheck
	}
}
func DirStructure(t *testing.T, structure map[string]interface{}) (string, func()) {
	tmpDir, cleanup := TempDir(t)
	dirStructureGo(t, tmpDir, structure)
	return tmpDir, cleanup
}

func dirStructureGo(t *testing.T, parentDir string, structure map[string]interface{}) {
	for key, val := range structure {
		// If val is nil then key is a filename and we just create it
		if val == nil {
			_, err := os.Create(filepath.Join(parentDir, key))
			Ok(t, err)
			continue
		}
		// If val is another map then key is a dir
		if dirContents, ok := val.(map[string]interface{}); ok {
			subDir := filepath.Join(parentDir, key)
			Ok(t, os.Mkdir(subDir, 0700))
			// Recurse and create contents.
			dirStructureGo(t, subDir, dirContents)
		}
	}
}
