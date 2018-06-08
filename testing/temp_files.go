package testing

import (
	"io/ioutil"
	"os"
	"testing"
)

// TempDir creates a temporary directory and returns its path along
// with a cleanup function to be called via defer, ex:
//   dir, cleanup := TempDir()
//   defer cleanup()
func TempDir(t *testing.T) (string, func()) {
	tmpDir, err := ioutil.TempDir("", "")
	Ok(t, err)
	return tmpDir, func() { os.RemoveAll(tmpDir) }
}
