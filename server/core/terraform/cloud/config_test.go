package cloud

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	. "github.com/runatlantis/atlantis/testing"
)

// Test that we write the file as expected
func TestGenerateRCFile_WritesFile(t *testing.T) {
	tmp, cleanup := TempDir(t)
	defer cleanup()

	err := GenerateConfigFile("token", "hostname", tmp)
	Ok(t, err)

	expContents := `credentials "hostname" {
  token = "token"
}`
	actContents, err := os.ReadFile(filepath.Join(tmp, ".terraformrc"))
	Ok(t, err)
	Equals(t, expContents, string(actContents))
}

// Test that if the file already exists and its contents will be modified if
// we write our config that we error out.
func TestGenerateRCFile_WillNotOverwrite(t *testing.T) {
	tmp, cleanup := TempDir(t)
	defer cleanup()

	rcFile := filepath.Join(tmp, ".terraformrc")
	err := os.WriteFile(rcFile, []byte("contents"), 0600)
	Ok(t, err)

	actErr := GenerateConfigFile("token", "hostname", tmp)
	expErr := fmt.Sprintf("can't write TFE token to %s because that file has contents that would be overwritten", tmp+"/.terraformrc")
	ErrEquals(t, expErr, actErr)
}

// Test that if the file already exists and its contents will NOT be modified if
// we write our config that we don't error.
func TestGenerateRCFile_NoErrIfContentsSame(t *testing.T) {
	tmp, cleanup := TempDir(t)
	defer cleanup()

	rcFile := filepath.Join(tmp, ".terraformrc")
	contents := `credentials "app.terraform.io" {
  token = "token"
}`
	err := os.WriteFile(rcFile, []byte(contents), 0600)
	Ok(t, err)

	err = GenerateConfigFile("token", "app.terraform.io", tmp)
	Ok(t, err)
}

// Test that if we can't read the existing file to see if the contents will be
// the same that we just error out.
func TestGenerateRCFile_ErrIfCannotRead(t *testing.T) {
	tmp, cleanup := TempDir(t)
	defer cleanup()

	rcFile := filepath.Join(tmp, ".terraformrc")
	err := os.WriteFile(rcFile, []byte("can't see me!"), 0000)
	Ok(t, err)

	expErr := fmt.Sprintf("trying to read %s to ensure we're not overwriting it: open %s: permission denied", rcFile, rcFile)
	actErr := GenerateConfigFile("token", "hostname", tmp)
	ErrEquals(t, expErr, actErr)
}

// Test that if we can't write, we error out.
func TestGenerateRCFile_ErrIfCannotWrite(t *testing.T) {
	rcFile := "/this/dir/does/not/exist/.terraformrc"
	expErr := fmt.Sprintf("writing generated .terraformrc file with TFE token to %s: open %s: no such file or directory", rcFile, rcFile)
	actErr := GenerateConfigFile("token", "hostname", "/this/dir/does/not/exist")
	ErrEquals(t, expErr, actErr)
}
