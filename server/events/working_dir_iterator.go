package events

import (
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/logging"
)

type WorkDirIterator interface {
	ListCurrentWorkingDirPulls() ([]models.PullRequest, error)
}

type FileWorkDirIterator struct {
	Log          logging.SimpleLogging
	DataDir      string
	GithubClient vcs.GithubPullRequestGetter
	EventParser  EventParsing
}

func NewFileWorkDirIterator(
	githubClient vcs.GithubPullRequestGetter,
	eventParser EventParsing,
	dataDir string,
	log logging.SimpleLogging,
) *FileWorkDirIterator {
	return &FileWorkDirIterator{
		Log:          log,
		DataDir:      dataDir,
		EventParser:  eventParser,
		GithubClient: githubClient,
	}
}

func (f *FileWorkDirIterator) ListCurrentWorkingDirPulls() ([]models.PullRequest, error) {
	var results []models.PullRequest

	baseFilePath := filepath.Join(f.DataDir, workingDirPrefix)

	if _, err := os.Stat(baseFilePath); os.IsNotExist(err) {
		f.Log.Warn("cannot list working dirs, %s doesn't exist", baseFilePath)
		return results, nil
	}

	err := filepath.WalkDir(baseFilePath, func(path string, d fs.DirEntry, err error) error {

		if err != nil {
			return err
		}

		relativePath, err := filepath.Rel(baseFilePath, path)

		if err != nil {
			return errors.Wrap(err, "finding relative path")
		}

		pathComponents := strings.Split(relativePath, string(os.PathSeparator))

		if len(pathComponents) < 3 {
			return nil
		}

		ownerName := pathComponents[0]
		repoName := pathComponents[1]
		pullNum, err := strconv.Atoi(pathComponents[2])

		if err != nil {
			return errors.Wrapf(err, "parsing pull num %s", pathComponents[2])
		}

		f.Log.Debug("Fetching pull for %s/%s #%d", ownerName, repoName, pullNum)

		pull, err := f.GithubClient.GetPullRequestFromName(repoName, ownerName, pullNum)

		if err != nil {
			// let's just continue if we can't find the pull, this is rare and has happened in situations
			// where the repository is renamed
			notFoundErr, ok := err.(*vcs.PullRequestNotFound)

			if !ok {
				return errors.Wrapf(err, "fetching pull for %s", filepath.Join(pathComponents...))
			}

			f.Log.Warn("%s/%s/#%d not found, %s", ownerName, repoName, pullNum, notFoundErr)

			return fs.SkipDir
		}

		internalPull, _, _, err := f.EventParser.ParseGithubPull(pull)

		if err != nil {
			return errors.Wrap(err, "parsing pull request")
		}

		results = append(results, internalPull)

		// if we've made it here we don't want to traverse further in the file tree
		return fs.SkipDir
	})

	if err != nil {
		return results, errors.Wrap(err, "listing current working dir prs")
	}

	return results, nil
}
