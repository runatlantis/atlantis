package tfconfig

import (
	"io/ioutil"
	"os"
)

// FS represents a minimal filesystem implementation
// See io/fs.FS in http://golang.org/s/draft-iofs-design
type FS interface {
	Open(name string) (File, error)
	ReadFile(name string) ([]byte, error)
	ReadDir(dirname string) ([]os.FileInfo, error)
}

// File represents an open file in FS
// See io/fs.File in http://golang.org/s/draft-iofs-design
type File interface {
	Stat() (os.FileInfo, error)
	Read([]byte) (int, error)
	Close() error
}

type osFs struct{}

func (fs *osFs) Open(name string) (File, error) {
	return os.Open(name)
}

func (fs *osFs) ReadFile(name string) ([]byte, error) {
	return ioutil.ReadFile(name)
}

func (fs *osFs) ReadDir(dirname string) ([]os.FileInfo, error) {
	return ioutil.ReadDir(dirname)
}

// NewOsFs provides a basic implementation of FS for an OS filesystem
func NewOsFs() FS {
	return &osFs{}
}
