package activities

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/models"
	"go.temporal.io/sdk/activity"
)

type CloneActivityRequest struct {
	Repo   models.Repo
	Branch string
	Dir    string
	Revision string
}

type CloneActivityResponse struct {
	Dir string
}

func Clone(ctx context.Context, request CloneActivityRequest) (CloneActivityResponse, error) {
	log := activity.GetLogger(ctx)

	cloneDir := request.Dir
	headRepo := request.Repo

	err := os.RemoveAll(cloneDir)
	if err != nil {
		return CloneActivityResponse{}, errors.Wrapf(err, "deleting dir %q before cloning", cloneDir)
	}

	// Create the directory and parents if necessary.
	log.Info("creating dir %q", cloneDir)
	if err := os.MkdirAll(cloneDir, 0700); err != nil {
		return CloneActivityResponse{}, errors.Wrap(err, "creating new workspace")
	}

	headCloneURL := headRepo.CloneURL

	var cmds = [][]string{
		{
			"git", "clone", "--branch", request.Branch, "--single-branch", headCloneURL, cloneDir,
		},
		{
			"git", "checkout", request.Revision,
		},
	}

	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...) // nolint: gosec
		cmd.Dir = cloneDir
		// The git merge command requires these env vars are set.
		cmd.Env = append(os.Environ(), []string{
			"EMAIL=atlantis@runatlantis.io",
			"GIT_AUTHOR_NAME=atlantis",
			"GIT_COMMITTER_NAME=atlantis",
		}...)

		cmdStr := sanitizeGitCredentials(strings.Join(cmd.Args, " "), headRepo)
		output, err := cmd.CombinedOutput()
		sanitizedOutput := sanitizeGitCredentials(string(output), headRepo)
		if err != nil {
			sanitizedErrMsg := sanitizeGitCredentials(err.Error(), headRepo)
			return CloneActivityResponse{}, fmt.Errorf("running %s: %s: %s", cmdStr, sanitizedOutput, sanitizedErrMsg)
		}
		log.Debug("ran: %s. Output: %s", cmdStr, strings.TrimSuffix(sanitizedOutput, "\n"))
	}
	return CloneActivityResponse{Dir: cloneDir}, nil
}

// sanitizeGitCredentials replaces any git clone urls that contain credentials
// in s with the sanitized versions.
func sanitizeGitCredentials(s string, head models.Repo) string {
	return strings.Replace(s, head.CloneURL, head.SanitizedCloneURL, -1)
}
