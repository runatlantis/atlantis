package models

import (
	"os"
	"path/filepath"
)

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_filepath.go FilePath

type FilePath interface {
	NotExists() bool
	Join(elem ...string) FilePath
	Symlink(newname string) (FilePath, error)
	Resolve() string
}

type LocalFilePath string

func (fp LocalFilePath) NotExists() bool {
	_, err := os.Stat(string(fp))

	return os.IsNotExist(err)
}

func (fp LocalFilePath) Join(elem ...string) FilePath {
	pathComponents := []string{}

	pathComponents = append(pathComponents, string(fp))
	pathComponents = append(pathComponents, elem...)

	return LocalFilePath(filepath.Join(pathComponents...))
}

func (fp LocalFilePath) Symlink(newname string) (FilePath, error) {
	return LocalFilePath(newname), os.Symlink(fp.Resolve(), newname)
}

func (fp LocalFilePath) Resolve() string {
	return string(fp)
}
