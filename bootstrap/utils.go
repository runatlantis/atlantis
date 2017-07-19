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
var terraformVersion = "0.9.11"
var ngrokDownloadURL = "https://bin.equinox.io/c/4VmDzA7iaHb"
var ngrokApiURL = "http://localhost:4040"

func readPassword() (string, error) {
	password, err := terminal.ReadPassword(int(syscall.Stdin))
	return string(password), err
}

func downloadFile(url string, path string) error {
	output, err := os.Create(path)
	if err != nil {
		return err
	}
	defer output.Close()

	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if _, err := io.Copy(output, response.Body); err != nil {
		return err
	}

	return nil
}

func unzip(archive, target string) error {
	reader, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}

	for _, file := range reader.File {
		path := filepath.Join(target, file.Name)
		if file.FileInfo().IsDir() {
			os.MkdirAll(path, file.Mode())
			continue
		}

		fileReader, err := file.Open()
		if err != nil {
			return err
		}
		defer fileReader.Close()

		targetFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}
		defer targetFile.Close()

		if _, err := io.Copy(targetFile, fileReader); err != nil {
			return err
		}
	}

	return nil
}

func getTunnelAddr() (string, error) {
	response, err := http.Get(fmt.Sprintf("%s/api/tunnels", ngrokApiURL))
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

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

func downloadAndUnzip(url string, path string, target string) error {
	if err := downloadFile(url, path); err != nil {
		return err
	}
	return unzip(path, target)
}

func executeCmd(cmd string, args []string) (*exec.Cmd, error) {
	command := exec.Command(cmd, args...)
	err := command.Start()
	if err != nil {
		return nil, err
	}
	return command, nil
}
