package terraform

import (
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// TfOutput hold terraform output file data
type TfOutput struct {
	CreatedAt time.Time
	TfCommand string
	Path      string
}

// Output define the operations related to the terraform output files
type Output interface {
	// List list all files from terraform output dir
	List() ([]TfOutput, error)
	// ParseFileName converts the terraform output file name to TfOutput as it contains some info from the tf command.
	ParseFileName(fileName string) (TfOutput, error)
}

type FileOutput struct {
	outputCmdDir string
}

// NewOutput creates a terraform output
func NewOutput(outputCmdDir string) (Output, error) {
	fileInfo, err := os.Stat(outputCmdDir)
	if err != nil || !fileInfo.IsDir() {
		return nil, errors.Wrapf(err, "can't stat %q dir or it's not a directory", outputCmdDir)
	}
	return &FileOutput{outputCmdDir: outputCmdDir}, nil
}

func (f *FileOutput) List() ([]TfOutput, error) {
	var tfOutputs []TfOutput
	err := filepath.Walk(f.outputCmdDir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			tfOutput, err := f.ParseFileName(info.Name())
			if err != nil {
				return errors.Wrapf(err, "can't convert tf output for file %q", info.Name())
			}

			tfOutputs = append(tfOutputs, tfOutput)
		}

		return nil
	})
	if err != nil {
		return []TfOutput{}, errors.Wrapf(err, "can't walk through %q path", f.outputCmdDir)
	}

	return tfOutputs, nil
}

// ConvertTfOutput converts the terraform output file name to the struct.
func (f *FileOutput) ParseFileName(fileName string) (TfOutput, error) {
	parts := strings.Split(fileName, "-")

	createdAt, err := time.Parse("20060102150405", parts[0])
	if err != nil {
		return TfOutput{}, errors.Wrapf(err, "can't parse date on %q", fileName)
	}

	// The path is always the last part in the file name
	path := strings.ReplaceAll(parts[len(parts)-1], "_", "/")

	return TfOutput{
		CreatedAt: createdAt,
		TfCommand: parts[1],
		Path:      path,
	}, nil
}
