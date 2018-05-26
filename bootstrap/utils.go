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
//
package bootstrap

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh/terminal"
)

const hashicorpReleasesURL = "https://releases.hashicorp.com"
const terraformVersion = "0.10.8"
const ngrokDownloadURL = "https://bin.equinox.io/c/4VmDzA7iaHb"
const ngrokAPIURL = "localhost:41414" // We hope this isn't used.
const atlantisPort = 4141

func readPassword() (string, error) {
	password, err := terminal.ReadPassword(syscall.Stdin)
	return string(password), err
}

func downloadFile(url string, path string) error {
	output, err := os.Create(path)
	if err != nil {
		return err
	}
	defer output.Close() // nolint: errcheck

	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close() // nolint: errcheck

	_, err = io.Copy(output, response.Body)
	return err
}

func unzip(archive, target string) error {
	reader, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}

	for _, file := range reader.File {
		path := filepath.Join(target, file.Name)
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

		if _, err := io.Copy(targetFile, fileReader); err != nil {
			return err
		}
	}

	return nil
}

func getTunnelAddr() (string, error) {
	tunAPI := fmt.Sprintf("http://%s/api/tunnels", ngrokAPIURL)
	response, err := http.Get(tunAPI)
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

	if err = json.NewDecoder(response.Body).Decode(&t); err != nil {
		return "", errors.Wrapf(err, "parsing ngrok api at %s", tunAPI)
	}

	// Find the tunnel we just created.
	expAtlantisURL := fmt.Sprintf("localhost:%d", atlantisPort)
	for _, tun := range t.Tunnels {
		if tun.Proto == "https" && tun.Config.Addr == expAtlantisURL {
			return tun.PublicURL, nil
		}
	}

	return "", fmt.Errorf("did not find ngrok tunnel with proto 'https' and config.addr '%s' in list of tunnels at %s", expAtlantisURL, tunAPI)
}

// nolint: unparam
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

// executeBackgroundCmd executes a command in the background. The function returns a context so
// that the caller may cancel the command prematurely if necessary, as well as an errors channel.
func executeBackgroundCmd(wg *sync.WaitGroup, cmd string, args ...string) (context.CancelFunc, <-chan error) {
	ctx, cancel := context.WithCancel(context.Background())
	command := exec.CommandContext(ctx, cmd, args...) // #nosec

	errChan := make(chan error, 1)
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := command.Run()
		errChan <- err
	}()

	return cancel, errChan
}
