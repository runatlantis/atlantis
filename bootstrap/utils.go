package bootstrap

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
)

var hashicorpReleasesURL = "https://releases.hashicorp.com"
var terraformVersion = "0.10.8"
var ngrokDownloadURL = "https://bin.equinox.io/c/4VmDzA7iaHb"
var ngrokAPIURL = "http://localhost:4040"

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
	response, err := http.Get(fmt.Sprintf("%s/api/tunnels", ngrokAPIURL))
	if err != nil {
		return "", err
	}
	defer response.Body.Close() // nolint: errcheck

	type tunnel struct {
		Name      string `json:"name"`
		URI       string `json:"uri"`
		PublicURL string `json:"public_url"`
		Proto     string `json:"http"`
	}

	type tunnels struct {
		Tunnels []tunnel
	}

	var t tunnels

	err = json.NewDecoder(response.Body).Decode(&t)
	if err != nil {
		return "", err
	}

	if len(t.Tunnels) != 2 {
		return "", fmt.Errorf("didn't find tunnels that were expected to be created")
	}

	return t.Tunnels[1].PublicURL, nil
}

// nolint: unparam
func downloadAndUnzip(url string, path string, target string) error {
	if err := downloadFile(url, path); err != nil {
		return err
	}
	return unzip(path, target)
}

func executeCmd(cmd string, args []string) (*exec.Cmd, error) {
	command := exec.Command(cmd, args...) // #nosec
	err := command.Start()
	if err != nil {
		return nil, err
	}
	return command, nil
}
