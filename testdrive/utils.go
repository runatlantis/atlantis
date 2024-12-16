// Copyright 2017 HootSuite Media Inc.
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// Modified hereafter by contributors to runatlantis/atlantis.

package testdrive

import (
	"archive/zip"
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/term"
)

const hashicorpReleasesURL = "https://releases.hashicorp.com"
const terraformVersion = "1.10.1" // renovate: datasource=github-releases depName=hashicorp/terraform versioning=hashicorp
const ngrokDownloadURL = "https://bin.equinox.io/c/4VmDzA7iaHb"
const ngrokAPIURL = "localhost:41414" // We hope this isn't used.
const atlantisPort = 4141

func readPassword() (string, error) {
	password, err := term.ReadPassword(int(syscall.Stdin)) // nolint: unconvert
	return string(password), err
}

func downloadFile(url string, path string) error {
	output, err := os.Create(path)
	if err != nil {
		return err
	}
	defer output.Close() // nolint: errcheck

	response, err := http.Get(url) // nolint: gosec
	if err != nil {
		return err
	}
	defer response.Body.Close() // nolint: errcheck

	_, err = io.Copy(output, response.Body)
	return err
}

// This function is used to sanitize the file path to avoid a "zip slip" attack
// source: https://github.com/securego/gosec/issues/324#issuecomment-935927967
func sanitizeArchivePath(d, t string) (v string, err error) {
	v = filepath.Join(d, t)
	if strings.HasPrefix(v, filepath.Clean(d)) {
		return v, nil
	}

	return "", fmt.Errorf("%s: %s", "content filepath is tainted", t)
}

func unzip(archive, target string) error {
	reader, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}

	for _, file := range reader.File {
		path, err := sanitizeArchivePath(target, file.Name)
		if err != nil {
			return err
		}

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(path, file.Mode()); err != nil {
				return err
			}
			continue
		}

		fileReader, err := file.Open()
		if err != nil {
			return err
		}
		defer fileReader.Close() // nolint: errcheck

		targetFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}
		defer targetFile.Close() // nolint: errcheck

		for {
			_, err := io.CopyN(targetFile, fileReader, 1024)
			if err != nil {
				if err == io.EOF {
					break
				}
				return err
			}
		}
	}

	return nil
}

func getTunnelAddr() (string, error) {
	tunAPI := fmt.Sprintf("http://%s/api/tunnels", ngrokAPIURL)
	response, err := http.Get(tunAPI) // nolint: gosec
	if err != nil {
		return "", err
	}
	defer response.Body.Close() // nolint: errcheck

	type tunnels struct {
		Tunnels []struct {
			PublicURL string `json:"public_url"`
			Proto     string `json:"proto"`
			Config    struct {
				Addr string `json:"addr"`
			} `json:"config"`
		} `json:"tunnels"`
	}

	var t tunnels

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", errors.Wrap(err, "reading ngrok api")
	}
	if err = json.Unmarshal(body, &t); err != nil {
		return "", errors.Wrapf(err, "parsing ngrok api: %s", string(body))
	}

	// Find the tunnel we just created.
	expAtlantisURL := fmt.Sprintf("http://localhost:%d", atlantisPort)
	for _, tun := range t.Tunnels {
		if tun.Proto == "https" && tun.Config.Addr == expAtlantisURL {
			return tun.PublicURL, nil
		}
	}

	return "", fmt.Errorf("did not find ngrok tunnel with proto 'https' and config.addr '%s' in list of tunnels at %s\n%s", expAtlantisURL, tunAPI, string(body))
}

func downloadAndUnzip(url string, path string, target string) error {
	if err := downloadFile(url, path); err != nil {
		return err
	}
	return unzip(path, target)
}

// executeCmd executes a command, waits for it to finish and returns any errors.
func executeCmd(cmd string, args ...string) error {
	command := exec.Command(cmd, args...) // #nosec
	bytes, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %s", err, bytes)
	}
	return nil
}

// execAndWaitForStderr executes a command with name and args. It waits until
// timeout for the stderr output of the command to match stderrMatch. If the
// timeout comes first, then it cancels the command and returns the error as
// error (not on the channel). Otherwise the function returns and the command
// continues to run in the background. Any errors after this point are passed
// onto the error channel and the command is stopped. We increment the wg
// so that callers can wait until command is killed before exiting.
// The cancelFunc can be used to stop the command but callers should still wait
// for the wg to be Done to ensure the command completes its cancellation
// process.
func execAndWaitForStderr(wg *sync.WaitGroup, stderrMatch *regexp.Regexp, timeout time.Duration, name string, args ...string) (context.CancelFunc, <-chan error, error) {
	ctx, cancel := context.WithCancel(context.Background())
	errChan := make(chan error, 1)

	// Set up the command and stderr pipe.
	command := exec.CommandContext(ctx, name, args...) // #nosec
	stderr, err := command.StderrPipe()
	if err != nil {
		return cancel, errChan, errors.Wrap(err, "creating stderr pipe")
	}

	// Start the command in the background. This will only return error if the
	// command is not executable.
	err = command.Start()
	if err != nil {
		return cancel, errChan, fmt.Errorf("starting command: %v", err)
	}

	// Wait until we see the desired output or time out.
	foundLine := make(chan bool, 1)
	scanner := bufio.NewScanner(stderr)
	var log string

	// This goroutine watches the process stderr and sends true along the
	// foundLine channel if a line matches.
	go func() {
		for scanner.Scan() {
			text := scanner.Text()
			log += text + "\n"
			if stderrMatch.MatchString(text) {
				foundLine <- true
				break
			}
		}
	}()

	// Block on either finding a matching line or timeout.
	select {
	case <-foundLine:
		// If we find the line, continue.
	case <-time.After(timeout):
		// If it's a timeout we cancel the command ourselves.
		cancel()
		// We still need to wait for the command to finish.
		command.Wait()                                                  // nolint: errcheck
		return cancel, errChan, fmt.Errorf("timeout, logs:\n%s\n", log) // nolint: staticcheck, revive
	}

	// Increment the wait group so callers can wait for the command to finish.
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := command.Wait()
		errChan <- err
	}()

	return cancel, errChan, nil
}
