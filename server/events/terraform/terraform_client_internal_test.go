package terraform

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	version "github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

// Test that we write the file as expected
func TestGenerateRCFile_WritesFile(t *testing.T) {
	tmp, cleanup := TempDir(t)
	defer cleanup()

	err := generateRCFile("token", "hostname", tmp)
	Ok(t, err)

	expContents := `credentials "hostname" {
  token = "token"
}`
	actContents, err := ioutil.ReadFile(filepath.Join(tmp, ".terraformrc"))
	Ok(t, err)
	Equals(t, expContents, string(actContents))
}

// Test that if the file already exists and its contents will be modified if
// we write our config that we error out.
func TestGenerateRCFile_WillNotOverwrite(t *testing.T) {
	tmp, cleanup := TempDir(t)
	defer cleanup()

	rcFile := filepath.Join(tmp, ".terraformrc")
	err := ioutil.WriteFile(rcFile, []byte("contents"), 0600)
	Ok(t, err)

	actErr := generateRCFile("token", "hostname", tmp)
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
	err := ioutil.WriteFile(rcFile, []byte(contents), 0600)
	Ok(t, err)

	err = generateRCFile("token", "app.terraform.io", tmp)
	Ok(t, err)
}

// Test that if we can't read the existing file to see if the contents will be
// the same that we just error out.
func TestGenerateRCFile_ErrIfCannotRead(t *testing.T) {
	tmp, cleanup := TempDir(t)
	defer cleanup()

	rcFile := filepath.Join(tmp, ".terraformrc")
	err := ioutil.WriteFile(rcFile, []byte("can't see me!"), 0000)
	Ok(t, err)

	expErr := fmt.Sprintf("trying to read %s to ensure we're not overwriting it: open %s: permission denied", rcFile, rcFile)
	actErr := generateRCFile("token", "hostname", tmp)
	ErrEquals(t, expErr, actErr)
}

// Test that if we can't write, we error out.
func TestGenerateRCFile_ErrIfCannotWrite(t *testing.T) {
	rcFile := "/this/dir/does/not/exist/.terraformrc"
	expErr := fmt.Sprintf("writing generated .terraformrc file with TFE token to %s: open %s: no such file or directory", rcFile, rcFile)
	actErr := generateRCFile("token", "hostname", "/this/dir/does/not/exist")
	ErrEquals(t, expErr, actErr)
}

// Test that it executes with the expected env vars.
func TestDefaultClient_RunCommandWithVersion_EnvVars(t *testing.T) {
	v, err := version.NewVersion("0.11.11")
	Ok(t, err)
	tmp, cleanup := TempDir(t)
	defer cleanup()
	client := &DefaultClient{
		defaultVersion:          v,
		terraformPluginCacheDir: tmp,
		overrideTF:              "echo",
		usePluginCache:          true,
	}

	args := []string{
		"TF_IN_AUTOMATION=$TF_IN_AUTOMATION",
		"TF_PLUGIN_CACHE_DIR=$TF_PLUGIN_CACHE_DIR",
		"WORKSPACE=$WORKSPACE",
		"ATLANTIS_TERRAFORM_VERSION=$ATLANTIS_TERRAFORM_VERSION",
		"DIR=$DIR",
	}
	// If this runs in Jenkins WORKSPACE is set to the jenkins workspace so this can't be set to a fix value
	CurrentWorkspace := os.Getenv("WORKSPACE")
	if CurrentWorkspace == "" {
		CurrentWorkspace = "workspace"
	}
	out, err := client.RunCommandWithVersion(nil, tmp, args, map[string]string{}, nil, CurrentWorkspace)
	Ok(t, err)

	exp := fmt.Sprintf("TF_IN_AUTOMATION=true TF_PLUGIN_CACHE_DIR=%s WORKSPACE=%s ATLANTIS_TERRAFORM_VERSION=0.11.11 DIR=%s\n", tmp, CurrentWorkspace, tmp)
	Equals(t, exp, out)
}

// Test that it returns an error on error.
func TestDefaultClient_RunCommandWithVersion_Error(t *testing.T) {
	v, err := version.NewVersion("0.11.11")
	Ok(t, err)
	tmp, cleanup := TempDir(t)
	defer cleanup()
	client := &DefaultClient{
		defaultVersion:          v,
		terraformPluginCacheDir: tmp,
		overrideTF:              "echo",
	}

	args := []string{
		"dying",
		"&&",
		"exit",
		"1",
	}
	log := logging.NewSimpleLogger("test", false, logging.Debug)
	out, err := client.RunCommandWithVersion(log, tmp, args, map[string]string{}, nil, "workspace")
	ErrEquals(t, fmt.Sprintf(`running "echo dying && exit 1" in %q: exit status 1`, tmp), err)
	// Test that we still get our output.
	Equals(t, "dying\n", out)
}

func TestDefaultClient_RunCommandAsync_Success(t *testing.T) {
	v, err := version.NewVersion("0.11.11")
	Ok(t, err)
	tmp, cleanup := TempDir(t)
	defer cleanup()
	client := &DefaultClient{
		defaultVersion:          v,
		terraformPluginCacheDir: tmp,
		overrideTF:              "echo",
		usePluginCache:          true,
	}

	args := []string{
		"TF_IN_AUTOMATION=$TF_IN_AUTOMATION",
		"TF_PLUGIN_CACHE_DIR=$TF_PLUGIN_CACHE_DIR",
		"WORKSPACE=$WORKSPACE",
		"ATLANTIS_TERRAFORM_VERSION=$ATLANTIS_TERRAFORM_VERSION",
		"DIR=$DIR",
	}
	// If this runs in Jenkins WORKSPACE is set to the jenkins workspace so this can't be set to a fix value
	CurrentWorkspace := os.Getenv("WORKSPACE")
	if CurrentWorkspace == "" {
		CurrentWorkspace = "workspace"
	}
	_, outCh := client.RunCommandAsync(nil, tmp, args, map[string]string{}, nil, CurrentWorkspace)

	out, err := waitCh(outCh)
	Ok(t, err)
	exp := fmt.Sprintf("TF_IN_AUTOMATION=true TF_PLUGIN_CACHE_DIR=%s WORKSPACE=%s ATLANTIS_TERRAFORM_VERSION=0.11.11 DIR=%s", tmp, CurrentWorkspace, tmp)
	Equals(t, exp, out)
}

func TestDefaultClient_RunCommandAsync_BigOutput(t *testing.T) {
	v, err := version.NewVersion("0.11.11")
	Ok(t, err)
	tmp, cleanup := TempDir(t)
	defer cleanup()
	client := &DefaultClient{
		defaultVersion:          v,
		terraformPluginCacheDir: tmp,
		overrideTF:              "cat",
	}
	filename := filepath.Join(tmp, "data")
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	Ok(t, err)

	var exp string
	for i := 0; i < 1024; i++ {
		s := strings.Repeat("0", 10) + "\n"
		exp += s
		_, err = f.WriteString(s)
		Ok(t, err)
	}
	_, outCh := client.RunCommandAsync(nil, tmp, []string{filename}, map[string]string{}, nil, "workspace")

	out, err := waitCh(outCh)
	Ok(t, err)
	Equals(t, strings.TrimRight(exp, "\n"), out)
}

func TestDefaultClient_RunCommandAsync_StderrOutput(t *testing.T) {
	v, err := version.NewVersion("0.11.11")
	Ok(t, err)
	tmp, cleanup := TempDir(t)
	defer cleanup()
	client := &DefaultClient{
		defaultVersion:          v,
		terraformPluginCacheDir: tmp,
		overrideTF:              "echo",
	}
	log := logging.NewSimpleLogger("test", false, logging.Debug)
	_, outCh := client.RunCommandAsync(log, tmp, []string{"stderr", ">&2"}, map[string]string{}, nil, "workspace")

	out, err := waitCh(outCh)
	Ok(t, err)
	Equals(t, "stderr", out)
}

func TestDefaultClient_RunCommandAsync_ExitOne(t *testing.T) {
	v, err := version.NewVersion("0.11.11")
	Ok(t, err)
	tmp, cleanup := TempDir(t)
	defer cleanup()
	client := &DefaultClient{
		defaultVersion:          v,
		terraformPluginCacheDir: tmp,
		overrideTF:              "echo",
	}
	log := logging.NewSimpleLogger("test", false, logging.Debug)

	_, outCh := client.RunCommandAsync(log, tmp, []string{"dying", "&&", "exit", "1"}, map[string]string{}, nil, "workspace")

	out, err := waitCh(outCh)
	ErrEquals(t, fmt.Sprintf(`running "echo dying && exit 1" in %q: exit status 1`, tmp), err)
	// Test that we still get our output.
	Equals(t, "dying", out)
}

func TestDefaultClient_RunCommandAsync_Input(t *testing.T) {
	v, err := version.NewVersion("0.11.11")
	Ok(t, err)
	tmp, cleanup := TempDir(t)
	defer cleanup()
	client := &DefaultClient{
		defaultVersion:          v,
		terraformPluginCacheDir: tmp,
		overrideTF:              "read",
	}
	log := logging.NewSimpleLogger("test", false, logging.Debug)
	inCh, outCh := client.RunCommandAsync(log, tmp, []string{"a", "&&", "echo", "$a"}, map[string]string{}, nil, "workspace")
	inCh <- "echo me\n"

	out, err := waitCh(outCh)
	Ok(t, err)
	Equals(t, "echo me", out)
}

func waitCh(ch <-chan Line) (string, error) {
	var ls []string
	for line := range ch {
		if line.Err != nil {
			return strings.Join(ls, "\n"), line.Err
		}
		ls = append(ls, line.Line)
	}
	return strings.Join(ls, "\n"), nil
}
