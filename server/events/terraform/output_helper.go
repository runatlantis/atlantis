package terraform

import (
	"bufio"
	"fmt"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/logging"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	outputTimeFmt = "20060102150405"
)

// TfOutputFile hold terraform output file data
type TfOutputFile struct {
	CreatedAt     time.Time
	FullRepoName  string
	PullRequestNr int
	HeadCommit    string
	Project       string
	TfCommand     string
	Workspace     string
}

// OutputHelper define the operations related to the terraform output files
type OutputHelper interface {
	// List list all files from terraform output dir
	List() ([]TfOutputFile, error)
	// ParseFileName converts the terraform output file name to TfOutputFile as it contains some info from the tf command.
	ParseFileName(fileName string) (TfOutputFile, error)
	// CreateFileName creates the file name for a terraform output.
	CreateFileName(fullRepoName string, pullRequestNr int, headCommit string, project string, workspace string, tfCommand string) string
	// ContinueReadFile continue reads the tf output file
	ContinueReadFile(Log *logging.SimpleLogger, fileName string, fileLines chan<- string, done chan bool) error
	// FindOutputFile finds the output file
	FindOutputFile(createdAt, fullRepoName, pullNr, headCommit, project, workspace, tfCommand string) (string, error)
}

type FileOutputHelper struct {
	outputCmdDir string
}

// NewOutputHelper creates a terraform output helper
func NewOutputHelper(outputCmdDir string) (*FileOutputHelper, error) {
	fileInfo, err := os.Stat(outputCmdDir)
	if err != nil || !fileInfo.IsDir() {
		return nil, errors.Wrapf(err, "can't stat %q dir or it's not a directory", outputCmdDir)
	}
	return &FileOutputHelper{outputCmdDir: outputCmdDir}, nil
}

func (f *FileOutputHelper) List() ([]TfOutputFile, error) {
	var tfOutputs []TfOutputFile
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
		return []TfOutputFile{}, errors.Wrapf(err, "can't walk through %q path", f.outputCmdDir)
	}

	return tfOutputs, nil
}

// ConvertTfOutput converts the terraform output file name to the struct.
func (f *FileOutputHelper) ParseFileName(fileName string) (TfOutputFile, error) {
	parts := strings.Split(fileName, "-")

	createdAt, err := time.Parse(outputTimeFmt, parts[0])
	if err != nil {
		return TfOutputFile{}, errors.Wrapf(err, "can't parse date on %q", fileName)
	}

	// Put back "/"
	fullRepoName := strings.ReplaceAll(parts[1], "__", "-")
	fullRepoName = strings.ReplaceAll(fullRepoName, "_", "/")

	pullRequestNr, err := strconv.Atoi(parts[2])
	if err != nil {
		return TfOutputFile{}, errors.Wrap(err, "can't convert pull request number to string")
	}

	return TfOutputFile{
		CreatedAt:     createdAt,
		FullRepoName:  fullRepoName,
		PullRequestNr: pullRequestNr,
		HeadCommit:    parts[3],
		Project:       parts[4],
		Workspace:     parts[5],
		TfCommand:     parts[6],
	}, nil
}

func (f *FileOutputHelper) CreateFileName(fullRepoName string, pullRequestNr int, headCommit string, project string, workspace string, tfCommand string) string {
	// Use UTC to avoid any time zone
	now := time.Now().UTC()

	// Format full repo name to be able to parse it back to its original value
	fullRepoName = strings.ReplaceAll(fullRepoName, "-", "__")
	fullRepoName = strings.ReplaceAll(fullRepoName, "/", "_")

	// Short the head commit
	headCommit = headCommit[:7]

	// OutputHelper file has to be unique per repo, pull request, commit and command
	return fmt.Sprintf("%s-%s-%d-%s-%s-%s-%s",
		now.Format(outputTimeFmt),
		fullRepoName,
		pullRequestNr,
		headCommit,
		project,
		workspace,
		tfCommand,
	)
}

func (f *FileOutputHelper) ContinueReadFile(log *logging.SimpleLogger, fileName string, fileLInes chan<- string, done chan bool) error {
	tfOutputFileName := filepath.Join(f.outputCmdDir, fileName)
	file, _ := os.Open(tfOutputFileName)
	reader := bufio.NewReader(file)

	log.Debug("starting to tail file %q", tfOutputFileName)

	for {
		select {
		case <-done:
			close(fileLInes)
			log.Debug("stopping tailing file %q", tfOutputFileName)
			return nil
		default:
			// Read the line until the break line
			line, err := reader.ReadString('\n')

			if err != nil {
				// If it has reached the end of file, wait a bit to continue reading.
				if err == io.EOF {
					time.Sleep(time.Second)
					continue
				}

				return errors.Wrapf(err, "error reading line from file %q", tfOutputFileName)
			}

			// Post the new line into the string channel
			fileLInes <- line
		}
	}
}

func (f *FileOutputHelper) FindOutputFile(createdAt, fullRepoName, pullNr, headCommit, project, workspace, tfCommand string) (string, error) {
	// Format the repo name
	fullRepoName = strings.ReplaceAll(fullRepoName, "-", "__")
	fullRepoName = strings.ReplaceAll(fullRepoName, "/", "_")

	// Format the file name
	fileName := fmt.Sprintf("%s-%s-%s-%s-%s-%s-%s",
		createdAt,
		fullRepoName,
		pullNr,
		headCommit,
		project,
		workspace,
		tfCommand,
	)

	// Verify if the file exists
	stat, err := os.Stat(filepath.Join(f.outputCmdDir, fileName))
	if err != nil || stat.IsDir() {
		return "", errors.Wrapf(err, "can't stat the file %q or it's a dir", fileName)
	}

	return fileName, nil
}
